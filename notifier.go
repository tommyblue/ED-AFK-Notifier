package notifier

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tommyblue/ED-AFK-Notifier/bots"
)

type Notifier struct {
	bot                 bots.Bot
	journalFile         string
	journalChanged      chan struct{}
	totalPiratesReward  int
	killedPirates       int
	activeMissions      int
	loggedMissions      map[int]bool
	totalMissionsReward int
	cfg                 *config
	gui                 *GUI

	propagateConfCh chan struct{} // signal received by the GUI when the conf is changed
	syncConfCh      chan struct{} // signal used by the app to know that new configs must be read (called after propagate). Propagates the signal to confObservers
	confObservers   map[string]chan struct{}
	stopCh          chan struct{}
}

func (e *Notifier) newConfObserver() (string, chan struct{}) {
	id := uuid.New().String()
	ch := make(chan struct{})
	e.confObservers[id] = ch

	return id, ch
}

func (e *Notifier) delConfObserver(id string) {
	delete(e.confObservers, id)
}

// New initializes the notifier.
func New(stopCh chan struct{}) (*Notifier, error) {
	e := &Notifier{
		cfg:             &config{},
		propagateConfCh: make(chan struct{}),
		syncConfCh:      make(chan struct{}),
		confObservers:   make(map[string]chan struct{}),
		stopCh:          stopCh,
	}
	e.gui = newGUI(e.propagateConfCh)

	e.syncConfig()

	if e.cfg.logDebug {
		log.SetLevel(log.DebugLevel)
	}

	go func() {
		for {
			select {
			case <-e.stopCh:
				log.Debugln("closing signal propagator")

				return
			case <-e.syncConfCh:
				log.Debugln("multiplexing conf sync signal")
				for _, o := range e.confObservers {
					o <- struct{}{}
				}
			}
		}
	}()

	go func() {
		id, ch := e.newConfObserver()
		for {
			select {
			case <-e.stopCh:
				log.Debugln("closing log level observer")
				e.delConfObserver(id)

				return
			case <-ch:
				log.Infoln("Log level debug:", e.cfg.logDebug)
				if e.cfg.logDebug {
					log.SetLevel(log.DebugLevel)
				} else {
					log.SetLevel(log.InfoLevel)
				}
			}
		}
	}()

	if err := e.initBot(); err != nil {
		return nil, fmt.Errorf("cannot setup the Telegram bot: %v", err)
	}

	if err := e.initJournal(); err != nil {
		return nil, fmt.Errorf("cannot setup the journal file: %v", err)
	}

	return e, nil
}

// Start the Notifier engine, thus reading the Journal and sending notifications through the bot
func (e *Notifier) Start() {
	if e.bot != nil {
		e.bot.Start()
	}

	e.parseJournal()
	e.gui.run()
}

func (e *Notifier) initBot() error {
	if e.cfg.botToken == "" || e.cfg.botChannelId == 0 {
		log.Warningln("Bot not configured, skipping...")

		return nil
	}

	var err error
	e.bot, err = bots.NewTelegram(e.cfg.botToken, e.cfg.botChannelId)

	return err
}

func (e *Notifier) initNotifier() {
	e.totalPiratesReward = 0
	e.killedPirates = 0
	e.activeMissions = 0
	e.loggedMissions = make(map[int]bool)
	e.totalMissionsReward = 0

	file, err := os.Open(e.journalFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lastMissionsTs time.Time
	for scanner.Scan() {
		var j struct {
			Timestamp time.Time `json:"timestamp"`
			Active    []struct {
				MissionID int `json:"MissionID"`
				Expires   int `json:"Expires"`
			} `json:"Active"` // contains active missions, logged at login
			Event       string `json:"event"`
			MissionID   int    `json:"MissionID"`
			Reward      int    `json:"Reward"`
			TotalReward int    `json:"TotalReward"`
		}

		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &j); err != nil {
			log.Infof("Cannot unmarshal %s", line)

			continue
		}

		switch j.Event {
		case "Bounty":
			e.totalPiratesReward += j.TotalReward
			e.killedPirates++
		case "Missions":
			lastMissionsTs = j.Timestamp
			e.activeMissions = 0
			for _, m := range j.Active {
				if m.Expires != 0 {
					e.activeMissions++
				}
			}

		// The following actions must be accepted only if their timestamp is newer the last
		// "Missions" event or the missions count will be wrong.
		case "MissionAccepted":
			if j.Timestamp.After(lastMissionsTs) {
				continue
			}
			e.activeMissions++
		case "MissionRedirected":
			if j.Timestamp.After(lastMissionsTs) {
				continue
			}
			if e.loggedMissions[j.MissionID] {
				continue
			}

			e.activeMissions--
			e.loggedMissions[j.MissionID] = true
		case "MissionCompleted":
			if j.Timestamp.After(lastMissionsTs) {
				continue
			}
			if e.loggedMissions[j.MissionID] {
				continue
			}

			e.activeMissions--
			delete(e.loggedMissions, j.MissionID)

			e.totalMissionsReward += j.Reward
		case "MissionAbandoned":
			if j.Timestamp.After(lastMissionsTs) {
				continue
			}
			e.activeMissions--
			delete(e.loggedMissions, j.MissionID)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

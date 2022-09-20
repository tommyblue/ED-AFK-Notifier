package notifier

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hpcloud/tail"
	log "github.com/sirupsen/logrus"
	"github.com/tommyblue/ED-AFK-Notifier/bots"
)

type Notifier struct {
	bot                 bots.Bot
	journalFile         string
	journalChanged      chan struct{}
	cfg                 *Cfg
	totalPiratesReward  int
	killedPirates       int
	activeMissions      int
	loggedMissions      map[int]bool
	totalMissionsReward int
}

type Cfg struct {
	Token             string
	ChannelId         int64
	JournalPath       string
	FighterNotifs     bool
	ShieldsNotifs     bool // notify about shields state
	KillsNotifs       bool // notify about killed pirates
	KillsSilentNotifs bool // reduce number of notifications for killed pirates, sending a notification every 10 kills
}

// New returns a Notifier with provided configuration
func New(cfg *Cfg) (*Notifier, error) {
	bot, err := bots.NewTelegram(cfg.Token, cfg.ChannelId)
	if err != nil {
		return nil, fmt.Errorf("cannot setup the Telegram bot: %v", err)
	}

	j, err := journalFile(cfg.JournalPath)
	if err != nil {
		return nil, err
	}
	log.Infoln("Found most recent journal file:", j)

	e := &Notifier{
		bot:            bot,
		journalFile:    filepath.Join(cfg.JournalPath, j),
		journalChanged: make(chan struct{}),
		cfg:            cfg,
	}

	e.initNotifier()

	e.watchJournal()

	return e, nil
}

func (e *Notifier) watchJournal() {
	go func() {
		for range time.Tick(30 * time.Second) {

			oldJournal := e.journalFile
			j, err := journalFile(e.cfg.JournalPath)
			if err != nil {
				continue
			}

			if oldJournal != filepath.Join(e.cfg.JournalPath, j) {
				log.Infoln("Found new journal file:", j)
				e.journalFile = filepath.Join(e.cfg.JournalPath, j)

				e.initNotifier()

				e.journalChanged <- struct{}{}
			}
		}
	}()
}

type eventType string

var (
	bountyEventType            eventType = "Bounty"
	missionsEventType          eventType = "Missions"
	missionAcceptedEventType   eventType = "MissionAccepted"
	missionRedirectedEventType eventType = "MissionRedirected"
	missionCompletedEventType  eventType = "MissionCompleted"
	missionAbandonedEventType  eventType = "MissionAbandoned"
	hullDamageEventType        eventType = "HullDamage"
	diedEventType              eventType = "Died"
	shieldStateEventType       eventType = "ShieldState"
)

func (e *Notifier) initCounters() {
	e.totalPiratesReward = 0
	e.killedPirates = 0
	e.activeMissions = 0
	e.loggedMissions = make(map[int]bool)
	e.totalMissionsReward = 0
}

func (e *Notifier) initNotifier() {
	e.initCounters()

	file, err := os.Open(e.journalFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lastMissionsTs time.Time
	for scanner.Scan() {
		var j journalEvent

		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &j); err != nil {
			log.Infof("Cannot unmarshal %s", line)

			continue
		}

		switch j.Event {

		case bountyEventType:
			e.totalPiratesReward += j.TotalPiratesReward
			e.killedPirates++

			log.Debugf("Total reward: %d\n", e.totalPiratesReward)
			log.Debugf("Killed pirates: %d\n", e.killedPirates)

		case missionsEventType:
			lastMissionsTs = j.Timestamp
			e.activeMissions = 0
			for _, m := range j.Active {
				if m.Expires != 0 {
					e.activeMissions++
				}
			}

			log.Debugf("Active missions: %d\n", e.activeMissions)

		// The following actions must be accepted only if their timestamp is newer the last
		// "Missions" event or the missions count will be wrong.
		case missionAcceptedEventType:
			if j.Timestamp.After(lastMissionsTs) {
				continue
			}
			e.activeMissions++
			log.Debugf("Active missions: %d\n", e.activeMissions)

		case missionRedirectedEventType:
			if j.Timestamp.After(lastMissionsTs) {
				continue
			}
			if e.loggedMissions[j.MissionID] {
				continue
			}

			e.activeMissions--
			e.loggedMissions[j.MissionID] = true

			log.Debugf("Active missions: %d\n", e.activeMissions)

		case missionCompletedEventType:
			if j.Timestamp.After(lastMissionsTs) {
				continue
			}
			if e.loggedMissions[j.MissionID] {
				continue
			}

			e.activeMissions--
			delete(e.loggedMissions, j.MissionID)

			e.totalMissionsReward += j.MissionReward

			log.Debugf("Active missions: %d\n", e.activeMissions)
			log.Debugf("Total missions reward: %d\n", e.totalMissionsReward)

		case missionAbandonedEventType:
			if j.Timestamp.After(lastMissionsTs) {
				continue
			}
			e.activeMissions--
			delete(e.loggedMissions, j.MissionID)

			log.Debugf("Active missions: %d\n", e.activeMissions)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// Start the Notifier engine, thus reading the Journal and sending notifications through the bot
func (e *Notifier) Start() {
	e.bot.Start()

	for {
		log.Infoln("Reading journal...")
		t, err := tail.TailFile(e.journalFile, tail.Config{Follow: true, Poll: true})
		if err != nil {
			log.Fatalf("cannot tail the log file: %v\n", err)
		}

		go func() {
			<-e.journalChanged
			log.Infoln("Journal changed, reloading...")
			t.Stop()
		}()

		startTime := time.Now()

		events := map[eventType]eventFn{
			hullDamageEventType:        hullDamageEvent,
			diedEventType:              diedEvent,
			shieldStateEventType:       shieldStateEvent,
			bountyEventType:            bountyEvent,
			missionAcceptedEventType:   missionAcceptedEvent,
			missionCompletedEventType:  missionCompletedEvent,
			missionRedirectedEventType: missionRedirectedEvent,
			missionAbandonedEventType:  missionAbandonedEvent,
			missionsEventType:          missionsInitEvent,
		}

		for line := range t.Lines {
			var j journalEvent
			if err := json.Unmarshal([]byte(line.Text), &j); err != nil {
				log.Infof("Cannot unmarshal %s", line.Text)
			}

			// Skip logs already in the journal befor this app has started
			var skipNotify bool
			if j.Timestamp.Before(startTime) {
				skipNotify = true
			}

			// log.Debugln(line.Text)

			if fn, ok := events[j.Event]; ok {
				if err := fn(e, j, skipNotify); err != nil {
					log.Infoln("[ERROR]", err)
				}
			}
		}
	}
}

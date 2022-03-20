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
	totalPiratesReward  int
	killedPirates       int
	activeMissions      int
	loggedMissions      map[int]bool
	totalMissionsReward int
	gui                 *GUI
}

// New returns a Notifier with provided configuration
func New() (*Notifier, error) {
	e := &Notifier{}
	e.gui = newGUI()

	if e.configBool(CONFIG_LOG_DEBUG) {
		log.SetLevel(log.DebugLevel)
	}

	if err := e.initBot(); err != nil {
		return nil, fmt.Errorf("cannot setup the Telegram bot: %v", err)
	}

	if err := e.initJournal(); err != nil {
		return nil, fmt.Errorf("cannot setup the journal file: %v", err)
	}

	return e, nil
}

func (e *Notifier) configString(key string) string {
	return e.gui.App.Preferences().String(key)
}

func (e *Notifier) configBool(key string) bool {
	return e.gui.App.Preferences().Bool(key)
}

func (e *Notifier) configInt64(key string) int64 {
	return int64(e.gui.App.Preferences().Int(key))
}

func (e *Notifier) initJournal() error {
	jFile := e.configString(CONFIG_JOURNAL_PATH)

	if jFile == "" {
		log.Warningln("Journal file not configured, skipping...")
	}

	j, err := journalFile(jFile)
	if err != nil {
		log.Warnf("journal file error: %v", err)
		return nil
	}
	log.Infoln("Found most recent journal file:", j)

	e.journalFile = filepath.Join(jFile, j)
	e.journalChanged = make(chan struct{})

	e.initNotifier()

	e.watchJournal()

	return nil
}

func (e *Notifier) initBot() error {
	token := e.configString(CONFIG_BOT_TOKEN)
	channelId := e.configInt64(CONFIG_BOT_CHANNEL_ID)

	if token == "" || channelId == 0 {
		log.Warningln("Bot not configured, skipping...")

		return nil
	}

	var err error
	e.bot, err = bots.NewTelegram(token, channelId)

	return err

}

func (e *Notifier) watchJournal() {
	go func() {
		for range time.Tick(30 * time.Second) {

			oldJournal := e.journalFile
			j, err := journalFile(e.configString(CONFIG_JOURNAL_PATH))
			if err != nil {
				continue
			}

			if oldJournal != filepath.Join(e.configString(CONFIG_JOURNAL_PATH), j) {
				log.Infoln("Found new journal file:", j)
				e.journalFile = filepath.Join(e.configString(CONFIG_JOURNAL_PATH), j)

				e.initNotifier()

				e.journalChanged <- struct{}{}
			}
		}
	}()
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

func (e *Notifier) LogConfig() {
	log.Infof("Config:")
	log.Infof("  Notify fighter status: %t", e.configBool(CONFIG_NOTIFY_FIGHTER))
	log.Infof("  Notify shields status: %t", e.configBool(CONFIG_NOTIFY_SHIELDS))
	log.Infof("  Notify on kills: %t (silent: %t)", e.configBool(CONFIG_NOTIFY_KILLS), e.configBool(CONFIG_NOTIFY_FIGHTER))
	log.Infof("  Journal file path: %s", e.configString(CONFIG_JOURNAL_PATH))
}

// Start the Notifier engine, thus reading the Journal and sending notifications through the bot
func (e *Notifier) Start() {
	e.gui.run()

	if e.bot != nil {
		e.bot.Start()
	}

	if e.journalFile != "" {

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

			events := map[string]eventFn{
				"HullDamage":        hullDamageEvent,
				"Died":              diedEvent,
				"ShieldState":       shieldStateEvent,
				"Bounty":            bountyEvent,
				"MissionAccepted":   missionAcceptedEvent,
				"MissionCompleted":  missionCompletedEvent,
				"MissionRedirected": missionRedirectedEvent,
				"MissionAbandoned":  missionAbandonedEvent,
				"Missions":          missionsInitEvent,
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

				log.Debugln(line.Text)

				if fn, ok := events[j.Event]; ok {
					if err := fn(e, j, skipNotify); err != nil {
						log.Infoln("[ERROR]", err)
					}
				}
			}
		}
	}
}

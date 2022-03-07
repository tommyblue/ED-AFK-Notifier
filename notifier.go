package notifier

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hpcloud/tail"
	"github.com/spf13/viper"
	"github.com/tommyblue/ED-AFK-Notifier/bots"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Notifier struct {
	bot                 bots.Bot
	journalFile         string
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
	log.Println("Found most recent journal file:", j)

	e := &Notifier{
		bot:            bot,
		journalFile:    filepath.Join(cfg.JournalPath, j),
		cfg:            cfg,
		loggedMissions: make(map[int]bool),
	}

	e.initMissions()

	e.watchJournal()

	return e, nil
}

func (e *Notifier) watchJournal() {
	go func() {
		for range time.Tick(3 * time.Minute) {

			oldJournal := e.journalFile
			j, err := journalFile(e.cfg.JournalPath)
			if err != nil {
				continue
			}

			if oldJournal != filepath.Join(e.cfg.JournalPath, j) {
				log.Println("Found new journal file:", j)
				e.journalFile = filepath.Join(e.cfg.JournalPath, j)

				e.initMissions()
			}
		}
	}()
}

func (e *Notifier) initMissions() {
	file, err := os.Open(e.journalFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var j struct {
			Active []struct {
				MissionID int `json:"MissionID"`
			} `json:"Active"` // contains active missions, logged at login
			Event       string `json:"event"`
			MissionID   int    `json:"MissionID"`
			Reward      int    `json:"Reward"`
			TotalReward int    `json:"TotalReward"`
		}
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &j); err != nil {
			log.Printf("Cannot unmarshal %s", line)
		}

		switch j.Event {
		case "Missions":
			e.activeMissions = len(j.Active)

			for _, m := range j.Active {
				e.loggedMissions[m.MissionID] = false
			}

		case "MissionAccepted":
			e.activeMissions++
		case "MissionRedirected":
			if e.loggedMissions[j.MissionID] {
				continue
			}

			e.activeMissions--
			e.loggedMissions[j.MissionID] = true
		case "MissionCompleted":
			if e.loggedMissions[j.MissionID] {
				continue
			}

			e.activeMissions--
			delete(e.loggedMissions, j.MissionID)

			e.totalMissionsReward += j.Reward
		case "MissionAbandoned":
			e.activeMissions--
			delete(e.loggedMissions, j.MissionID)
		case "Bounty":
			e.totalPiratesReward += j.TotalReward
			e.killedPirates++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	log.Println("Active missions:", e.activeMissions)
	p := message.NewPrinter(language.Make("en"))
	if e.totalMissionsReward > 0 {
		log.Println("Obtained reward for missions until now:", p.Sprintf("%f", e.totalMissionsReward))
	}

	if e.totalPiratesReward > 0 {
		log.Printf("Obtained reward for killed pirates (%d) until now: %s", e.killedPirates, p.Sprintf("%d", e.totalPiratesReward))
	}
}

// Start the Notifier engine, thus reading the Journal and sending notifications through the bot
func (e *Notifier) Start() error {
	e.bot.Start()

	t, err := tail.TailFile(e.journalFile, tail.Config{Follow: true, Poll: true})
	if err != nil {
		return fmt.Errorf("cannot tail the log file: %v", err)
	}

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
	}

	for line := range t.Lines {
		var j journalEvent
		if err := json.Unmarshal([]byte(line.Text), &j); err != nil {
			log.Printf("Cannot unmarshal %s", line.Text)
		}

		// Skip logs already in the journal befor this app has started
		if j.Timestamp.Before(startTime) {
			continue
		}

		if viper.GetBool("journal.debug") {
			log.Println(line.Text)
		}

		if fn, ok := events[j.Event]; ok {
			if err := fn(e, j); err != nil {
				log.Println(err)
			}
		}
	}

	return nil
}

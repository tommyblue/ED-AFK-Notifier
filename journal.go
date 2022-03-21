package notifier

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hpcloud/tail"
	log "github.com/sirupsen/logrus"
)

type journalEvent struct {
	Timestamp          time.Time `json:"timestamp"`
	Event              string    `json:"event"`
	Health             float64   `json:"Health"`
	PlayerPilot        bool      `json:"PlayerPilot"`
	Fighter            bool      `json:"Fighter"`
	ShieldsUp          bool      `json:"ShieldsUp"`   // whether shields are up or down
	TotalPiratesReward int       `json:"TotalReward"` // total credits earned by killing pirates
	MissionID          int       `json:"MissionID"`
	MissionReward      int       `json:"Reward"` // credits earned by completing a mission
}

func (e *Notifier) initJournal() error {
	jPath := e.cfg.journalPath

	if jPath == "" {
		log.Warningln("Journal path not configured, skipping...")
	}

	j, err := getJournalFile(jPath)
	if err != nil {
		log.Warnf("journal file error: %v", err)
		return nil
	}
	log.Infoln("Found most recent journal file:", j)

	e.journalFile = filepath.Join(jPath, j)
	e.journalChanged = make(chan struct{})

	e.initNotifier()

	e.watchJournal()

	return nil
}

func (e *Notifier) watchJournal() {
	checkNewJournal := func() {
		log.Debugln("checking for new journal file")
		oldJournal := e.journalFile
		j, err := getJournalFile(e.cfg.journalPath)
		if err != nil {
			log.Infof("getJournalFile error: %w", err)

			return
		}

		log.Debugln("comparing", oldJournal, "and", filepath.Join(e.cfg.journalPath, j))
		if oldJournal != filepath.Join(e.cfg.journalPath, j) {
			log.Infoln("found new journal file:", j)
			e.journalFile = filepath.Join(e.cfg.journalPath, j)

			e.initNotifier()

			e.journalChanged <- struct{}{}
		}
	}

	go func() {
		id, ch := e.newConfObserver()

		for {
			select {
			case <-e.stopCh:
				log.Debugln("closing journal watch")
				e.delConfObserver(id)

				return
			case <-time.Tick(30 * time.Second):
				checkNewJournal()
			case <-ch:
				log.Debugln("here")
				checkNewJournal()
			}
		}
	}()
}

func (e *Notifier) parseJournal() {
	if e.journalFile != "" {
		go func() {
			for {
				log.Debugln("Reading journal...")
				t, err := tail.TailFile(e.journalFile, tail.Config{Follow: true, Poll: true})
				if err != nil {
					log.Fatalf("cannot tail the log file: %v\n", err)
				}

				go func() {
					for {
						select {
						case <-e.stopCh:
							log.Debugln("closing journal parser")

							return
						case <-e.journalChanged:
							log.Infoln("Journal changed, reloading...")
							t.Stop()
						}
					}
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
						log.Warningf("Cannot unmarshal %s", line.Text)
					}

					// Skip logs already in the journal befor this app has started
					var skipNotify bool
					if j.Timestamp.Before(startTime) {
						skipNotify = true
					}

					log.Debugln(line.Text)

					if fn, ok := events[j.Event]; ok {
						if err := fn(e, j, skipNotify); err != nil {
							log.Errorln(err)
						}
					}
				}
			}
		}()
	}
}

func getJournalFile(logPath string) (string, error) {
	files, err := os.ReadDir(logPath)
	if err != nil {
		return "", err
	}

	var lastFile fs.DirEntry
	var lastMod time.Time

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasPrefix(file.Name(), "Journal") || !strings.HasSuffix(file.Name(), ".log") {
			continue
		}

		finfo, err := file.Info()
		if err != nil {
			log.Debugf("cannot get info for %s", file.Name())

			continue
		}

		if lastMod.IsZero() || finfo.ModTime().After(lastMod) {
			lastMod = finfo.ModTime()
			lastFile = file
		}
	}

	if lastMod.IsZero() || lastFile.Name() == "" {
		return "", fmt.Errorf("cannot find the journal file")
	}

	return lastFile.Name(), nil
}

func (j *journalEvent) printLog(v ...interface{}) {
	if j.Timestamp.Add(10 * time.Second).Before(time.Now()) {
		return
	}

	log.Debugln(v...)
}

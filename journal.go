package notifier

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func journalPath(logPath string) (string, error) {
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
			log.Printf("cannot get info for %s", file.Name())
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

	log.Println("Found journal file:", lastFile.Name())

	return filepath.Join(logPath, lastFile.Name()), nil
}

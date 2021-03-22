package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Journal struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
	Health    float64   `json:"Health"`
}

func JournalFilePath(logPath string) string {
	files, err := os.ReadDir(logPath)
	if err != nil {
		log.Fatal(err)
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
			log.Printf("Cannot get info for %s", file.Name())
			continue
		}

		if lastMod.IsZero() || finfo.ModTime().After(lastMod) {
			lastMod = finfo.ModTime()
			lastFile = file
		}
	}

	return filepath.Join(logPath, lastFile.Name())
}

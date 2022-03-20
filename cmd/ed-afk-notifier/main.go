package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	notifier "github.com/tommyblue/ED-AFK-Notifier"
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	notifier, err := notifier.New()
	if err != nil {
		log.Fatalf("Cannot initialize the notifier: %v", err)
	}

	notifier.LogConfig()

	notifier.Start()
}

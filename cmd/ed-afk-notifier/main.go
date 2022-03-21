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
	stopCh := make(chan struct{})
	notifier, err := notifier.New(stopCh)
	if err != nil {
		log.Fatalf("Cannot initialize the notifier: %v", err)
	}

	notifier.LogConfig()

	notifier.Start()
	close(stopCh)
}

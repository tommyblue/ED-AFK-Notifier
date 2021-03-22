package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hpcloud/tail"
	"github.com/spf13/viper"
)

func main() {
	if err := setupConfig(); err != nil {
		log.Fatal(err)
	}

	bot, err := NewBot(viper.GetString("telegram.token"), viper.GetInt64("telegram.channelId"))
	if err != nil {
		log.Fatalf("Cannot setup the Telegram bot: %v", err)
	}

	bot.Start()

	fpath := JournalFilePath(viper.GetString("journal.path"))
	log.Printf("Found the journal path %s", fpath)

	t, err := tail.TailFile(fpath, tail.Config{Follow: true, Poll: true})
	if err != nil {
		log.Fatalf("Cannot tail the log file: %v", err)
	}

	startTime := time.Now()

	for line := range t.Lines {
		var j Journal
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

		if j.Event == "HullDamage" {
			if err := bot.Send(fmt.Sprintf("Hull damage detected, integrity is %2f", j.Health)); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}
}

func setupConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	return viper.ReadInConfig()
}

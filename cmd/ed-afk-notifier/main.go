package main

import (
	"log"

	"github.com/spf13/viper"
	notifier "github.com/tommyblue/ED-AFK-Notifier"
)

func main() {
	if err := setupConfig(); err != nil {
		log.Fatalf("Cannot read config: %v", err)
	}

	cfg := &notifier.Cfg{
		Token:         viper.GetString("telegram.token"),
		ChannelId:     viper.GetInt64("telegram.channelId"),
		JournalPath:   viper.GetString("journal.path"),
		FighterNotifs: viper.GetBool("journal.fighter"),
	}

	notifier, err := notifier.New(cfg)

	if err != nil {
		log.Fatalf("Cannot initialize the notifier: %v", err)
	}

	if err := notifier.Start(); err != nil {
		log.Fatal(err)
	}
}

func setupConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	return viper.ReadInConfig()
}

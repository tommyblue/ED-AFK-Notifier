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
		Token:             viper.GetString("telegram.token"),
		ChannelId:         viper.GetInt64("telegram.channelId"),
		JournalPath:       viper.GetString("journal.path"),
		FighterNotifs:     viper.GetBool("journal.fighter"),
		ShieldsNotifs:     viper.GetBool("journal.shields"),
		KillsNotifs:       viper.GetBool("journal.kills"),
		KillsSilentNotifs: viper.GetBool("journal.silent_kills"),
	}

	logConfig(cfg)

	notifier, err := notifier.New(cfg)

	if err != nil {
		log.Fatalf("Cannot initialize the notifier: %v", err)
	}

	notifier.Start()
}

func setupConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	return viper.ReadInConfig()
}

func logConfig(cfg *notifier.Cfg) {
	log.Printf("Config:")
	log.Printf("  Notify fighter status: %t", cfg.FighterNotifs)
	log.Printf("  Notify shields status: %t", cfg.ShieldsNotifs)
	log.Printf("  Notify on kills: %t (silent: %t)", cfg.KillsNotifs, cfg.KillsSilentNotifs)
	log.Printf("  Journal file path: %s", cfg.JournalPath)
}

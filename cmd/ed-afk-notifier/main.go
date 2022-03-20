package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	notifier "github.com/tommyblue/ED-AFK-Notifier"
	"github.com/tommyblue/ED-AFK-Notifier/gui"
	"github.com/tommyblue/ED-AFK-Notifier/gui/types"
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

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

	if viper.GetBool("journal.debug") {
		log.SetLevel(log.DebugLevel)
	}

	logConfig(cfg)

	notifier, err := notifier.New(cfg)

	if err != nil {
		log.Fatalf("Cannot initialize the notifier: %v", err)
	}

	go notifier.Start()

	gui.Run(&types.Config{
		AppName: "ED AFK Notifier",
		Debug:   viper.GetBool("journal.debug"),
	})
	log.Println("2")
}

func setupConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	return viper.ReadInConfig()
}

func logConfig(cfg *notifier.Cfg) {
	log.Infof("Config:")
	log.Infof("  Notify fighter status: %t", cfg.FighterNotifs)
	log.Infof("  Notify shields status: %t", cfg.ShieldsNotifs)
	log.Infof("  Notify on kills: %t (silent: %t)", cfg.KillsNotifs, cfg.KillsSilentNotifs)
	log.Infof("  Journal file path: %s", cfg.JournalPath)
}

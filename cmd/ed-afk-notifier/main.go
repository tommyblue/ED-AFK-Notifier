package main

import (
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	notifier "github.com/tommyblue/ED-AFK-Notifier"
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
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
	gui()
}

func gui() {
	a := app.NewWithID("io.github.tommyblue.ed-afk-notifier.preferences")
	w := a.NewWindow("ED AFK Notifier")
	w.SetMaster()
	w.SetContent(widget.NewLabel("Hello"))
	w.SetContent(widget.NewButton("Open config", func() {
		w3 := a.NewWindow("Config")
		var debug bool

		timeoutSelector := widget.NewRadioGroup([]string{"On", "Off"}, func(selected string) {
			switch selected {
			case "On":
				debug = true
			case "Off":
				debug = false
			}

			a.Preferences().SetString("EnableDebug", selected)
			viper.Set("journal.debug", debug)
		})

		timeoutSelector.SetSelected(a.Preferences().StringWithFallback("EnableDebug", "Off"))
		w3.SetContent(timeoutSelector)
		w3.Show()
	}))
	w.Show()
	a.Run()
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

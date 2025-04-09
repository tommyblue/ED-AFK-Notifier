package main

import (
	"os"

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
	log.Infof("Starting ED-AFK-Notifier v%s", notifier.Version)

	if err := setupConfig(); err != nil {
		log.Fatalf("Cannot read config: %v", err)
	}

	// Get notification service from config, default to telegram for backward compatibility
	service := viper.GetString("notification.service")
	if service == "" {
		service = "telegram"
	}

	cfg := &notifier.Cfg{
		NotificationService: service,
		JournalPath:         viper.GetString("journal.path"),
		FighterNotifs:       viper.GetBool("journal.fighter"),
		ShieldsNotifs:       viper.GetBool("journal.shields"),
		KillsNotifs:         viper.GetBool("journal.kills"),
		KillsSilentNotifs:   viper.GetBool("journal.silent_kills"),
	}

	// Set service-specific configuration
	switch service {
	case "telegram":
		cfg.TelegramToken = viper.GetString("telegram.token")
		cfg.TelegramChannelId = viper.GetInt64("telegram.channelId")
	case "gotify":
		cfg.GotifyURL = viper.GetString("gotify.url")
		cfg.GotifyToken = viper.GetString("gotify.token")
		cfg.GotifyTitle = viper.GetString("gotify.title")
		cfg.GotifyPriority = viper.GetInt("gotify.priority")
	default:
		log.Fatalf("Unknown notification service: %s", service)
	}

	if viper.GetBool("journal.debug") {
		log.SetLevel(log.DebugLevel)
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
	log.Infof("Config:")
	log.Infof("  Notification service: %s", cfg.NotificationService)
	log.Infof("  Notify fighter status: %t", cfg.FighterNotifs)
	log.Infof("  Notify shields status: %t", cfg.ShieldsNotifs)
	log.Infof("  Notify on kills: %t (silent: %t)", cfg.KillsNotifs, cfg.KillsSilentNotifs)
	log.Infof("  Journal file path: %s", cfg.JournalPath)

	switch cfg.NotificationService {
	case "telegram":
		if cfg.TelegramToken == "" {
			log.Warn("  Telegram token not set")
		}
		log.Infof("  Telegram channel ID: %d", cfg.TelegramChannelId)
	case "gotify":
		log.Infof("  Gotify URL: %s", cfg.GotifyURL)
		if cfg.GotifyToken == "" {
			log.Warn("  Gotify token not set")
		}
		if cfg.GotifyTitle != "" {
			log.Infof("  Gotify notification title: %s", cfg.GotifyTitle)
		}
		log.Infof("  Gotify notification priority: %d", cfg.GotifyPriority)
	}
}

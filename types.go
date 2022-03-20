package notifier

import "fyne.io/fyne/v2"

type GUI struct {
	App        fyne.App
	MainWindow fyne.Window
}

const (
	CONFIG_JOURNAL_PATH        = "journal.path"
	CONFIG_BOT_TOKEN           = "bot.token"
	CONFIG_BOT_CHANNEL_ID      = "bot.channelId"
	CONFIG_LOG_DEBUG           = "log.debug"
	CONFIG_NOTIFY_SHIELDS      = "notify.shields"
	CONFIG_NOTIFY_FIGHTER      = "notify.fighter"
	CONFIG_NOTIFY_KILLS        = "notify.kills"
	CONFIG_NOTIFY_SILENT_KILLS = "notify.silent_kills"
)

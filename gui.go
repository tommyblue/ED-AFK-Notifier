package notifier

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	log "github.com/sirupsen/logrus"
)

type GUI struct {
	App        fyne.App
	MainWindow fyne.Window

	propagateCfgCh chan struct{}
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

func newGUI(propagateCh chan struct{}) *GUI {
	gui := &GUI{
		propagateCfgCh: propagateCh,
	}
	gui.App = app.NewWithID("io.github.tommyblue.ed-afk-notifier.preferences")

	gui.MainWindow = gui.App.NewWindow("ED AFK Notifier")
	gui.MainWindow.SetMaster()
	gui.MainWindow.SetContent(widget.NewLabel("Hello"))

	gui.MainWindow.SetContent(widget.NewButton("Open config", gui.configPage()))

	gui.MainWindow.SetCloseIntercept(func() {
		log.Debugln("Closing main window")
		gui.MainWindow.Close()
	})

	return gui
}

func (g *GUI) run() {
	g.MainWindow.ShowAndRun()
}

func (g *GUI) configPage() func() {
	return func() {
		w := g.App.NewWindow("Config")

		propagateConf := func() {
			g.propagateCfgCh <- struct{}{}
		}

		form := &widget.Form{
			Items:      []*widget.FormItem{},
			SubmitText: "Close",
			OnSubmit: func() {
				log.Debugln("Submitted conf")
				propagateConf()
				w.Close()
			},
		}

		form.Append("Journal path", FolderSelector(g, CONFIG_JOURNAL_PATH, g.App.Preferences().StringWithFallback(CONFIG_JOURNAL_PATH, "")))
		form.Append("Telegram bot token", TextField(g, CONFIG_BOT_TOKEN, "Insert your bot token here"))
		form.Append("Telegram bot channel ID", TextField(g, CONFIG_BOT_CHANNEL_ID, "Insert your bot channel ID here"))
		form.Append("Enable Debug", BoolSelector(g, CONFIG_LOG_DEBUG, nil))
		form.Append("Shields status", BoolSelector(g, CONFIG_NOTIFY_SHIELDS, nil))
		form.Append("Fighter status", BoolSelector(g, CONFIG_NOTIFY_FIGHTER, nil))

		silentKills := BoolSelector(g, CONFIG_NOTIFY_SILENT_KILLS, nil)

		kills := binding.NewBool()
		callback := binding.NewDataListener(func() {
			v, _ := kills.Get()
			if v {
				silentKills.(*widget.RadioGroup).Enable()
			} else {
				silentKills.(*widget.RadioGroup).Disable()
			}
		})
		kills.AddListener(callback)

		form.Append("Count kills", BoolSelector(g, CONFIG_NOTIFY_KILLS, kills))
		form.Append("Reduce kills noise", silentKills)

		w.SetContent(form)

		w.SetCloseIntercept(func() {
			log.Debugln("Closing conf window")
			propagateConf()
			w.Close()
		})

		w.Show()
	}
}

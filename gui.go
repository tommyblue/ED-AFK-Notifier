package notifier

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	log "github.com/sirupsen/logrus"
)

// New GUI instance
func newGUI() *GUI {
	gui := &GUI{}
	gui.App = app.NewWithID("io.github.tommyblue.ed-afk-notifier.preferences")

	gui.MainWindow = gui.App.NewWindow("ED AFK Notifier")
	gui.MainWindow.SetMaster()
	gui.MainWindow.SetContent(widget.NewLabel("Hello"))

	gui.MainWindow.SetContent(widget.NewButton("Open config", gui.configPage()))

	return gui
}

func (g *GUI) run() {
	g.MainWindow.ShowAndRun()
	// TODO: send signal to close the whole app
}

type preference struct {
	key   string
	vtype string
}

const (
	TYPE_BOOL = "bool"
	TYPE_STR  = "string"
)

var preferences = []preference{
	{key: CONFIG_JOURNAL_PATH, vtype: TYPE_STR},
	{key: CONFIG_BOT_TOKEN, vtype: TYPE_STR},
	{key: CONFIG_BOT_CHANNEL_ID, vtype: TYPE_STR},
	{key: CONFIG_LOG_DEBUG, vtype: TYPE_BOOL},
	{key: CONFIG_NOTIFY_SHIELDS, vtype: TYPE_BOOL},
	{key: CONFIG_NOTIFY_FIGHTER, vtype: TYPE_BOOL},
	{key: CONFIG_NOTIFY_KILLS, vtype: TYPE_BOOL},
	{key: CONFIG_NOTIFY_SILENT_KILLS, vtype: TYPE_BOOL},
}

func (g *GUI) configPage() func() {
	return func() {
		w := g.App.NewWindow("Config")

		undo := func() {
			// TODO: get the previous values
			// for _, pref := range preferences {
			// 	switch pref.vtype {
			// 	case TYPE_BOOL:
			// 		// v := viper.GetBool(pref.key)
			// 		log.Debugln("Restoring preference:", pref, "=>", v)
			// 		g.App.Preferences().SetBool(pref.key, v)
			// 	case TYPE_STR:
			// 		v := viper.GetString(pref.key)
			// 		log.Debugln("Restoring preference:", pref, "=>", v)
			// 		g.App.Preferences().SetString(pref.key, v)
			// 	}
			// }
		}

		form := &widget.Form{
			Items:      []*widget.FormItem{},
			SubmitText: "Save",
			CancelText: "Undo",
			OnCancel: func() {
				undo()
				w.Close()
			},
			OnSubmit: func() {
				// TODO: decide how to manage persistency. It should be done on save, not
				// when changing the values in the form (or maybe yes?)
				for _, pref := range preferences {
					switch pref.vtype {
					case TYPE_BOOL:
						v := g.App.Preferences().Bool(pref.key)
						log.Debugln("Saving preference:", pref, "=>", v)
						// viper.Set(pref.key, v)
					case TYPE_STR:
						v := g.App.Preferences().String(pref.key)
						log.Debugln("Saving preference:", pref, "=>", v)
						// viper.Set(pref.key, v)
					}
				}
				w.Close()
			},
		}

		form.Append("Journal path", FolderSelector(g, CONFIG_JOURNAL_PATH))
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
			log.Debugln("Closing")
			undo()
			w.Close()
		})
		w.Show()
	}
}

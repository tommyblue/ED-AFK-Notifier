package pages

import (
	log "github.com/sirupsen/logrus"

	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/viper"
	"github.com/tommyblue/ED-AFK-Notifier/gui/internal/components"
	"github.com/tommyblue/ED-AFK-Notifier/gui/types"
)

type preference struct {
	key   string
	vtype string
}

const (
	TYPE_BOOL = "bool"
	TYPE_STR  = "string"
)

var preferences = []preference{
	{key: types.CONFIG_JOURNAL_PATH, vtype: TYPE_STR},
	{key: types.CONFIG_BOT_TOKEN, vtype: TYPE_STR},
	{key: types.CONFIG_BOT_CHANNEL_ID, vtype: TYPE_STR},
	{key: types.CONFIG_LOG_DEBUG, vtype: TYPE_BOOL},
	{key: types.CONFIG_NOTIFY_SHIELDS, vtype: TYPE_BOOL},
	{key: types.CONFIG_NOTIFY_FIGHTER, vtype: TYPE_BOOL},
	{key: types.CONFIG_NOTIFY_KILLS, vtype: TYPE_BOOL},
	{key: types.CONFIG_NOTIFY_SILENT_KILLS, vtype: TYPE_BOOL},
}

func Config(g *types.GUI) func() {
	return func() {
		w := g.App.NewWindow("Config")

		undo := func() {
			for _, pref := range preferences {
				switch pref.vtype {
				case TYPE_BOOL:
					v := viper.GetBool(pref.key)
					log.Debugln("Restoring preference:", pref, "=>", v)
					g.App.Preferences().SetBool(pref.key, v)
				case TYPE_STR:
					v := viper.GetString(pref.key)
					log.Debugln("Restoring preference:", pref, "=>", v)
					g.App.Preferences().SetString(pref.key, v)
				}
			}
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
				for _, pref := range preferences {
					switch pref.vtype {
					case TYPE_BOOL:
						v := g.App.Preferences().Bool(pref.key)
						log.Debugln("Saving preference:", pref, "=>", v)
						viper.Set(pref.key, v)
					case TYPE_STR:
						v := g.App.Preferences().String(pref.key)
						log.Debugln("Saving preference:", pref, "=>", v)
						viper.Set(pref.key, v)
					}
				}
				w.Close()
			},
		}

		form.Append("Journal path", components.FolderSelector(g, types.CONFIG_JOURNAL_PATH))
		form.Append("Telegram bot token", components.TextField(g, types.CONFIG_BOT_TOKEN, "Insert your bot token here"))
		form.Append("Telegram bot channel ID", components.TextField(g, types.CONFIG_BOT_CHANNEL_ID, "Insert your bot channel ID here"))
		form.Append("Enable Debug", components.BoolSelector(g, types.CONFIG_LOG_DEBUG, nil))
		form.Append("Shields status", components.BoolSelector(g, types.CONFIG_NOTIFY_SHIELDS, nil))
		form.Append("Fighter status", components.BoolSelector(g, types.CONFIG_NOTIFY_FIGHTER, nil))

		silentKills := components.BoolSelector(g, types.CONFIG_NOTIFY_SILENT_KILLS, nil)

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

		form.Append("Count kills", components.BoolSelector(g, types.CONFIG_NOTIFY_KILLS, kills))
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

package pages

import (
	log "github.com/sirupsen/logrus"

	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/viper"
	"github.com/tommyblue/ED-AFK-Notifier/gui/internal/components"
	"github.com/tommyblue/ED-AFK-Notifier/gui/types"
)

type getPref func() bool

var preferences = []string{
	types.CONFIG_LOG_DEBUG,
	types.CONFIG_NOTIFY_SHIELDS,
	types.CONFIG_NOTIFY_FIGHTER,
	types.CONFIG_NOTIFY_KILLS,
	types.CONFIG_NOTIFY_SILENT_KILLS,
}

func Config(g *types.GUI) func() {
	return func() {
		w := g.App.NewWindow("Config")

		undo := func() {
			for _, pref := range preferences {
				v := viper.GetBool(pref)
				log.Debugln("Restoring preference:", pref, "=>", v)
				g.App.Preferences().SetBool(pref, v)
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
				log.Debugln("Config form submitted")
				for _, pref := range preferences {
					v := g.App.Preferences().Bool(pref)
					log.Debugln("Saving preference:", pref, "=>", v)
					viper.Set(pref, v)
				}
				w.Close()
			},
		}

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

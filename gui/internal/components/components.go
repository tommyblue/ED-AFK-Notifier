package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	log "github.com/sirupsen/logrus"
	"github.com/tommyblue/ED-AFK-Notifier/gui/types"
)

// BoolSelector creates a On/Off selector. The label is used to store the value in the config env.
func BoolSelector(g *types.GUI, label string, v binding.Bool) fyne.CanvasObject {
	if v == nil {
		v = binding.NewBool()
	}

	selector := widget.NewRadioGroup([]string{"On", "Off"}, func(selected string) {
		switch selected {
		case "On":
			v.Set(true)
		case "Off":
			v.Set(false)
		}

		log.Debugln(label, ":", v)

		value, _ := v.Get()
		g.App.Preferences().SetBool(label, value)
	})

	selector.Horizontal = true

	s := "Off"
	if g.App.Preferences().BoolWithFallback(label, false) {
		s = "On"
	}
	selector.SetSelected(s)

	return selector
}

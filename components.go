package notifier

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	log "github.com/sirupsen/logrus"
)

// BoolSelector creates a On/Off selector. The label is used to store the value in the config env.
func BoolSelector(g *GUI, label string, v binding.Bool) fyne.CanvasObject {
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

func FolderSelector(g *GUI, label, location string) fyne.CanvasObject {
	w := g.App.NewWindow(fmt.Sprintf("folder_selector.%s", label))

	path := g.App.Preferences().StringWithFallback(label, "")
	v := binding.NewString()
	v.Set(path)

	l := widget.NewLabel(path)

	callback := binding.NewDataListener(func() {
		v, _ := v.Get()
		l.SetText(v)
	})
	v.AddListener(callback)

	cb := func(uri fyne.ListableURI, err error) {
		if uri == nil {
			return
		}

		log.Debugln("Selected Path:", uri.Path())
		g.App.Preferences().SetString(label, uri.Path())
		v.Set(uri.Path())
		w.Hide()
	}
	var dlg *dialog.FileDialog
	btn := widget.NewButton("Select folder", func() {
		dlg = dialog.NewFolderOpen(cb, w)
		dlg.SetOnClosed(func() {
			w.Hide()
		})
		// TODO: close window on cancel
		if location != "" {
			lister, err := storage.ListerForURI(storage.NewFileURI(location))
			if err == nil {
				dlg.SetLocation(lister)
			}
		}

		dlg.Resize(fyne.NewSize(800, 600))
		dlg.Show()
		w.Resize(fyne.NewSize(800, 600))
		w.Show()
	})

	content := container.New(layout.NewHBoxLayout(), l, layout.NewSpacer(), btn)

	return content
}

func TextField(g *GUI, label, placeholder string) fyne.CanvasObject {
	input := widget.NewEntry()
	input.SetPlaceHolder(placeholder)

	v := binding.NewString()
	v.Set(g.App.Preferences().StringWithFallback(label, ""))
	callback := binding.NewDataListener(func() {
		v, _ := v.Get()
		log.Debugln("new", label, ":", v)
		input.SetText(v)
		g.App.Preferences().SetString(label, v)
	})
	v.AddListener(callback)

	input.Bind(v)

	return input
}

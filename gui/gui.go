package gui

import (
	"log"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/tommyblue/ED-AFK-Notifier/gui/internal/pages"
	"github.com/tommyblue/ED-AFK-Notifier/gui/types"
)

// New GUI instance
func Run(c *types.Config) {
	g := &types.GUI{}
	g.App = app.NewWithID("io.github.tommyblue.ed-afk-notifier.preferences")

	g.MainWindow = g.App.NewWindow(c.AppName)
	g.MainWindow.SetMaster()
	g.MainWindow.SetContent(widget.NewLabel("Hello"))

	g.MainWindow.SetContent(widget.NewButton("Open config", pages.Config(g)))

	g.MainWindow.ShowAndRun()

	// TODO: send signal to close the whole app
	log.Println("1")
}

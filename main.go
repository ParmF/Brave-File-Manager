package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/yourusername/brave-file-manager/internal/ui"
)

func main() {
	application := app.New()
	window := application.NewWindow("Brave File Manager")

	manager, err := ui.NewApp(window)
	if err != nil {
		panic(err)
	}
	manager.RunOnClose()

	window.ShowAndRun()
}

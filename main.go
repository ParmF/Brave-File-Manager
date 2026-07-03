package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/asnasn/Brave-File-Manager/internal/ui"
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

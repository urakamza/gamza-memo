package main

import (
	"embed"
	_ "embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	ok, cleanup := ensureSingleInstance()
	if !ok {
		return
	}
	defer cleanup()

	noteService := NewNoteService()

	app := application.New(application.Options{
		Name:        "감자 메모",
		Description: "감자 메모",
		Services: []application.Service{
			application.NewService(noteService),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assets),
			Middleware: noteService.ImageMiddleware,
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: true,
			AdditionalBrowserArgs:         []string{"--disable-minimum-font-size"},
		},
	})
	noteService.app = app

	tray := app.SystemTray.New()
	tray.SetLabel("감자 메모")
	tray.SetTooltip("감자 메모")
	tray.OnDoubleClick(func() {
		noteService.OpenMainWindow()
	})

	menu := application.NewMenu()
	menu.Add("새 메모").OnClick(func(ctx *application.Context) {
		noteService.CreateNote()
	})
	menu.Add("메모 목록").OnClick(func(ctx *application.Context) {
		noteService.OpenMainWindow()
	})
	menu.AddSeparator()
	menu.Add("종료").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	tray.SetMenu(menu)

	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}

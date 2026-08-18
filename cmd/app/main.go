package main

import (
	"embed"
	"log"

	"sing-scope/internal/app"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend_dist
var assets embed.FS

func main() {
	appService := app.NewAppService()
	defer appService.Close()

	wailsApp := application.New(application.Options{
		Name:        "sing-scope",
		Description: "Cross-Platform sing-box Traffic Analyzer",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Services: []application.Service{
			application.NewService(appService),
		},
	})

	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "sing-scope | sing-box Traffic Analyzer",
		Width:     1280,
		Height:    800,
		MinWidth:  960,
		MinHeight: 600,
	})

	err := wailsApp.Run()
	if err != nil {
		log.Fatal(err)
	}
}

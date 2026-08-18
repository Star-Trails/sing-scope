package main

import (
	"embed"
	"io/fs"
	"log"

	"sing-scope/internal/app"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend_dist
var rawAssets embed.FS

func main() {
	appService := app.NewAppService()
	defer appService.Close()

	assetsSubFS, err := fs.Sub(rawAssets, "frontend_dist")
	if err != nil {
		log.Fatalf("failed to open frontend assets: %v", err)
	}

	handler := app.NewAssetHandler(appService, assetsSubFS)

	wailsApp := application.New(application.Options{
		Name:        "sing-scope",
		Description: "Cross-Platform sing-box Traffic Analyzer",
		Assets: application.AssetOptions{
			Handler: handler,
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

	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}

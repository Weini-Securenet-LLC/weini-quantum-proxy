package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	assetserver "github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"hermes-agent/proxy-node-studio/internal/wailsapp"
)

//go:embed web/*
var assets embed.FS

func main() {
	relaunched, err := wailsapp.MaybeRelaunchElevated()
	if err != nil {
		log.Fatal(err)
	}
	if relaunched {
		return
	}
	app := wailsapp.New()
	webAssets, err := fs.Sub(assets, "web")
	if err != nil {
		log.Fatal(err)
	}
	err = wails.Run(&options.App{
		Title:            "维尼量子节点",
		Width:            1440,
		Height:           920,
		MinWidth:         1180,
		MinHeight:        760,
		MaxWidth:         1440,
		MaxHeight:        920,
		DisableResize:    false,
		Frameless:        true,
		AssetServer:      &assetserver.Options{Assets: webAssets},
		BackgroundColour: &options.RGBA{R: 6, G: 10, B: 19, A: 1},
		OnStartup:        app.Startup,
		OnBeforeClose:    app.BeforeClose,
		OnShutdown:       app.Shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

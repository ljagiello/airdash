package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/objc"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Load config first
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Getting user home directory", "error", err)
		os.Exit(1)
	}

	cfg, err := LoadConfig(filepath.Join(homeDirectory, ".airdash", "config.yaml"))
	if err != nil {
		logger.Error("Loading config", "error", err)
		os.Exit(1)
	}

	if cfg.Interval == 0 {
		cfg.Interval = 60
	}

	airGradientAPIURL := getAirGradientAPIURL(cfg.LocationID)

	// Create the app manually instead of using RunApp
	app := appkit.Application_SharedApplication()
	app.SetActivationPolicy(appkit.ApplicationActivationPolicyAccessory)

	item := appkit.StatusBar_SystemStatusBar().StatusItemWithLength(-1)
	objc.Retain(&item)
	item.Button().SetTitle("🔄 AirDash")

	updateStatus := func() {
		measures, err := getAirGradientMeasures(airGradientAPIURL, cfg.Token)
		if err != nil {
			logger.Error("Fetching measures", "error", err)
			return
		}
		logger.Debug("AirGradientMeasures", "measures", measures)

		// convert the temperature to the desired unit
		temperature := convertTemperature(measures.Atmp, cfg.TempUnit)

		// updates to the ui should happen on the main thread to avoid segfaults
		dispatch.MainQueue().DispatchAsync(func() {
			item.Button().SetTitle(fmt.Sprintf("🌡️ %.2f  💨 %.0f  💧 %.1f  🫧 %.0f",
				temperature,
				measures.Pm02,
				measures.Rhum,
				measures.Rco2,
			))
		})
	}

	go func() {
		// Fetch data immediately on startup
		updateStatus()

		ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
		defer ticker.Stop()

		// Continue fetching at regular intervals
		for range ticker.C {
			updateStatus()
		}
	}()

	itemQuit := appkit.NewMenuItem()
	itemQuit.SetTitle("Quit")
	itemQuit.SetAction(objc.Sel("terminate:"))

	menu := appkit.NewMenu()
	menu.AddItem(itemQuit)
	item.SetMenu(menu)

	app.Run()
}

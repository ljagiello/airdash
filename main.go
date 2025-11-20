package main

import (
	"fmt"
	"os"
	"time"

	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/objc"
)

func main() {
	// Load config first
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Getting user home directory", "error", err)
		os.Exit(1)
	}

	cfg, err := LoadConfig(homeDirectory + "/.airdash/config.yaml")
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

	go func() {
		var airGradientMeasures AirGradientMeasures
		ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
		defer ticker.Stop()

		// Fetch data immediately on startup
		airGradientMeasures, err = getAirGradientMeasures(airGradientAPIURL, cfg.Token)
		if err != nil {
			logger.Error("Fetching measures", "error", err)
		} else {
			logger.Debug("AirGradientMeasures", "measures", airGradientMeasures)

			// convert the temperature to the desired unit
			temperature := convertTemperature(airGradientMeasures.Atmp, cfg.TempUnit)

			// updates to the ui should happen on the main thread to avoid segfaults
			dispatch.MainQueue().DispatchAsync(func() {
				item.Button().SetTitle(fmt.Sprintf("🌡️ %.2f  💨 %.0f  💧 %.1f  🫧 %.0f",
					temperature,
					airGradientMeasures.Pm02,
					airGradientMeasures.Rhum,
					airGradientMeasures.Rco2,
				))
			})
		}

		// Continue fetching at regular intervals
		for range ticker.C {
			airGradientMeasures, err = getAirGradientMeasures(airGradientAPIURL, cfg.Token)
			if err != nil {
				logger.Error("Fetching measures", "error", err)
				continue
			}
			logger.Debug("AirGradientMeasures", "measures", airGradientMeasures)

			// convert the temperature to the desired unit
			temperature := convertTemperature(airGradientMeasures.Atmp, cfg.TempUnit)

			// updates to the ui should happen on the main thread to avoid segfaults
			dispatch.MainQueue().DispatchAsync(func() {
				item.Button().SetTitle(fmt.Sprintf("🌡️ %.2f  💨 %.0f  💧 %.1f  🫧 %.0f",
					temperature,
					airGradientMeasures.Pm02,
					airGradientMeasures.Rhum,
					airGradientMeasures.Rco2,
				))
			})
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

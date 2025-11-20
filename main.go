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

	var airGradientMeasures AirGradientMeasures

	item := appkit.StatusBar_SystemStatusBar().StatusItemWithLength(-1)
	objc.Retain(&item)
	item.Button().SetTitle("🔄 AirDash")

	go func() {
		for {
			time.Sleep(time.Duration(cfg.Interval) * time.Second)
			airGradientMeasures, err = getAirGradientMeasures(airGradientAPIURL, cfg.Token)
			if err != nil {
				logger.Error("Fetching measures", "error", err)
				continue
			}
			logger.Debug("AirGradientMeasures", "measures", airGradientMeasures)

			// updates to the ui should happen on the main thread to avoid segfaults
			dispatch.MainQueue().DispatchAsync(func() {
				item.Button().SetTitle(fmt.Sprintf("💨 %.0f  🫧 %.0f",
					airGradientMeasures.Pm02,
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

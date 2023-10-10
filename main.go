package main

import (
	"fmt"
	"os"
	"time"

	"github.com/progrium/macdriver/dispatch"
	"github.com/progrium/macdriver/macos"
	"github.com/progrium/macdriver/macos/appkit"
	"github.com/progrium/macdriver/objc"
)

const AIR_GRADIENT_API_URL = "https://api.airgradient.com/public/api/v1/locations/measures/current"

func main() {
	macos.RunApp(launched)
}

func launched(app appkit.Application, delegate *appkit.ApplicationDelegate) {
	delegate.SetApplicationShouldTerminateAfterLastWindowClosed(func(appkit.Application) bool {
		return false
	})

	var airGradientMeasures AirGradientMeasures

	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Getting user home directory", "error", err)
	}

	cfg, err := LoadConfig(homeDirectory + "/.airdash/config.yaml")
	if err != nil {
		logger.Error("Loading config", "error", err)
	}

	if cfg.Interval == 0 {
		cfg.Interval = 60
	}

	item := appkit.StatusBar_SystemStatusBar().StatusItemWithLength(-1)
	objc.Retain(&item)
	item.Button().SetTitle("🔄 AirDash")

	go func() {
		for {
			select {
			case <-time.After(time.Duration(cfg.Interval) * time.Second):
				airGradientMeasures, err = getAirGradientMeasures(AIR_GRADIENT_API_URL, cfg.Token)
				if err != nil {
					logger.Error("Fetching measures", "error", err)
					continue
				}
			}

			if len(airGradientMeasures) == 0 {
				logger.Error("No measurements found")
				return
			}

			logger.Debug("AirGradientMeasures", "measures", airGradientMeasures[0])

			temperature := airGradientMeasures[0].Atmp
			if cfg.TempUnit == "F" {
				temperature = (airGradientMeasures[0].Atmp * 9 / 5) + 32
			}

			// updates to the ui should happen on the main thread to avoid segfaults
			dispatch.MainQueue().DispatchAsync(func() {
				item.Button().SetTitle(fmt.Sprintf("🌡️ %.2f  💨 %.0f  💧 %.1f  🫧 %.0f",
					temperature,
					airGradientMeasures[0].Pm02,
					airGradientMeasures[0].Rhum,
					airGradientMeasures[0].Rco2,
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
}

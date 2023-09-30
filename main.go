package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/progrium/macdriver/dispatch"
	"github.com/progrium/macdriver/macos"
	"github.com/progrium/macdriver/macos/appkit"
	"github.com/progrium/macdriver/objc"
)

type AirGradientMeasures []struct {
	LocationID         int       `json:"locationId"`
	LocationName       string    `json:"locationName"`
	Pm01               any       `json:"pm01"`
	Pm02               int       `json:"pm02"`
	Pm10               any       `json:"pm10"`
	Pm003Count         any       `json:"pm003Count"`
	Atmp               float64   `json:"atmp"`
	Rhum               int       `json:"rhum"`
	Rco2               int       `json:"rco2"`
	Tvoc               float64   `json:"tvoc"`
	Wifi               int       `json:"wifi"`
	Timestamp          time.Time `json:"timestamp"`
	LedMode            string    `json:"ledMode"`
	LedCo2Threshold1   int       `json:"ledCo2Threshold1"`
	LedCo2Threshold2   int       `json:"ledCo2Threshold2"`
	LedCo2ThresholdEnd int       `json:"ledCo2ThresholdEnd"`
	Serialno           string    `json:"serialno"`
	FirmwareVersion    any       `json:"firmwareVersion"`
	TvocIndex          int       `json:"tvocIndex"`
	NoxIndex           int       `json:"noxIndex"`
}

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
				payload, err := fetchMeasures(cfg.Token)
				if err != nil {
					logger.Error("Fetching measures", "error", err)
					return
				}

				err = json.Unmarshal(payload, &airGradientMeasures)
				if err != nil {
					logger.Error("Parsing JSON payload", "error", err)
					return
				}
			}
			logger.Debug("AirGradientMeasures", "measures", airGradientMeasures[0])

			// updates to the ui should happen on the main thread to avoid segfaults
			dispatch.MainQueue().DispatchAsync(func() {
				item.Button().SetTitle(fmt.Sprintf("🌡️%.2fF  💨 %d  💦 %d  🫧 %d",
					(airGradientMeasures[0].Atmp*9/5)+32,
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

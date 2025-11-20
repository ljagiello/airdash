package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var aboutWindow objc.Object

func showAboutWindow() {
	// If window already exists, just bring it to front
	if !aboutWindow.IsNil() {
		window := appkit.WindowFrom(aboutWindow.Ptr())
		window.MakeKeyAndOrderFront(nil)
		appkit.Application_SharedApplication().ActivateIgnoringOtherApps(true)
		return
	}

	// Create window
	rect := foundation.Rect{
		Origin: foundation.Point{X: 0, Y: 0},
		Size:   foundation.Size{Width: 400, Height: 200},
	}

	window := appkit.NewWindowWithContentRectStyleMaskBackingDefer(
		rect,
		appkit.WindowStyleMaskTitled|appkit.WindowStyleMaskClosable,
		appkit.BackingStoreBuffered,
		false,
	)
	window.SetTitle("About AirDash")
	window.Center()

	// Create content view
	contentView := window.ContentView()

	// App name label
	nameLabel := appkit.NewTextField()
	nameLabel.SetStringValue("AirDash")
	nameLabel.SetEditable(false)
	nameLabel.SetBordered(false)
	nameLabel.SetDrawsBackground(false)
	nameLabel.SetFont(appkit.Font_SystemFontOfSize(18))
	nameLabel.SetFrame(foundation.Rect{
		Origin: foundation.Point{X: 20, Y: 140},
		Size:   foundation.Size{Width: 360, Height: 30},
	})
	contentView.AddSubview(nameLabel)

	// Version label
	versionLabel := appkit.NewTextField()
	versionLabel.SetStringValue(fmt.Sprintf("Version: %s", version))
	versionLabel.SetEditable(false)
	versionLabel.SetBordered(false)
	versionLabel.SetDrawsBackground(false)
	versionLabel.SetFont(appkit.Font_SystemFontOfSize(13))
	versionLabel.SetFrame(foundation.Rect{
		Origin: foundation.Point{X: 20, Y: 100},
		Size:   foundation.Size{Width: 360, Height: 20},
	})
	contentView.AddSubview(versionLabel)

	// Commit label
	commitLabel := appkit.NewTextField()
	commitLabel.SetStringValue(fmt.Sprintf("Commit: %s", commit))
	commitLabel.SetEditable(false)
	commitLabel.SetBordered(false)
	commitLabel.SetDrawsBackground(false)
	commitLabel.SetFont(appkit.Font_SystemFontOfSize(11))
	commitLabel.SetFrame(foundation.Rect{
		Origin: foundation.Point{X: 20, Y: 70},
		Size:   foundation.Size{Width: 360, Height: 20},
	})
	contentView.AddSubview(commitLabel)

	// Build date label
	dateLabel := appkit.NewTextField()
	dateLabel.SetStringValue(fmt.Sprintf("Built: %s", date))
	dateLabel.SetEditable(false)
	dateLabel.SetBordered(false)
	dateLabel.SetDrawsBackground(false)
	dateLabel.SetFont(appkit.Font_SystemFontOfSize(11))
	dateLabel.SetFrame(foundation.Rect{
		Origin: foundation.Point{X: 20, Y: 40},
		Size:   foundation.Size{Width: 360, Height: 20},
	})
	contentView.AddSubview(dateLabel)

	aboutWindow = window.Object
	window.MakeKeyAndOrderFront(nil)
	appkit.Application_SharedApplication().ActivateIgnoringOtherApps(true)
}

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

	// Create About menu item with callback
	itemAbout := appkit.NewMenuItem()
	itemAbout.SetTitle("About AirDash")

	// Create custom handler class and register method
	handlerClass := objc.AllocateClass(objc.GetClass("NSObject"), "AboutMenuHandler", 0)
	objc.RegisterClass(handlerClass)
	objc.AddMethod(handlerClass, objc.Sel("showAbout:"), func(self objc.Object, cmd objc.Selector) {
		dispatch.MainQueue().DispatchAsync(func() {
			showAboutWindow()
		})
	})

	// Create instance of our handler class
	handler := objc.Call[objc.Object](handlerClass, objc.Sel("alloc"))
	handler = objc.Call[objc.Object](handler, objc.Sel("init"))
	objc.Retain(&handler)

	itemAbout.SetTarget(handler)
	itemAbout.SetAction(objc.Sel("showAbout:"))

	// Create Quit menu item
	itemQuit := appkit.NewMenuItem()
	itemQuit.SetTitle("Quit")
	itemQuit.SetAction(objc.Sel("terminate:"))

	// Build menu
	menu := appkit.NewMenu()
	menu.AddItem(itemAbout)
	menu.AddItem(appkit.MenuItem_SeparatorItem())
	menu.AddItem(itemQuit)
	item.SetMenu(menu)

	app.Run()
}

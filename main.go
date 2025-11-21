package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/helper/action"
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

//go:embed assets/app/logo.svg
var logoSVG []byte

func showAboutWindow() {
	// If window already exists, just bring it to front
	if !aboutWindow.IsNil() {
		window := appkit.WindowFrom(aboutWindow.Ptr())
		window.MakeKeyAndOrderFront(nil)
		appkit.Application_SharedApplication().ActivateIgnoringOtherApps(true)
		return
	}

	// Create window - more compact
	rect := foundation.Rect{
		Origin: foundation.Point{X: 0, Y: 0},
		Size:   foundation.Size{Width: 400, Height: 340},
	}

	window := appkit.NewWindowWithContentRectStyleMaskBackingDefer(
		rect,
		appkit.WindowStyleMaskTitled|appkit.WindowStyleMaskClosable,
		appkit.BackingStoreBuffered,
		false,
	)
	window.SetTitle("About AirDash")
	window.SetReleasedWhenClosed(false) // Keep window in memory when closed
	window.Center()

	// Create content view
	contentView := window.ContentView()

	// Load SVG logo - smaller size
	logoImage := appkit.NewImageWithData(logoSVG)

	// Create image view for logo
	logoView := appkit.NewImageView()
	logoView.SetImage(logoImage)
	logoView.SetFrame(foundation.Rect{
		Origin: foundation.Point{X: 150, Y: 230},
		Size:   foundation.Size{Width: 100, Height: 100},
	})
	contentView.AddSubview(logoView)

	// App name label - centered, larger, bold
	nameLabel := appkit.NewTextField()
	nameLabel.SetStringValue("AirDash")
	nameLabel.SetEditable(false)
	nameLabel.SetBordered(false)
	nameLabel.SetDrawsBackground(false)
	nameLabel.SetFont(appkit.Font_BoldSystemFontOfSize(28))
	nameLabel.SetAlignment(appkit.TextAlignmentCenter)
	nameLabel.SetFrame(foundation.Rect{
		Origin: foundation.Point{X: 0, Y: 185},
		Size:   foundation.Size{Width: 400, Height: 35},
	})
	contentView.AddSubview(nameLabel)

	// Version label - centered, tighter spacing
	versionLabel := appkit.NewTextField()
	versionLabel.SetStringValue(fmt.Sprintf("Version  %s", version))
	versionLabel.SetEditable(false)
	versionLabel.SetBordered(false)
	versionLabel.SetDrawsBackground(false)
	versionLabel.SetFont(appkit.Font_SystemFontOfSize(13))
	versionLabel.SetAlignment(appkit.TextAlignmentCenter)
	versionLabel.SetFrame(foundation.Rect{
		Origin: foundation.Point{X: 0, Y: 140},
		Size:   foundation.Size{Width: 400, Height: 18},
	})
	contentView.AddSubview(versionLabel)

	// Build label - centered, tighter spacing
	buildLabel := appkit.NewTextField()
	buildLabel.SetStringValue(fmt.Sprintf("Build  %s", date))
	buildLabel.SetEditable(false)
	buildLabel.SetBordered(false)
	buildLabel.SetDrawsBackground(false)
	buildLabel.SetFont(appkit.Font_SystemFontOfSize(13))
	buildLabel.SetAlignment(appkit.TextAlignmentCenter)
	buildLabel.SetFrame(foundation.Rect{
		Origin: foundation.Point{X: 0, Y: 120},
		Size:   foundation.Size{Width: 400, Height: 18},
	})
	contentView.AddSubview(buildLabel)

	// Commit label - centered, tighter spacing
	commitLabel := appkit.NewTextField()
	commitLabel.SetStringValue(fmt.Sprintf("Commit  %s", commit))
	commitLabel.SetEditable(false)
	commitLabel.SetBordered(false)
	commitLabel.SetDrawsBackground(false)
	commitLabel.SetFont(appkit.Font_SystemFontOfSize(13))
	commitLabel.SetAlignment(appkit.TextAlignmentCenter)
	commitLabel.SetTextColor(appkit.Color_LinkColor())
	commitLabel.SetFrame(foundation.Rect{
		Origin: foundation.Point{X: 0, Y: 100},
		Size:   foundation.Size{Width: 400, Height: 18},
	})
	contentView.AddSubview(commitLabel)

	// GitHub button - centered
	githubButton := appkit.Button_ButtonWithTitleTargetAction("GitHub", nil, objc.Selector{})
	githubButton.SetBezelStyle(appkit.BezelStyleRounded)
	githubButton.SetFrame(foundation.Rect{
		Origin: foundation.Point{X: 150, Y: 40},
		Size:   foundation.Size{Width: 100, Height: 32},
	})

	// Set button action to open GitHub URL
	githubButton.SetTarget(githubButton.Object)
	githubButton.SetAction(objc.Sel("performAction:"))

	// Use action helper to handle click
	action.Set(githubButton, func(sender objc.Object) {
		url := foundation.URL_URLWithString("https://github.com/ljagiello/airdash")
		appkit.Workspace_SharedWorkspace().OpenURL(url)
	})

	contentView.AddSubview(githubButton)

	aboutWindow = window.Object
	objc.Retain(&aboutWindow) // Retain to prevent deallocation
	window.MakeKeyAndOrderFront(nil)
	appkit.Application_SharedApplication().ActivateIgnoringOtherApps(true)
}

func main() {
	// Handle subcommands first (install/uninstall/version)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			if err := installDaemon(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "uninstall":
			if err := uninstallDaemon(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "version", "--version", "-version":
			fmt.Printf("airdash %s (commit: %s, built: %s)\n", version, commit, date)
			return
		}
	}

	// Parse flags
	configPath := flag.String("config", getDefaultConfigPath(), "path to config file")
	flag.Parse()

	// Load config
	cfg, err := LoadConfig(*configPath)
	if err != nil {
		logger.Error("Loading config", "error", err, "path", *configPath)
		os.Exit(1)
	}

	// Set defaults
	if cfg.Interval == 0 {
		cfg.Interval = 60
	}

	// Run GUI
	runGUI(cfg)
}

func runGUI(cfg *Config) {
	airGradientAPIURL := getAirGradientAPIURL(cfg.LocationID)

	// Create the app manually instead of using RunApp
	app := appkit.Application_SharedApplication()
	app.SetActivationPolicy(appkit.ApplicationActivationPolicyAccessory)

	// Schedule UI setup to run on main queue after app.Run() starts
	dispatch.MainQueue().DispatchAsync(func() {
		// Auto-install LaunchAgent silently on first launch
		if !isDaemonInstalled() {
			logger.Info("First launch detected - installing LaunchAgent")
			if err := installDaemon(); err != nil {
				// Log error but continue running in GUI mode
				logger.Error("Failed to install LaunchAgent - running in GUI mode only", "error", err)
			} else {
				logger.Info("LaunchAgent installed successfully - exiting to let launchd start")
				// Success - quit and let launchd start
				app.Terminate(nil)
				return
			}
		}

		item := appkit.StatusBar_SystemStatusBar().StatusItemWithLength(-1)
		objc.Retain(&item)

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
		itemAbout := appkit.NewMenuItemWithAction("About AirDash", "", func(sender objc.Object) {
			showAboutWindow()
		})

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
	})

	app.Run()
}

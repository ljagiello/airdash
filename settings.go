package main

import (
	"github.com/progrium/darwinkit/helper/action"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"
)

const (
	settingsWidth  = 480.0
	settingsHeight = 660.0
	settingsMargin = 20.0
	contentWidth   = settingsWidth - 2*settingsMargin
)

// sampleMeasures feeds the live preview before the first real fetch arrives.
var sampleMeasures = AirGradientMeasures{Atmp: 24.5, Pm02: 4, Rhum: 66, Rco2: 679}

// settingsWindow is the AirDash Settings panel. Every control applies
// immediately to the status bar; there is no save step.
type settingsWindow struct {
	sb     *statusBar
	window appkit.Window

	preview        appkit.Button
	cardButtons    map[IconStyle]appkit.Button
	valueLabels    map[MetricID]appkit.TextField
	metricSwitches map[MetricID]appkit.Switch
	roundSwitch    appkit.Switch
	tabularSwitch  appkit.Switch
	loginSwitch    appkit.Switch
}

func (sb *statusBar) showSettingsWindow() {
	if sb.settingsUI == nil {
		sb.settingsUI = newSettingsWindow(sb)
	}
	sb.settingsUI.refresh()
	sb.settingsUI.window.MakeKeyAndOrderFront(nil)
	appkit.Application_SharedApplication().ActivateIgnoringOtherApps(true)
}

func (sw *settingsWindow) previewMeasures() AirGradientMeasures {
	if sw.sb.measures != nil {
		return *sw.sb.measures
	}
	return sampleMeasures
}

// refresh syncs every control and the preview strip with the current state.
func (sw *settingsWindow) refresh() {
	s := sw.sb.settings
	m := sw.previewMeasures()

	dark := isDarkAppearance(sw.window.ContentView())
	sw.preview.SetAttributedTitle(buildStatusTitle(m, sw.sb.cfg.TempUnit, s, dark))

	for style, card := range sw.cardButtons {
		card.SetState(boolState(style == s.IconStyle))
	}
	for _, info := range allMetrics {
		value := "—"
		if sw.sb.measures != nil {
			value = formatMetric(info.ID, m, sw.sb.cfg.TempUnit, s.RoundValues)
		}
		sw.valueLabels[info.ID].SetStringValue(value)
		sw.metricSwitches[info.ID].SetState(boolState(s.Visible[info.ID]))
	}
	sw.roundSwitch.SetState(boolState(s.RoundValues))
	sw.tabularSwitch.SetState(boolState(s.TabularDigits))
	sw.loginSwitch.SetState(boolState(isDaemonInstalled()))
}

func boolState(on bool) appkit.ControlStateValue {
	if on {
		return appkit.ControlStateValueOn
	}
	return appkit.ControlStateValueOff
}

func newSettingsWindow(sb *statusBar) *settingsWindow {
	sw := &settingsWindow{
		sb:             sb,
		cardButtons:    map[IconStyle]appkit.Button{},
		valueLabels:    map[MetricID]appkit.TextField{},
		metricSwitches: map[MetricID]appkit.Switch{},
	}

	window := appkit.NewWindowWithContentRectStyleMaskBackingDefer(
		foundation.Rect{Size: foundation.Size{Width: settingsWidth, Height: settingsHeight}},
		appkit.WindowStyleMaskTitled|appkit.WindowStyleMaskClosable,
		appkit.BackingStoreBuffered,
		false,
	)
	window.SetTitle("AirDash Settings")
	window.SetReleasedWhenClosed(false)
	window.Center()
	objc.Retain(&window)
	sw.window = window

	content := window.ContentView()

	// Layout cursor: place() converts a top-down flow into AppKit's
	// bottom-left coordinates.
	top := settingsMargin
	place := func(height float64) float64 {
		y := settingsHeight - top - height
		top += height
		return y
	}
	gap := func(height float64) { top += height }

	sectionHeader := func(title string) {
		label := makeLabel(title, appkit.Font_SystemFontOfSizeWeight(11, appkit.FontWeightSemibold),
			appkit.Color_SecondaryLabelColor())
		label.SetFrame(frame(settingsMargin, place(14), contentWidth, 14))
		content.AddSubview(label)
		gap(6)
	}

	groupBox := func(height float64) (appkit.Box, float64) {
		y := place(height)
		box := appkit.NewBox()
		box.SetBoxType(appkit.BoxCustom)
		box.SetTitlePosition(appkit.NoTitle)
		box.SetCornerRadius(8)
		box.SetBorderWidth(1)
		box.SetBorderColor(appkit.Color_LabelColor().ColorWithAlphaComponent(0.1))
		box.SetFillColor(appkit.Color_LabelColor().ColorWithAlphaComponent(0.05))
		box.SetFrame(frame(settingsMargin, y, contentWidth, height))
		content.AddSubview(box)
		return box, y
	}

	// --- Preview ---
	sectionHeader("PREVIEW")
	_, previewY := groupBox(36)
	preview := appkit.NewButton()
	preview.SetBordered(false)
	preview.SetFrame(frame(settingsMargin+8, previewY+6, contentWidth-16, 24))
	content.AddSubview(preview)
	sw.preview = preview
	gap(16)

	// --- Icon style cards ---
	sectionHeader("ICON STYLE")
	cardsY := place(72)
	cardWidth := (contentWidth - 2*12) / 3
	for i, style := range []struct {
		id   IconStyle
		name string
	}{
		{IconStyleHairline, "Hairline"},
		{IconStyleSolid, "Solid"},
		{IconStyleLabels, "Labels"},
	} {
		style := style
		card := appkit.NewButton()
		card.SetButtonType(appkit.ButtonTypePushOnPushOff)
		card.SetBezelStyle(appkit.BezelStyleRegularSquare)
		card.SetTitle(style.name)
		card.SetFont(appkit.Font_SystemFontOfSize(13))
		if style.id != IconStyleLabels {
			card.SetImage(menubarIcon(style.id, "temperature"))
			card.SetImagePosition(appkit.ImageAbove)
		}
		card.SetFrame(frame(settingsMargin+float64(i)*(cardWidth+12), cardsY, cardWidth, 72))
		action.Set(card, func(sender objc.Object) {
			sb.settings.IconStyle = style.id
			sb.applySettings()
		})
		content.AddSubview(card)
		sw.cardButtons[style.id] = card
	}
	gap(16)

	// --- Metric visibility ---
	sectionHeader("SHOW IN MENU BAR")
	rowHeight := 36.0
	boxHeight := rowHeight * float64(len(allMetrics))
	_, metricsY := groupBox(boxHeight)
	for i, info := range allMetrics {
		info := info
		rowY := metricsY + boxHeight - float64(i+1)*rowHeight

		name := makeLabel(info.Name, appkit.Font_SystemFontOfSize(13), appkit.Color_LabelColor())
		name.SetFrame(frame(settingsMargin+16, rowY+9, 180, 18))
		content.AddSubview(name)

		value := makeLabel("—", appkit.Font_MonospacedDigitSystemFontOfSizeWeight(13, appkit.FontWeightRegular),
			appkit.Color_SecondaryLabelColor())
		value.SetAlignment(appkit.TextAlignmentRight)
		value.SetFrame(frame(settingsWidth-settingsMargin-200, rowY+9, 120, 18))
		content.AddSubview(value)
		sw.valueLabels[info.ID] = value

		toggle := appkit.NewSwitch()
		toggle.SetFrame(frame(settingsWidth-settingsMargin-62, rowY+7, 46, 22))
		action.Set(toggle, func(sender objc.Object) {
			sb.settings.Visible[info.ID] = toggle.State() == appkit.ControlStateValueOn
			sb.applySettings()
		})
		content.AddSubview(toggle)
		sw.metricSwitches[info.ID] = toggle
	}
	gap(6)

	footnote := makeLabel("Hidden metrics stay available in the dropdown.",
		appkit.Font_SystemFontOfSize(11), appkit.Color_TertiaryLabelColor())
	footnote.SetFrame(frame(settingsMargin, place(14), contentWidth, 14))
	content.AddSubview(footnote)
	gap(16)

	// --- Formatting ---
	sectionHeader("FORMATTING")
	formatRow := 44.0
	formatRows := []struct {
		title    string
		subtitle string
		assign   func(appkit.Switch)
		onToggle func(on bool)
	}{
		{
			"Round Values", "One decimal for temperature, whole numbers elsewhere",
			func(s appkit.Switch) { sw.roundSwitch = s },
			func(on bool) { sb.settings.RoundValues = on; sb.applySettings() },
		},
		{
			"Monospaced Digits", "Keeps the bar from shifting on refresh",
			func(s appkit.Switch) { sw.tabularSwitch = s },
			func(on bool) { sb.settings.TabularDigits = on; sb.applySettings() },
		},
		{
			"Launch at Login", "Starts AirDash automatically when you log in",
			func(s appkit.Switch) { sw.loginSwitch = s },
			func(on bool) { sw.setLaunchAtLogin(on) },
		},
	}
	formatHeight := formatRow * float64(len(formatRows))
	_, formatY := groupBox(formatHeight)
	for i, row := range formatRows {
		row := row
		rowY := formatY + formatHeight - float64(i+1)*formatRow

		title := makeLabel(row.title, appkit.Font_SystemFontOfSize(13), appkit.Color_LabelColor())
		title.SetFrame(frame(settingsMargin+16, rowY+22, 320, 18))
		content.AddSubview(title)

		subtitle := makeLabel(row.subtitle, appkit.Font_SystemFontOfSize(11), appkit.Color_SecondaryLabelColor())
		subtitle.SetFrame(frame(settingsMargin+16, rowY+7, 340, 14))
		content.AddSubview(subtitle)

		toggle := appkit.NewSwitch()
		toggle.SetFrame(frame(settingsWidth-settingsMargin-62, rowY+11, 46, 22))
		action.Set(toggle, func(sender objc.Object) {
			row.onToggle(toggle.State() == appkit.ControlStateValueOn)
		})
		content.AddSubview(toggle)
		row.assign(toggle)
	}
	gap(20)

	// --- Footer ---
	footerY := place(32)
	applyNote := makeLabel("Changes apply immediately.",
		appkit.Font_SystemFontOfSize(11), appkit.Color_SecondaryLabelColor())
	applyNote.SetFrame(frame(settingsMargin, footerY+8, 200, 16))
	content.AddSubview(applyNote)

	resetButton := appkit.NewButtonWithTitle("Reset")
	resetButton.SetBezelStyle(appkit.BezelStyleRounded)
	resetButton.SetFrame(frame(settingsWidth-settingsMargin-170, footerY, 80, 32))
	action.Set(resetButton, func(sender objc.Object) {
		sb.settings = defaultDisplaySettings()
		sb.applySettings()
	})
	content.AddSubview(resetButton)

	doneButton := appkit.NewButtonWithTitle("Done")
	doneButton.SetBezelStyle(appkit.BezelStyleRounded)
	doneButton.SetKeyEquivalent("\r")
	doneButton.SetFrame(frame(settingsWidth-settingsMargin-80, footerY, 80, 32))
	action.Set(doneButton, func(sender objc.Object) {
		window.Close()
	})
	content.AddSubview(doneButton)

	return sw
}

// setLaunchAtLogin toggles the LaunchAgent plist without touching launchctl,
// so the running instance is unaffected either way.
func (sw *settingsWindow) setLaunchAtLogin(on bool) {
	var err error
	switch {
	case on && !isDaemonInstalled():
		_, err = installDaemonFiles()
	case !on:
		err = removeDaemonPlist()
	}
	if err != nil {
		logger.Error("Toggling launch at login", "enabled", on, "error", err)
	}
	sw.loginSwitch.SetState(boolState(isDaemonInstalled()))
}

func makeLabel(text string, font appkit.Font, color appkit.Color) appkit.TextField {
	label := appkit.NewLabel(text)
	label.SetFont(font)
	label.SetTextColor(color)
	return label
}

func frame(x, y, w, h float64) foundation.Rect {
	return foundation.Rect{
		Origin: foundation.Point{X: x, Y: y},
		Size:   foundation.Size{Width: w, Height: h},
	}
}

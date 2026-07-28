package main

import (
	"embed"
	"strings"

	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/coregraphics"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"
)

//go:embed assets/menubar
var menubarAssets embed.FS

// Raw attribute names (NSFontAttributeName etc. are not exposed by darwinkit).
const (
	attrFont  = foundation.AttributedStringKey("NSFont")
	attrColor = foundation.AttributedStringKey("NSColor")
	attrKern  = foundation.AttributedStringKey("NSKern")
)

const (
	iconSize      = 16.0
	iconBaseline  = -3.0
	iconValueGap  = 5.0
	labelValueGap = 4.0
	groupGap      = 20.0
	labelGroupGap = 18.0
	valueFontSize = 14.0
	labelFontSize = 10.0
)

// menubarIcon loads one of the embedded SVG glyphs as a 16x16 template image
// so macOS tints it for light/dark mode and menu bar highlighting.
func menubarIcon(style IconStyle, name string) appkit.Image {
	data, err := menubarAssets.ReadFile("assets/menubar/" + string(style) + "/" + name + ".svg")
	if err != nil {
		logger.Error("Loading menubar icon", "style", style, "icon", name, "error", err)
		return appkit.Image{}
	}
	img := appkit.NewImageWithData(data)
	img.SetSize(foundation.Size{Width: iconSize, Height: iconSize})
	img.SetTemplate(true)
	return img
}

// trimmedIcon returns the glyph cropped to its horizontal artwork bounds. The
// SVGs are drawn on a 16x16 canvas with uneven side margins, so laying them
// out untrimmed makes the fixed gaps look inconsistent between metrics.
var trimmedIconCache = map[string]appkit.Image{}

func trimmedIcon(style IconStyle, name string) appkit.Image {
	key := string(style) + "/" + name
	if img, ok := trimmedIconCache[key]; ok {
		return img
	}
	base := menubarIcon(style, name)

	rep := appkit.BitmapImageRep_ImageRepWithData(base.TIFFRepresentation())
	pixelsWide, pixelsHigh := rep.PixelsWide(), rep.PixelsHigh()
	minX, maxX := pixelsWide, -1
	for x := range pixelsWide {
		for y := range pixelsHigh {
			if rep.ColorAtXY(x, y).AlphaComponent() > 0.1 {
				minX, maxX = min(minX, x), max(maxX, x)
				break
			}
		}
	}
	trimmed := base
	if maxX >= 0 {
		scale := float64(pixelsWide) / iconSize
		left := float64(minX) / scale
		width := float64(maxX-minX+1) / scale
		trimmed = appkit.NewImageWithSize(foundation.Size{Width: width, Height: iconSize})
		objc.Call[objc.Void](trimmed, objc.Sel("lockFocus"))
		base.DrawInRectFromRectOperationFraction(frame(-left, 0, iconSize, iconSize),
			foundation.Rect{}, appkit.CompositingOperationSourceOver, 1)
		objc.Call[objc.Void](trimmed, objc.Sel("unlockFocus"))
		trimmed.SetTemplate(true)
	}
	objc.Retain(&trimmed)
	trimmedIconCache[key] = trimmed
	return trimmed
}

// tintedIcon renders the trimmed glyph in (approximately) the label color.
// NSStatusBarButton only auto-tints template images set directly on the
// button — attachments in an attributed title draw their raw (black)
// artwork — so we tint ourselves: 85% white/black matches how labelColor
// resolves in dark/light appearances.
func tintedIcon(style IconStyle, name string, dark bool) appkit.Image {
	tint := appkit.Color_ColorWithWhiteAlpha(0, 0.85)
	if dark {
		tint = appkit.Color_ColorWithWhiteAlpha(1, 0.85)
	}
	base := trimmedIcon(style, name)
	size := base.Size()
	img := appkit.NewImageWithSize(size)
	objc.Call[objc.Void](img, objc.Sel("lockFocus"))
	tint.SetFill()
	appkit.BezierPath_BezierPathWithRect(frame(0, 0, size.Width, size.Height)).Fill()
	base.DrawInRectFromRectOperationFraction(frame(0, 0, size.Width, size.Height),
		foundation.Rect{}, appkit.CompositingOperationDestinationIn, 1)
	objc.Call[objc.Void](img, objc.Sel("unlockFocus"))
	return img
}

func attributedSize(a foundation.IAttributedString) foundation.Size {
	return objc.Call[foundation.Size](a, objc.Sel("size"))
}

// isDarkAppearance reports whether a view currently renders in a dark
// appearance (covers both darkAqua and the vibrant menu bar variants).
func isDarkAppearance(view objc.IObject) bool {
	appearance := objc.Call[appkit.Appearance](view, objc.Sel("effectiveAppearance"))
	if appearance.IsNil() {
		return false
	}
	return strings.Contains(string(appearance.Name()), "Dark")
}

// gapRun returns a single space stretched (via kerning) to exactly width points.
func gapRun(width float64, font appkit.Font) foundation.AttributedString {
	space := foundation.NewAttributedStringWithStringAttributes(" ", map[foundation.AttributedStringKey]objc.IObject{
		attrFont: font,
	})
	kern := width - attributedSize(space).Width
	return foundation.NewAttributedStringWithStringAttributes(" ", map[foundation.AttributedStringKey]objc.IObject{
		attrFont: font,
		attrKern: foundation.Number_NumberWithDouble(kern),
	})
}

// buildStatusTitle renders measures into the status bar attributed string.
// Shared by the status item itself and the settings window live preview.
func buildStatusTitle(m AirGradientMeasures, tempUnit string, s DisplaySettings, dark bool) foundation.MutableAttributedString {
	segments := statusSegments(m, tempUnit, s)

	valueFont := appkit.Font_SystemFontOfSize(valueFontSize)
	if s.TabularDigits {
		valueFont = appkit.Font_MonospacedDigitSystemFontOfSizeWeight(valueFontSize, appkit.FontWeightRegular)
	}
	valueAttrs := map[foundation.AttributedStringKey]objc.IObject{
		attrFont:  valueFont,
		attrColor: appkit.Color_LabelColor(),
	}

	useLabels := s.IconStyle == IconStyleLabels
	labelAttrs := map[foundation.AttributedStringKey]objc.IObject{
		attrFont:  appkit.Font_SystemFontOfSizeWeight(labelFontSize, appkit.FontWeightSemibold),
		attrColor: appkit.Color_LabelColor().ColorWithAlphaComponent(0.72),
		// +0.06em tracking
		attrKern: foundation.Number_NumberWithDouble(0.06 * labelFontSize),
	}

	betweenGroups := groupGap
	if useLabels {
		betweenGroups = labelGroupGap
	}

	title := foundation.NewMutableAttributedString()
	if len(segments) == 0 {
		title.AppendAttributedString(foundation.NewAttributedStringWithStringAttributes("AirDash", valueAttrs))
		return title
	}
	for i, seg := range segments {
		if useLabels {
			title.AppendAttributedString(foundation.NewAttributedStringWithStringAttributes(seg.label, labelAttrs))
			title.AppendAttributedString(gapRun(labelValueGap, valueFont))
		} else {
			icon := tintedIcon(s.IconStyle, seg.icon, dark)
			attachment := appkit.NewTextAttachment()
			attachment.SetImage(icon)
			attachment.SetBounds(coregraphics.Rect{
				Origin: coregraphics.Point{X: 0, Y: iconBaseline},
				Size:   coregraphics.Size{Width: icon.Size().Width, Height: iconSize},
			})
			title.AppendAttributedString(foundation.AttributedString_AttributedStringWithAttachment(attachment))
			title.AppendAttributedString(gapRun(iconValueGap, valueFont))
		}
		title.AppendAttributedString(foundation.NewAttributedStringWithStringAttributes(seg.value, valueAttrs))
		if i < len(segments)-1 {
			title.AppendAttributedString(gapRun(betweenGroups, valueFont))
		}
	}
	return title
}

// statusBar owns the NSStatusItem, its menu, and the display state.
// All methods must be called on the main queue.
type statusBar struct {
	item     appkit.StatusItem
	cfg      *Config
	settings DisplaySettings
	measures *AirGradientMeasures

	readingItems map[MetricID]appkit.MenuItem
	styleItems   map[IconStyle]appkit.MenuItem
	settingsUI   *settingsWindow
}

func newStatusBar(cfg *Config) *statusBar {
	item := appkit.StatusBar_SystemStatusBar().StatusItemWithLength(-1)
	objc.Retain(&item)

	sb := &statusBar{
		item:         item,
		cfg:          cfg,
		settings:     loadDisplaySettings(),
		readingItems: map[MetricID]appkit.MenuItem{},
		styleItems:   map[IconStyle]appkit.MenuItem{},
	}
	item.SetMenu(sb.buildMenu())
	sb.render()
	return sb
}

func (sb *statusBar) setMeasures(m AirGradientMeasures) {
	sb.measures = &m
	sb.render()
	if sb.settingsUI != nil {
		sb.settingsUI.refresh()
	}
}

// render redraws the status item and dropdown readings from current state.
func (sb *statusBar) render() {
	for _, info := range allMetrics {
		value := "—"
		if sb.measures != nil {
			value = formatMetric(info.ID, *sb.measures, sb.cfg.TempUnit, sb.settings.RoundValues)
		}
		sb.readingItems[info.ID].SetTitle(info.Name + "   " + value)
	}

	if sb.measures == nil {
		sb.item.Button().SetTitle("…")
		return
	}
	// The icon tint must match the menu bar's appearance, which can differ
	// from the app's when the wallpaper tints the bar.
	button := sb.item.Button()
	title := buildStatusTitle(*sb.measures, sb.cfg.TempUnit, sb.settings, isDarkAppearance(button))
	button.SetAttributedTitle(title)
}

// applySettings persists the settings and refreshes every surface.
func (sb *statusBar) applySettings() {
	sb.settings.save()
	sb.render()
	sb.refreshMenuState()
	if sb.settingsUI != nil {
		sb.settingsUI.refresh()
	}
}

func (sb *statusBar) refreshMenuState() {
	for style, item := range sb.styleItems {
		state := appkit.ControlStateValueOff
		if style == sb.settings.IconStyle {
			state = appkit.ControlStateValueOn
		}
		item.SetState(state)
	}
}

func (sb *statusBar) buildMenu() appkit.Menu {
	menu := appkit.NewMenu()

	// Current readings for every metric — hidden ones stay available here.
	for _, info := range allMetrics {
		reading := appkit.NewMenuItem()
		reading.SetTitle(info.Name + "   —")
		menu.AddItem(reading) // no action: stays disabled, display-only
		sb.readingItems[info.ID] = reading
	}

	menu.AddItem(appkit.MenuItem_SeparatorItem())

	// Icon Style submenu (radio group) — the quick path for switching styles.
	styleMenu := appkit.NewMenu()
	for _, style := range []struct {
		id   IconStyle
		name string
	}{
		{IconStyleSolid, "Solid"},
		{IconStyleHairline, "Hairline"},
		{IconStyleLabels, "Labels"},
	} {
		style := style
		menuItem := appkit.NewMenuItemWithAction(style.name, "", func(sender objc.Object) {
			sb.settings.IconStyle = style.id
			sb.applySettings()
		})
		if style.id != IconStyleLabels {
			menuItem.SetImage(menubarIcon(style.id, "temperature"))
		}
		styleMenu.AddItem(menuItem)
		sb.styleItems[style.id] = menuItem
	}
	styleItem := appkit.NewMenuItem()
	styleItem.SetTitle("Icon Style")
	styleItem.SetSubmenu(styleMenu)
	menu.AddItem(styleItem)

	menu.AddItem(appkit.NewMenuItemWithAction("Settings…", ",", func(sender objc.Object) {
		sb.showSettingsWindow()
	}))

	menu.AddItem(appkit.MenuItem_SeparatorItem())

	menu.AddItem(appkit.NewMenuItemWithAction("About AirDash", "", func(sender objc.Object) {
		showAboutWindow()
	}))

	menu.AddItem(appkit.MenuItem_SeparatorItem())

	itemQuit := appkit.NewMenuItem()
	itemQuit.SetTitle("Quit")
	itemQuit.SetAction(objc.Sel("terminate:"))
	itemQuit.SetKeyEquivalent("q")
	menu.AddItem(itemQuit)

	sb.refreshMenuState()
	return menu
}

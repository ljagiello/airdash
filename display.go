package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/progrium/darwinkit/macos/foundation"
)

// IconStyle selects how metrics are rendered in the status bar.
type IconStyle string

const (
	IconStyleHairline IconStyle = "hairline"
	IconStyleSolid    IconStyle = "solid"
	IconStyleLabels   IconStyle = "labels"
)

type MetricID string

const (
	MetricTemperature MetricID = "temperature"
	MetricPM25        MetricID = "pm25"
	MetricHumidity    MetricID = "humidity"
	MetricCO2         MetricID = "co2"
	MetricHeatIndex   MetricID = "heat"
)

type metricInfo struct {
	ID   MetricID
	Name string // human name, used in the settings window and dropdown
	Abbr string // short label used by the Labels icon style
}

// allMetrics is the canonical metric order.
var allMetrics = []metricInfo{
	{MetricTemperature, "Temperature", "T"},
	{MetricPM25, "PM2.5", "PM"},
	{MetricHumidity, "Humidity", "RH"},
	{MetricCO2, "CO₂", "CO2"},
	{MetricHeatIndex, "Heat Index", "HI"},
}

func metricByID(id MetricID) metricInfo {
	for _, m := range allMetrics {
		if m.ID == id {
			return m
		}
	}
	return metricInfo{ID: id, Name: string(id), Abbr: strings.ToUpper(string(id))}
}

// DisplaySettings are the user-tunable status bar options, persisted in
// NSUserDefaults (independent from the connection config in config.yaml).
type DisplaySettings struct {
	IconStyle     IconStyle
	RoundValues   bool
	TabularDigits bool
	Visible       map[MetricID]bool
}

const (
	prefKeyIconStyle     = "iconStyle"
	prefKeyRoundValues   = "roundedValues"
	prefKeyTabularDigits = "tabularFigures"
	prefKeyMetrics       = "metrics"
)

func defaultDisplaySettings() DisplaySettings {
	return DisplaySettings{
		IconStyle:     IconStyleSolid,
		RoundValues:   true,
		TabularDigits: true,
		Visible: map[MetricID]bool{
			MetricTemperature: true,
			MetricPM25:        true,
			MetricHumidity:    true,
			MetricCO2:         true,
			MetricHeatIndex:   false,
		},
	}
}

// encodeMetrics serializes visibility as "temperature,pm25,humidity,co2,!heat"
// (a "!" prefix marks a hidden metric).
func encodeMetrics(visible map[MetricID]bool) string {
	parts := make([]string, 0, len(allMetrics))
	for _, m := range allMetrics {
		if visible[m.ID] {
			parts = append(parts, string(m.ID))
		} else {
			parts = append(parts, "!"+string(m.ID))
		}
	}
	return strings.Join(parts, ",")
}

func decodeMetrics(encoded string) map[MetricID]bool {
	visible := defaultDisplaySettings().Visible
	for _, part := range strings.Split(encoded, ",") {
		part = strings.TrimSpace(part)
		hidden := strings.HasPrefix(part, "!")
		id := MetricID(strings.TrimPrefix(part, "!"))
		if _, known := visible[id]; known {
			visible[id] = !hidden
		}
	}
	return visible
}

func loadDisplaySettings() DisplaySettings {
	s := defaultDisplaySettings()
	ud := foundation.UserDefaults_StandardUserDefaults()
	switch IconStyle(ud.StringForKey(prefKeyIconStyle)) {
	case IconStyleHairline:
		s.IconStyle = IconStyleHairline
	case IconStyleLabels:
		s.IconStyle = IconStyleLabels
	case IconStyleSolid:
		s.IconStyle = IconStyleSolid
	}
	if !ud.ObjectForKey(prefKeyRoundValues).IsNil() {
		s.RoundValues = ud.BoolForKey(prefKeyRoundValues)
	}
	if !ud.ObjectForKey(prefKeyTabularDigits).IsNil() {
		s.TabularDigits = ud.BoolForKey(prefKeyTabularDigits)
	}
	if encoded := ud.StringForKey(prefKeyMetrics); encoded != "" {
		s.Visible = decodeMetrics(encoded)
	}
	return s
}

func (s DisplaySettings) save() {
	ud := foundation.UserDefaults_StandardUserDefaults()
	ud.SetObjectForKey(foundation.String_StringWithString(string(s.IconStyle)), prefKeyIconStyle)
	ud.SetBoolForKey(s.RoundValues, prefKeyRoundValues)
	ud.SetBoolForKey(s.TabularDigits, prefKeyTabularDigits)
	ud.SetObjectForKey(foundation.String_StringWithString(encodeMetrics(s.Visible)), prefKeyMetrics)
}

// heatIndex returns the NOAA heat index (Rothfusz regression) in Celsius,
// given air temperature in Celsius and relative humidity in percent.
func heatIndex(tempC, rh float64) float64 {
	t := tempC*9/5 + 32 // regression is defined in °F

	// Simple formula, averaged with the temperature; below 80°F it's the result.
	simple := 0.5 * (t + 61.0 + (t-68.0)*1.2 + rh*0.094)
	hi := (simple + t) / 2
	if hi >= 80 {
		hi = -42.379 + 2.04901523*t + 10.14333127*rh -
			0.22475541*t*rh - 0.00683783*t*t - 0.05481717*rh*rh +
			0.00122874*t*t*rh + 0.00085282*t*rh*rh - 0.00000199*t*t*rh*rh
		switch {
		case rh < 13 && t >= 80 && t <= 112:
			hi -= ((13 - rh) / 4) * math.Sqrt((17-math.Abs(t-95))/17)
		case rh > 85 && t >= 80 && t <= 87:
			hi += ((rh - 85) / 10) * ((87 - t) / 2)
		}
	}
	return (hi - 32) * 5 / 9
}

// formatMetric renders one metric's value for display.
func formatMetric(id MetricID, m AirGradientMeasures, tempUnit string, round bool) string {
	degrees := func(c float64) string {
		v := convertTemperature(c, tempUnit)
		if round {
			return fmt.Sprintf("%.1f°", v)
		}
		return fmt.Sprintf("%.2f°", v)
	}
	switch id {
	case MetricTemperature:
		return degrees(m.Atmp)
	case MetricPM25:
		return fmt.Sprintf("%.0f", m.Pm02)
	case MetricHumidity:
		if round {
			return fmt.Sprintf("%.0f%%", m.Rhum)
		}
		return fmt.Sprintf("%.1f%%", m.Rhum)
	case MetricCO2:
		return fmt.Sprintf("%.0f", m.Rco2)
	case MetricHeatIndex:
		return degrees(heatIndex(m.Atmp, m.Rhum))
	}
	return ""
}

// metricSegment is one icon+value group in the status bar.
type metricSegment struct {
	icon  string // SVG base name in assets/menubar/<style>/
	label string // short label used by the Labels style
	value string
}

// statusSegments formats the visible metrics into display segments.
func statusSegments(m AirGradientMeasures, tempUnit string, s DisplaySettings) []metricSegment {
	var segments []metricSegment
	for _, info := range allMetrics {
		if !s.Visible[info.ID] {
			continue
		}
		segments = append(segments, metricSegment{
			icon:  string(info.ID),
			label: info.Abbr,
			value: formatMetric(info.ID, m, tempUnit, s.RoundValues),
		})
	}
	return segments
}

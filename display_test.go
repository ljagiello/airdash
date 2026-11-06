package main

import (
	"testing"
)

var testMeasures = AirGradientMeasures{Atmp: 24.5, Pm02: 4, Rhum: 66.4, Rco2: 679}

func TestStatusSegmentsDefaults(t *testing.T) {
	segments := statusSegments(testMeasures, "C", defaultDisplaySettings())

	want := []metricSegment{
		{icon: "temperature", label: "T", value: "24.5°"},
		{icon: "pm25", label: "PM", value: "4"},
		{icon: "humidity", label: "RH", value: "66%"},
		{icon: "co2", label: "CO2", value: "679"},
	}
	if len(segments) != len(want) {
		t.Fatalf("got %d segments, want %d: %+v", len(segments), len(want), segments)
	}
	for i, w := range want {
		if segments[i] != w {
			t.Errorf("segment %d = %+v, want %+v", i, segments[i], w)
		}
	}
}

func TestStatusSegmentsUnrounded(t *testing.T) {
	s := defaultDisplaySettings()
	s.RoundValues = false
	segments := statusSegments(testMeasures, "C", s)

	if segments[0].value != "24.50°" {
		t.Errorf("temperature = %q, want %q", segments[0].value, "24.50°")
	}
	if segments[2].value != "66.4%" {
		t.Errorf("humidity = %q, want %q", segments[2].value, "66.4%")
	}
}

func TestStatusSegmentsVisibility(t *testing.T) {
	s := defaultDisplaySettings()
	s.Visible[MetricPM25] = false
	s.Visible[MetricHeatIndex] = true
	segments := statusSegments(testMeasures, "C", s)

	var icons []string
	for _, seg := range segments {
		icons = append(icons, seg.icon)
	}
	want := []string{"temperature", "humidity", "co2", "heat"}
	if len(icons) != len(want) {
		t.Fatalf("visible icons = %v, want %v", icons, want)
	}
	for i := range want {
		if icons[i] != want[i] {
			t.Fatalf("visible icons = %v, want %v", icons, want)
		}
	}
}

func TestMetricsEncodingRoundTrip(t *testing.T) {
	visible := defaultDisplaySettings().Visible
	visible[MetricCO2] = false
	visible[MetricHeatIndex] = true

	encoded := encodeMetrics(visible)
	decoded := decodeMetrics(encoded)
	for id, want := range visible {
		if decoded[id] != want {
			t.Errorf("round trip %s = %v, want %v (encoded %q)", id, decoded[id], want, encoded)
		}
	}
}

func TestDecodeMetricsIgnoresUnknown(t *testing.T) {
	decoded := decodeMetrics("bogus,!temperature,co2")
	if decoded[MetricTemperature] {
		t.Error("temperature should be hidden")
	}
	if !decoded[MetricCO2] {
		t.Error("co2 should be visible")
	}
	if _, exists := decoded["bogus"]; exists {
		t.Error("unknown metric should not be stored")
	}
}

func TestHeatIndex(t *testing.T) {
	// Hot and humid: 30°C / 70% RH ≈ 35°C heat index (NOAA table: 95°F).
	if hi := heatIndex(30, 70); hi < 33 || hi > 37 {
		t.Errorf("heatIndex(30, 70) = %.1f, want ~35", hi)
	}
	// Mild conditions: heat index stays close to air temperature.
	if hi := heatIndex(20, 50); hi < 18 || hi > 22 {
		t.Errorf("heatIndex(20, 50) = %.1f, want ~20", hi)
	}
}

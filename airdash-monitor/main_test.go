package main

import (
	"testing"
)

// Mock notification sender for testing
type mockNotificationSender struct {
	notifications []mockNotification
}

type mockNotification struct {
	title    string
	message  string
	priority int
}

var mockSender *mockNotificationSender

func (m *mockNotificationSender) sendNotification(config *MonitorConfig, title, message string, priority int) {
	m.notifications = append(m.notifications, mockNotification{
		title:    title,
		message:  message,
		priority: priority,
	})
}

func (m *mockNotificationSender) reset() {
	m.notifications = []mockNotification{}
}

func TestPM25AlertTriggersWhenExceedingThreshold(t *testing.T) {
	// Reset alert states
	pm25AlertActive = false
	co2AlertActive = false
	mockSender = &mockNotificationSender{}

	config := &MonitorConfig{
		PM25Threshold: 10.0,
		CO2Threshold:  750.0,
	}

	measures := &AirGradientMeasures{
		Pm02: 15.0, // Above threshold
		Rco2: 700.0,
	}

	// Simulate the alert logic
	checkAlerts(measures, config, mockSender.sendNotification)

	if len(mockSender.notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(mockSender.notifications))
	}

	if mockSender.notifications[0].title != "PM2.5 Alert" {
		t.Errorf("Expected 'PM2.5 Alert', got '%s'", mockSender.notifications[0].title)
	}

	if mockSender.notifications[0].priority != 1 {
		t.Errorf("Expected high priority (1), got %d", mockSender.notifications[0].priority)
	}

	if !pm25AlertActive {
		t.Error("PM2.5 alert should be active")
	}
}

func TestPM25AllClearWhenBelowThreshold(t *testing.T) {
	// Set initial state: alert is active
	pm25AlertActive = true
	co2AlertActive = false
	mockSender = &mockNotificationSender{}

	config := &MonitorConfig{
		PM25Threshold: 10.0,
		CO2Threshold:  750.0,
	}

	measures := &AirGradientMeasures{
		Pm02: 8.0, // Below threshold
		Rco2: 700.0,
	}

	checkAlerts(measures, config, mockSender.sendNotification)

	if len(mockSender.notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(mockSender.notifications))
	}

	if mockSender.notifications[0].title != "PM2.5 Normal" {
		t.Errorf("Expected 'PM2.5 Normal', got '%s'", mockSender.notifications[0].title)
	}

	if mockSender.notifications[0].priority != 0 {
		t.Errorf("Expected normal priority (0), got %d", mockSender.notifications[0].priority)
	}

	if pm25AlertActive {
		t.Error("PM2.5 alert should not be active")
	}
}

func TestCO2AlertTriggersWhenExceedingThreshold(t *testing.T) {
	// Reset alert states
	pm25AlertActive = false
	co2AlertActive = false
	mockSender = &mockNotificationSender{}

	config := &MonitorConfig{
		PM25Threshold: 10.0,
		CO2Threshold:  750.0,
	}

	measures := &AirGradientMeasures{
		Pm02: 5.0,
		Rco2: 800.0, // Above threshold
	}

	checkAlerts(measures, config, mockSender.sendNotification)

	if len(mockSender.notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(mockSender.notifications))
	}

	if mockSender.notifications[0].title != "CO2 Alert" {
		t.Errorf("Expected 'CO2 Alert', got '%s'", mockSender.notifications[0].title)
	}

	if mockSender.notifications[0].priority != 1 {
		t.Errorf("Expected high priority (1), got %d", mockSender.notifications[0].priority)
	}

	if !co2AlertActive {
		t.Error("CO2 alert should be active")
	}
}

func TestCO2AllClearWhenBelowThreshold(t *testing.T) {
	// Set initial state: alert is active
	pm25AlertActive = false
	co2AlertActive = true
	mockSender = &mockNotificationSender{}

	config := &MonitorConfig{
		PM25Threshold: 10.0,
		CO2Threshold:  750.0,
	}

	measures := &AirGradientMeasures{
		Pm02: 5.0,
		Rco2: 700.0, // Below threshold
	}

	checkAlerts(measures, config, mockSender.sendNotification)

	if len(mockSender.notifications) != 1 {
		t.Errorf("Expected 1 notification, got %d", len(mockSender.notifications))
	}

	if mockSender.notifications[0].title != "CO2 Normal" {
		t.Errorf("Expected 'CO2 Normal', got '%s'", mockSender.notifications[0].title)
	}

	if mockSender.notifications[0].priority != 0 {
		t.Errorf("Expected normal priority (0), got %d", mockSender.notifications[0].priority)
	}

	if co2AlertActive {
		t.Error("CO2 alert should not be active")
	}
}

func TestBothAlertsCanTriggerSimultaneously(t *testing.T) {
	// Reset alert states
	pm25AlertActive = false
	co2AlertActive = false
	mockSender = &mockNotificationSender{}

	config := &MonitorConfig{
		PM25Threshold: 10.0,
		CO2Threshold:  750.0,
	}

	measures := &AirGradientMeasures{
		Pm02: 15.0, // Above threshold
		Rco2: 800.0, // Above threshold
	}

	checkAlerts(measures, config, mockSender.sendNotification)

	if len(mockSender.notifications) != 2 {
		t.Errorf("Expected 2 notifications, got %d", len(mockSender.notifications))
	}

	// Check both alerts are active
	if !pm25AlertActive || !co2AlertActive {
		t.Error("Both PM2.5 and CO2 alerts should be active")
	}
}

func TestNoNotificationWhenAlreadyActive(t *testing.T) {
	// Set initial state: alert is already active
	pm25AlertActive = true
	co2AlertActive = false
	mockSender = &mockNotificationSender{}

	config := &MonitorConfig{
		PM25Threshold: 10.0,
		CO2Threshold:  750.0,
	}

	measures := &AirGradientMeasures{
		Pm02: 15.0, // Still above threshold
		Rco2: 700.0,
	}

	checkAlerts(measures, config, mockSender.sendNotification)

	// Should not send notification since alert is already active
	if len(mockSender.notifications) != 0 {
		t.Errorf("Expected 0 notifications (alert already active), got %d", len(mockSender.notifications))
	}
}

func TestNoNotificationWhenBelowThresholdAndNotActive(t *testing.T) {
	// Reset alert states
	pm25AlertActive = false
	co2AlertActive = false
	mockSender = &mockNotificationSender{}

	config := &MonitorConfig{
		PM25Threshold: 10.0,
		CO2Threshold:  750.0,
	}

	measures := &AirGradientMeasures{
		Pm02: 5.0, // Below threshold
		Rco2: 700.0, // Below threshold
	}

	checkAlerts(measures, config, mockSender.sendNotification)

	// Should not send notification since values are normal and no active alerts
	if len(mockSender.notifications) != 0 {
		t.Errorf("Expected 0 notifications, got %d", len(mockSender.notifications))
	}
}

// Helper function to test the alert logic
func checkAlerts(measures *AirGradientMeasures, config *MonitorConfig, sendFunc func(*MonitorConfig, string, string, int)) {
	// Check PM2.5
	if measures.Pm02 > config.PM25Threshold && !pm25AlertActive {
		pm25AlertActive = true
		sendFunc(
			config,
			"PM2.5 Alert",
			"PM2.5 is elevated",
			1,
		)
	} else if measures.Pm02 <= config.PM25Threshold && pm25AlertActive {
		pm25AlertActive = false
		sendFunc(
			config,
			"PM2.5 Normal",
			"PM2.5 has returned to safe levels",
			0,
		)
	}

	// Check CO2
	if measures.Rco2 > config.CO2Threshold && !co2AlertActive {
		co2AlertActive = true
		sendFunc(
			config,
			"CO2 Alert",
			"CO2 is elevated",
			1,
		)
	} else if measures.Rco2 <= config.CO2Threshold && co2AlertActive {
		co2AlertActive = false
		sendFunc(
			config,
			"CO2 Normal",
			"CO2 has returned to safe levels",
			0,
		)
	}
}

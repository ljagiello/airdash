package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type AirGradientConfig struct {
	Token      string `yaml:"token"`
	LocationID int    `yaml:"locationId"`
}

type MonitorConfig struct {
	PushoverUserKey  string  `yaml:"pushoverUserKey"`
	PushoverAPIToken string  `yaml:"pushoverApiToken"`
	PM25Threshold    float64 `yaml:"pm25Threshold"`
	CO2Threshold     float64 `yaml:"co2Threshold"`
	CheckInterval    int     `yaml:"checkInterval"` // minutes
}

type AirGradientMeasures struct {
	LocationID   int       `yaml:"locationId"`
	LocationName string    `yaml:"locationName"`
	Pm02         float64   `yaml:"pm02"`
	Atmp         float64   `yaml:"atmp"`
	Rhum         float64   `yaml:"rhum"`
	Rco2         float64   `yaml:"rco2"`
	Timestamp    time.Time `yaml:"timestamp"`
}

var pm25AlertActive = false
var co2AlertActive = false

func main() {
	log.Println("AirDash Monitor starting...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get home directory:", err)
	}

	// Load AirGradient config
	agConfig, err := loadAirGradientConfig(homeDir + "/.airdash/config.yaml")
	if err != nil {
		log.Fatal("Failed to load AirGradient config:", err)
	}

	// Load monitor config
	monConfig, err := loadMonitorConfig(homeDir + "/.airdash/monitor-config.yaml")
	if err != nil {
		log.Fatal("Failed to load monitor config:", err)
	}

	log.Printf("Monitoring PM2.5 levels. Threshold: %.1f μg/m³", monConfig.PM25Threshold)
	log.Printf("Monitoring CO2 levels. Threshold: %.0f ppm", monConfig.CO2Threshold)
	log.Printf("Checking every %d minutes", monConfig.CheckInterval)

	// Run immediately, then on interval
	checkAirQuality(agConfig, monConfig)

	ticker := time.NewTicker(time.Duration(monConfig.CheckInterval) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		checkAirQuality(agConfig, monConfig)
	}
}

func checkAirQuality(agConfig *AirGradientConfig, monConfig *MonitorConfig) {
	measures, err := fetchAirGradientData(agConfig)
	if err != nil {
		log.Printf("Error fetching air quality data: %v", err)
		return
	}

	log.Printf("Current PM2.5: %.1f μg/m³, CO2: %.0f ppm", measures.Pm02, measures.Rco2)

	// Check PM2.5
	if measures.Pm02 > monConfig.PM25Threshold && !pm25AlertActive {
		// PM2.5 exceeded threshold - send alert
		pm25AlertActive = true
		sendPushoverNotification(
			monConfig,
			"PM2.5 Alert",
			fmt.Sprintf("PM2.5 is elevated at %.1f μg/m³ (threshold: %.1f)\n\nCurrent readings:\n• PM2.5: %.1f μg/m³\n• CO2: %.0f ppm",
				measures.Pm02, monConfig.PM25Threshold, measures.Pm02, measures.Rco2),
			1, // high priority
		)
		log.Printf("⚠️  ALERT: PM2.5 exceeded threshold!")
	} else if measures.Pm02 <= monConfig.PM25Threshold && pm25AlertActive {
		// PM2.5 back to normal - send all clear
		pm25AlertActive = false
		sendPushoverNotification(
			monConfig,
			"PM2.5 Normal",
			fmt.Sprintf("PM2.5 has returned to safe levels at %.1f μg/m³\n\nCurrent readings:\n• PM2.5: %.1f μg/m³\n• CO2: %.0f ppm",
				measures.Pm02, measures.Pm02, measures.Rco2),
			0, // normal priority
		)
		log.Printf("✓ All clear: PM2.5 back to normal")
	}

	// Check CO2
	if measures.Rco2 > monConfig.CO2Threshold && !co2AlertActive {
		// CO2 exceeded threshold - send alert
		co2AlertActive = true
		sendPushoverNotification(
			monConfig,
			"CO2 Alert",
			fmt.Sprintf("CO2 is elevated at %.0f ppm (threshold: %.0f)\n\nCurrent readings:\n• PM2.5: %.1f μg/m³\n• CO2: %.0f ppm",
				measures.Rco2, monConfig.CO2Threshold, measures.Pm02, measures.Rco2),
			1, // high priority
		)
		log.Printf("⚠️  ALERT: CO2 exceeded threshold!")
	} else if measures.Rco2 <= monConfig.CO2Threshold && co2AlertActive {
		// CO2 back to normal - send all clear
		co2AlertActive = false
		sendPushoverNotification(
			monConfig,
			"CO2 Normal",
			fmt.Sprintf("CO2 has returned to safe levels at %.0f ppm\n\nCurrent readings:\n• PM2.5: %.1f μg/m³\n• CO2: %.0f ppm",
				measures.Rco2, measures.Pm02, measures.Rco2),
			0, // normal priority
		)
		log.Printf("✓ All clear: CO2 back to normal")
	}
}

func fetchAirGradientData(config *AirGradientConfig) (*AirGradientMeasures, error) {
	apiURL := fmt.Sprintf("https://api.airgradient.com/public/api/v1/locations/%d/measures/current", config.LocationID)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("token", config.Token)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// API can return single object or array
	var measures AirGradientMeasures
	if err := json.Unmarshal(body, &measures); err != nil {
		// Try array format
		var measuresArray []AirGradientMeasures
		if err := json.Unmarshal(body, &measuresArray); err != nil {
			return nil, err
		}
		if len(measuresArray) == 0 {
			return nil, fmt.Errorf("no measurements returned")
		}
		measures = measuresArray[0]
	}

	return &measures, nil
}

func sendPushoverNotification(config *MonitorConfig, title, message string, priority int) {
	data := url.Values{}
	data.Set("token", config.PushoverAPIToken)
	data.Set("user", config.PushoverUserKey)
	data.Set("title", title)
	data.Set("message", message)
	data.Set("priority", fmt.Sprintf("%d", priority))

	resp, err := http.PostForm("https://api.pushover.net/1/messages.json", data)
	if err != nil {
		log.Printf("Failed to send notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Pushover API error (status %d): %s", resp.StatusCode, body)
		return
	}

	log.Println("✓ Push notification sent successfully")
}

func loadAirGradientConfig(path string) (*AirGradientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config AirGradientConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func loadMonitorConfig(path string) (*MonitorConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config MonitorConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults
	if config.PM25Threshold == 0 {
		config.PM25Threshold = 10
	}
	if config.CO2Threshold == 0 {
		config.CO2Threshold = 750
	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 5
	}

	return &config, nil
}

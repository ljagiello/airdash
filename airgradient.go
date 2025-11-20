package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AirGradientMeasures struct {
	LocationID         int       `json:"locationId"`
	LocationName       string    `json:"locationName"`
	Pm01               float64   `json:"pm01"`
	Pm02               float64   `json:"pm02"`
	Pm10               float64   `json:"pm10"`
	Pm003Count         float64   `json:"pm003Count"`
	Atmp               float64   `json:"atmp"`
	Rhum               float64   `json:"rhum"`
	Rco2               float64   `json:"rco2"`
	Tvoc               float64   `json:"tvoc"`
	Wifi               float64   `json:"wifi"`
	Timestamp          time.Time `json:"timestamp"`
	LedMode            string    `json:"ledMode"`
	LedCo2Threshold1   float64   `json:"ledCo2Threshold1"`
	LedCo2Threshold2   float64   `json:"ledCo2Threshold2"`
	LedCo2ThresholdEnd float64   `json:"ledCo2ThresholdEnd"`
	Serialno           string    `json:"serialno"`
	FirmwareVersion    string    `json:"firmwareVersion"`
	TvocIndex          float64   `json:"tvocIndex"`
	NoxIndex           float64   `json:"noxIndex"`
}

var (
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
	ErrBadPayload = errors.New("Error unmarshalling JSON")
)

// getAirGradientAPIURL returns the AirGradient API URL
func getAirGradientAPIURL(locationID int) string {
	if locationID != 0 {
		return fmt.Sprintf("https://api.airgradient.com/public/api/v1/locations/%d/measures/current", locationID)
	}
	return "https://api.airgradient.com/public/api/v1/locations/measures/current"
}

// convertTemperature converts the temperature from Celsius to Fahrenheit if the
// temperature unit is set to Fahrenheit
// By default the temperature unit is Celsius
func convertTemperature(temperature float64, tempUnit string) float64 {
	if tempUnit == "F" {
		return (temperature * 9 / 5) + 32
	}
	return temperature
}

// fetchMeasures fetches the measures from the AirGradient API
func fetchMeasures(airGradientAPIUrl string, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", airGradientAPIUrl, nil)
	if err != nil {
		logger.Error("Creating HTTP request", "error", err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("token", token)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error("Sending HTTP request", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Reading HTTP request", "error", err)
		return nil, err
	}

	return body, nil
}

func getAirGradientMeasures(airGradientAPIUrl string, token string) (AirGradientMeasures, error) {
	var measures AirGradientMeasures
	payload, err := fetchMeasures(airGradientAPIUrl, token)
	if err != nil {
		return measures, err
	}

	// Try to unmarshal as a single object first
	if err := json.Unmarshal(payload, &measures); err == nil {
		// If it worked, verify it's not an empty/zero value masked as success?
		// Actually, if it's an array "[]", Unmarshal to struct might succeed but result in zero fields?
		// No, unmarshalling "[]" into a struct returns an error.
		// So if this succeeds, it's an object.
		return measures, nil
	}

	// If that failed, try as an array
	var arrayMeasures []AirGradientMeasures
	if err := json.Unmarshal(payload, &arrayMeasures); err == nil {
		if len(arrayMeasures) > 0 {
			return arrayMeasures[0], nil
		}
	}

	return measures, ErrBadPayload
}

package main

import (
	"io"
	"net/http"
	"time"
)

const AIR_GRADIENT_API_URL = "https://api.airgradient.com/public/api/v1/locations/measures/current"

type AirGradientMeasures []struct {
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

func fetchMeasures(token string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", AIR_GRADIENT_API_URL, nil)
	if err != nil {
		logger.Error("Creating HTTP request", "error", err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("token", token)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
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

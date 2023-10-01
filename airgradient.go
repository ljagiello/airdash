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
	Pm01               any       `json:"pm01"`
	Pm02               int       `json:"pm02"`
	Pm10               any       `json:"pm10"`
	Pm003Count         any       `json:"pm003Count"`
	Atmp               float64   `json:"atmp"`
	Rhum               int       `json:"rhum"`
	Rco2               int       `json:"rco2"`
	Tvoc               float64   `json:"tvoc"`
	Wifi               int       `json:"wifi"`
	Timestamp          time.Time `json:"timestamp"`
	LedMode            string    `json:"ledMode"`
	LedCo2Threshold1   int       `json:"ledCo2Threshold1"`
	LedCo2Threshold2   int       `json:"ledCo2Threshold2"`
	LedCo2ThresholdEnd int       `json:"ledCo2ThresholdEnd"`
	Serialno           string    `json:"serialno"`
	FirmwareVersion    any       `json:"firmwareVersion"`
	TvocIndex          int       `json:"tvocIndex"`
	NoxIndex           int       `json:"noxIndex"`
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

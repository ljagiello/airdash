package main

import (
	"io"
	"net/http"
)

const AIR_GRADIENT_API_URL = "https://api.airgradient.com/public/api/v1/locations/measures/current"

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

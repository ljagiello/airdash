package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAirGradientAPIURL(t *testing.T) {
	testCases := []struct {
		name        string
		locationID  int
		expectedURL string
	}{
		{
			"location-id-0",
			0,
			"https://api.airgradient.com/public/api/v1/locations/measures/current",
		},
		{
			"location-id-12345",
			12345,
			"https://api.airgradient.com/public/api/v1/locations/12345/measures/current",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			assert.Equal(t, tC.expectedURL, getAirGradientAPIURL(tC.locationID))
		})
	}
}

func TestConvertTemp(t *testing.T) {
	testCases := []struct {
		name     string
		temp     float64
		tempUnit string
		expected float64
	}{
		{
			"convert-celsius-to-fahrenheit",
			20,
			"F",
			68,
		},
		{
			"no-conversion",
			20,
			"C",
			20,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			assert.Equal(t, tC.expected, convertTemperature(tC.temp, tC.tempUnit))
		})
	}
}

func TestGetAirGradientMeasures(t *testing.T) {
	testCases := []struct {
		name        string
		payloadFile string
		err         error
	}{
		{
			"correct-api-v1-locations-measures-current",
			"testdata/api-v1-locations-measures-current.json",
			nil,
		},
		{
			"correct-api-v1-locations-measures-current-with-more-float64",
			"testdata/api-v1-locations-measures-current-with-more-float64.json",
			nil,
		},
		{
			"correct-api-v1-locations-12345-measures-current",
			"testdata/api-v1-locations-12345-measures-current.json",
			nil,
		},
		{
			"incorrect-response-404",
			"testdata/incorrect-response-404.json",
			errors.New("Error unmarshalling JSON"),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, tC.payloadFile)
			}))
			defer server.Close()

			_, err := getAirGradientMeasures(server.URL, "SECRET-TOKEN")
			assert.Equal(t, tC.err, err)
		})
	}
}

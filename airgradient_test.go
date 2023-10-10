package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAirGradientMeasures(t *testing.T) {
	var testCases = []struct {
		name        string
		payloadFile string
		err         error
	}{
		{
			"correct-response",
			"testdata/correct_response1.json",
			nil,
		},
		{
			"correct-response2",
			"testdata/correct_response2.json",
			nil,
		},
		{
			"incorrect-response",
			"testdata/incorrect_response1.json",
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

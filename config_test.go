package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func CreateTestConfig(t *testing.T, configContent []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	require.NoError(t, os.WriteFile(path, configContent, 0o600))
	return path
}

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		name          string
		configContent []byte
		token         string
		locationID    int
		interval      int
		tempUnit      string
		err           error
	}{
		{
			"full-token",
			[]byte("token: \"1234567890\"\nlocationId: 0\ninterval: 60\ntempUnit: \"C\""),
			"1234567890",
			0,
			60,
			"C",
			nil,
		},
		{
			"missing-interval",
			[]byte("token: \"1234567890\"\ntempUnit: \"F\""),
			"1234567890",
			0,
			0,
			"F",
			nil,
		},
		{
			"invalid-config",
			[]byte(`foobar-invalid`),
			"",
			0,
			0,
			"",
			&yaml.TypeError{Errors: []string{"line 1: cannot unmarshal !!str `foobar-...` into main.Config"}},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			configPath := CreateTestConfig(t, tC.configContent)

			cfg, err := LoadConfig(configPath)
			if err == nil {
				assert.Equal(t, tC.token, cfg.Token)
				assert.Equal(t, tC.locationID, cfg.LocationID)
				assert.Equal(t, tC.interval, cfg.Interval)
				assert.Equal(t, tC.tempUnit, cfg.TempUnit)
			}
			assert.Equal(t, tC.err, err)
		})
	}
}

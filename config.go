package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Token    string `yaml:"token"`
	Interval int    `yaml:"interval"`
	TempUnit string `yaml:"tempUnit"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

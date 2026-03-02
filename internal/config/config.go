package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type BackendConfig struct {
	URL        string `json:"url" yaml:"url"`
	Weight     int    `json:"weight" yaml:"weight"`
	HealthPath string `json:"health_path" yaml:"health_path"`
}

type Config struct {
	Port                string          `json:"port" yaml:"port"`
	HealthCheckInterval int             `json:"health_check_interval" yaml:"health_check_interval"`
	Backends            []BackendConfig `json:"backends" yaml:"backends"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	switch {
	case strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml"):
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	default:
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}

	}
	if cfg.Port == "" {
		cfg.Port = ":3030"
	}
	if cfg.HealthCheckInterval <= 0 {
		cfg.HealthCheckInterval = 2
	}
	for i, b := range cfg.Backends {
		if b.URL == "" {
			return nil, fmt.Errorf("Backend %d missing URL", i)
		}
		if b.Weight <= 0 {
			cfg.Backends[i].Weight = 1
		}
		if b.HealthPath == "" {
			cfg.Backends[i].HealthPath = "/healthz"
		}
	}

	return cfg, nil
}

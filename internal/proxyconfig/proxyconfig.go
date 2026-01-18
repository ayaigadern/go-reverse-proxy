package proxyconfig

import (
	"encoding/json"
	"os"
	"time"
)

type ProxyConfig struct {
	Port            int           `json:"port"`
	Strategy        string        `json:"strategy"`
	HealthCheckFreq time.Duration `json:"health_check_frequency"`
	Backends        []string      `json:"backends"`
}

func LoadConfig(file string) (*ProxyConfig, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Unmarshal with a temporary struct to parse duration
	tmp := struct {
		Port            int      `json:"port"`
		Strategy        string   `json:"strategy"`
		HealthCheckFreq string   `json:"health_check_frequency"`
		Backends        []string `json:"backends"`
	}{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return nil, err
	}

	duration, err := time.ParseDuration(tmp.HealthCheckFreq)
	if err != nil {
		return nil, err
	}

	return &ProxyConfig{
		Port:            tmp.Port,
		Strategy:        tmp.Strategy,
		HealthCheckFreq: duration,
		Backends:        tmp.Backends,
	}, nil
}

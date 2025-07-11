package config

import (
	"encoding/json"
	"os"
)

type AgentConfig struct {
	Address        string `json:"address"`
	ReportInterval int    `json:"report_interval"`
	PollInterval   int    `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
	Key            string `json:"key"`
	RateLimit      int    `json:"rate_limit"`
}

func LoadAgentConfig(filePath string) (*AgentConfig, error) {
	config := &AgentConfig{
		Address:        "localhost:8080",
		ReportInterval: 10,
		PollInterval:   2,
		RateLimit:      2,
	}

	if filePath != "" {
		file, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(file, config)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

package config

import (
	"encoding/json"
	"os"
	"strconv"
)

// ServerConfig represents the configuration for the server
type ServerConfig struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval int    `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	Key           string `json:"key"`
}

// LoadServerConfig loads configuration from a JSON file
func LoadServerConfig(filePath string) (*ServerConfig, error) {
	// Default configuration
	config := &ServerConfig{
		Address:       "localhost:8080",
		Restore:       true,
		StoreInterval: 300,
		StoreFile:     "/tmp/metrics-db.json",
	}

	// Read and parse the config file if it exists
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

// GetStoreIntervalAsBool converts the StoreInterval string to a boolean value
func GetBoolFromString(value string) (bool, error) {
	return strconv.ParseBool(value)
}

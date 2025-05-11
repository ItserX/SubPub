package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server struct {
		Port string `json:"port"`
		Host string `json:"host"`
	} `json:"server"`
	Log struct {
		Level string `json:"level"`
		File  string `json:"file"`
	} `json:"log"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	if config.Server.Port == "" {
		config.Server.Port = "50051"
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Log.Level == "" {
		config.Log.Level = "info"
	}

	return &config, nil
}

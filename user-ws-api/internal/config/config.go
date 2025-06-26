package config

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	Database struct {
		Driver string `yaml:"driver"`
		URL    string `yaml:"url"`
	} `yaml:"database"`

	NATS struct {
		URL string `yaml:"url"`
	} `yaml:"nats"`
}

var AppConfig Config

func LoadConfig(path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Failed to read config file: %v", err)
		os.Exit(1)
	}

	if err := yaml.Unmarshal(file, &AppConfig); err != nil {
		slog.Error("Failed to parse config file: %v", err)
		os.Exit(1)
	}
	if AppConfig.Server.Port == "" {
		slog.Error("Server port is required")
		os.Exit(1)
	}
	if AppConfig.Database.URL == "" {
		slog.Error("Database URL is required")
		os.Exit(1)
	}
}

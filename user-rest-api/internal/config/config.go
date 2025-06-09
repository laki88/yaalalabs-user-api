package config

import (
	"log"
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
		log.Fatalf("Failed to read config file: %v", err)
	}

	if err := yaml.Unmarshal(file, &AppConfig); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
}

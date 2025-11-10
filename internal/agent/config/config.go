package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	API       APIConfig       `yaml:"api"`
	Reporting ReportingConfig `yaml:"reporting"`
	Logging   LoggingConfig   `yaml:"logging"`
}

type ServerConfig struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
}

type APIConfig struct {
	Endpoint string `yaml:"endpoint"`
	AgentKey string `yaml:"agent_key"`
}

type ReportingConfig struct {
	Interval int `yaml:"interval"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

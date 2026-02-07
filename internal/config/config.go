package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Discord  DiscordConfig  `yaml:"discord"`
	Claude   ClaudeConfig   `yaml:"claude"`
	Schedule ScheduleConfig `yaml:"schedule"`
	Database DatabaseConfig `yaml:"database"`
}

type DiscordConfig struct {
	Token     string `yaml:"token"`
	ChannelID string `yaml:"channel_id"`
}

type ClaudeConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

type ScheduleConfig struct {
	IntervalMinutes int `yaml:"interval_minutes"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// Load reads the config file and returns a Config struct
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

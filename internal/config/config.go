package config

import (
	"os"
	"strconv"
)

type Config struct {
	Discord  DiscordConfig
	Claude   ClaudeConfig
	Schedule ScheduleConfig
	Database DatabaseConfig
}

type DiscordConfig struct {
	Token     string
	ChannelID string
}

type ClaudeConfig struct {
	APIKey string
	Model  string
}

type ScheduleConfig struct {
	IntervalMinutes int
}

type DatabaseConfig struct {
	Path string
}

// Load reads configuration from environment variables
func Load() *Config {
	interval := 60
	if v := os.Getenv("SCHEDULE_INTERVAL"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			interval = parsed
		}
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./english_quiz.db"
	}

	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	return &Config{
		Discord: DiscordConfig{
			Token:     os.Getenv("DISCORD_TOKEN"),
			ChannelID: os.Getenv("DISCORD_CHANNEL_ID"),
		},
		Claude: ClaudeConfig{
			APIKey: os.Getenv("CLAUDE_API_KEY"),
			Model:  model,
		},
		Schedule: ScheduleConfig{
			IntervalMinutes: interval,
		},
		Database: DatabaseConfig{
			Path: dbPath,
		},
	}
}

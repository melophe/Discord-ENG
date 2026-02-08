package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set environment variables
	os.Setenv("DISCORD_TOKEN", "test-token")
	os.Setenv("DISCORD_CHANNEL_ID", "123456789")
	os.Setenv("CLAUDE_API_KEY", "test-api-key")
	os.Setenv("CLAUDE_MODEL", "claude-sonnet-4-20250514")
	os.Setenv("SCHEDULE_INTERVAL", "30")
	os.Setenv("DATABASE_PATH", "./test.db")
	defer func() {
		os.Unsetenv("DISCORD_TOKEN")
		os.Unsetenv("DISCORD_CHANNEL_ID")
		os.Unsetenv("CLAUDE_API_KEY")
		os.Unsetenv("CLAUDE_MODEL")
		os.Unsetenv("SCHEDULE_INTERVAL")
		os.Unsetenv("DATABASE_PATH")
	}()

	cfg := Load()

	if cfg.Discord.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", cfg.Discord.Token)
	}
	if cfg.Discord.ChannelID != "123456789" {
		t.Errorf("Expected channel_id '123456789', got '%s'", cfg.Discord.ChannelID)
	}
	if cfg.Claude.APIKey != "test-api-key" {
		t.Errorf("Expected api_key 'test-api-key', got '%s'", cfg.Claude.APIKey)
	}
	if cfg.Claude.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Expected model 'claude-sonnet-4-20250514', got '%s'", cfg.Claude.Model)
	}
	if cfg.Schedule.IntervalMinutes != 30 {
		t.Errorf("Expected interval_minutes 30, got %d", cfg.Schedule.IntervalMinutes)
	}
	if cfg.Database.Path != "./test.db" {
		t.Errorf("Expected path './test.db', got '%s'", cfg.Database.Path)
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Clear all env vars
	os.Unsetenv("DISCORD_TOKEN")
	os.Unsetenv("DISCORD_CHANNEL_ID")
	os.Unsetenv("CLAUDE_API_KEY")
	os.Unsetenv("CLAUDE_MODEL")
	os.Unsetenv("SCHEDULE_INTERVAL")
	os.Unsetenv("DATABASE_PATH")

	cfg := Load()

	// Check defaults
	if cfg.Claude.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Expected default model 'claude-sonnet-4-20250514', got '%s'", cfg.Claude.Model)
	}
	if cfg.Schedule.IntervalMinutes != 60 {
		t.Errorf("Expected default interval 60, got %d", cfg.Schedule.IntervalMinutes)
	}
	if cfg.Database.Path != "./english_quiz.db" {
		t.Errorf("Expected default path './english_quiz.db', got '%s'", cfg.Database.Path)
	}
}

func TestLoad_InvalidInterval(t *testing.T) {
	os.Setenv("SCHEDULE_INTERVAL", "invalid")
	defer os.Unsetenv("SCHEDULE_INTERVAL")

	cfg := Load()

	// Should use default when invalid
	if cfg.Schedule.IntervalMinutes != 60 {
		t.Errorf("Expected default interval 60 for invalid input, got %d", cfg.Schedule.IntervalMinutes)
	}
}

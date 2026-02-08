package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	content := `
discord:
  token: "test-token"
  channel_id: "123456789"

claude:
  api_key: "test-api-key"
  model: "claude-sonnet-4-20250514"

schedule:
  interval_minutes: 30

database:
  path: "./test.db"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Test loading
	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
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

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid YAML
	if _, err := tmpFile.WriteString("invalid: yaml: content: [}"); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	_, err = Load(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

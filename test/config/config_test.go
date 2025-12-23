package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ludwig/internal/config"
)

func TestLoadConfigFileNotExists(t *testing.T) {
	// When config file doesn't exist, should return nil without error
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Errorf("expected no error when file doesn't exist, got %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config when file doesn't exist, got %v", cfg)
	}
}

func TestLoadConfigValid(t *testing.T) {
	// Create a temporary config file
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	configDir := filepath.Join(home, ".ai-orchestrator")
	configFile := filepath.Join(configDir, "config.json")

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write test config
	testConfig := config.Config{DelayMs: 500}
	data, _ := json.Marshal(testConfig)
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	defer os.Remove(configFile)

	// Load and verify
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Errorf("expected config, got nil")
	}
	if cfg.DelayMs != 500 {
		t.Errorf("expected DelayMs 500, got %d", cfg.DelayMs)
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	configDir := filepath.Join(home, ".ai-orchestrator")
	configFile := filepath.Join(configDir, "config.json")

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Write invalid JSON
	if err := os.WriteFile(configFile, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	defer os.Remove(configFile)

	// Should error on invalid JSON
	cfg, err := config.LoadConfig()
	if err == nil {
		t.Errorf("expected error on invalid JSON, got nil")
	}
	if cfg != nil {
		t.Errorf("expected nil config on error, got %v", cfg)
	}
}

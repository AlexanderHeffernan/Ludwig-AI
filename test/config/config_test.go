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
	// Create a temporary config file in .ludwig directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}

	ludwigDir := filepath.Join(cwd, ".ludwig")
	configFile := filepath.Join(ludwigDir, "config.json")

	// Ensure directory exists
	if err := os.MkdirAll(ludwigDir, 0755); err != nil {
		t.Fatalf("failed to create .ludwig dir: %v", err)
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
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}

	ludwigDir := filepath.Join(cwd, ".ludwig")
	configFile := filepath.Join(ludwigDir, "config.json")

	// Ensure directory exists
	if err := os.MkdirAll(ludwigDir, 0755); err != nil {
		t.Fatalf("failed to create .ludwig dir: %v", err)
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

func TestSaveConfig(t *testing.T) {
	// Create config and save it
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}

	ludwigDir := filepath.Join(cwd, ".ludwig")
	configFile := filepath.Join(ludwigDir, "config.json")

	// Cleanup before test
	defer os.Remove(configFile)

	testConfig := &config.Config{
		DelayMs:       1000,
		AIProvider:    "ollama",
		OllamaBaseURL: "http://localhost:11434",
		OllamaModel:   "mistral",
	}

	// Save config
	if err := config.SaveConfig(testConfig); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configFile); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Load and verify content
	loadedCfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}
	if loadedCfg == nil {
		t.Fatalf("expected config, got nil")
	}
	if loadedCfg.DelayMs != 1000 {
		t.Errorf("expected DelayMs 1000, got %d", loadedCfg.DelayMs)
	}
	if loadedCfg.AIProvider != "ollama" {
		t.Errorf("expected AIProvider 'ollama', got %s", loadedCfg.AIProvider)
	}
	if loadedCfg.OllamaModel != "mistral" {
		t.Errorf("expected OllamaModel 'mistral', got %s", loadedCfg.OllamaModel)
	}
}

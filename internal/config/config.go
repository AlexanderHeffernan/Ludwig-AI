package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config represents the user's configuration
type Config struct {
	DelayMs int `json:"delayMs"` // Minimum delay in milliseconds between requests
}

// LoadConfig loads configuration from ~/.ai-orchestrator/config.json
// Returns nil if file doesn't exist (which is fine - optional config)
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".ai-orchestrator", "config.json")

	// File doesn't exist - that's okay
	file, err := os.Open(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

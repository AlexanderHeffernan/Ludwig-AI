package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the user's configuration
type Config struct {
	DelayMs    int    `json:"delayMs"`    // Minimum delay in milliseconds between requests
	AIProvider string `json:"aiProvider"` // "gemini" (default), "ollama", or "copilot"
	// Ollama-specific settings
	OllamaBaseURL string `json:"ollamaBaseURL"` // Base URL for Ollama (default: http://localhost:11434)
	OllamaModel   string `json:"ollamaModel"`   // Model name for Ollama (default: mistral)
	// Copilot-specific settings
	CopilotModel string `json:"copilotModel"` // Model name for Copilot (default: gpt-5)
}

// LoadConfig loads configuration from .ludwig/config.json in the current project
// Returns nil if file doesn't exist (which is fine - optional config)
func LoadConfig() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	ludwigDir := filepath.Join(cwd, ".ludwig")
	configPath := filepath.Join(ludwigDir, "config.json")

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

// SaveConfig saves configuration to .ludwig/config.json in the current project
func SaveConfig(cfg *Config) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	ludwigDir := filepath.Join(cwd, ".ludwig")
	if err := os.MkdirAll(ludwigDir, 0755); err != nil {
		return fmt.Errorf("failed to create .ludwig directory: %w", err)
	}

	configPath := filepath.Join(ludwigDir, "config.json")
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(cfg); err != nil {
		return err
	}

	return nil
}

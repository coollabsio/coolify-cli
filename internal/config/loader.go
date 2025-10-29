package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// LoadFromFile loads config from a specific file path
func LoadFromFile(path string) (*Config, error) {
	// Check if file exists
	if !fileExists(path) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.path = path
	return &cfg, nil
}

// SaveToFile saves config to a specific file path
func SaveToFile(path string, cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists checks if the config file exists at the default location
func Exists() bool {
	return fileExists(Path())
}

// CreateDefault creates a default config file with cloud and localhost instances
func CreateDefault() error {
	cfg := New()

	// Add default cloud instance
	cfg.Instances = append(cfg.Instances, Instance{
		Name:    "cloud",
		FQDN:    "https://app.coolify.io",
		Token:   "",
		Default: true,
	})

	// Add localhost instance
	cfg.Instances = append(cfg.Instances, Instance{
		Name:  "localhost",
		FQDN:  "http://localhost:8000",
		Token: "root",
	})

	return cfg.Save()
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

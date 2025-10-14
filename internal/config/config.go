package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
)

// Config holds all CLI configuration
type Config struct {
	Instances           []Instance `json:"instances"`
	LastUpdateCheckTime string     `json:"lastUpdateCheckTime"`
	path                string     // config file path (not serialized)
}

// New creates a new config with default values
func New() *Config {
	return &Config{
		Instances:           []Instance{},
		LastUpdateCheckTime: time.Now().Format(time.RFC3339),
		path:                Path(),
	}
}

// Load loads config from the default location
func Load() (*Config, error) {
	return LoadFromFile(Path())
}

// Save saves config to the default location
func (c *Config) Save() error {
	c.path = Path()
	return SaveToFile(c.path, c)
}

// GetDefault returns the default instance
func (c *Config) GetDefault() (*Instance, error) {
	for i := range c.Instances {
		if c.Instances[i].Default {
			return &c.Instances[i], nil
		}
	}
	return nil, errors.New("no default instance configured")
}

// SetDefault sets the default instance by name
func (c *Config) SetDefault(name string) error {
	found := false
	for i := range c.Instances {
		if c.Instances[i].Name == name {
			c.Instances[i].Default = true
			found = true
		} else {
			c.Instances[i].Default = false
		}
	}

	if !found {
		return fmt.Errorf("instance '%s' not found", name)
	}

	return nil
}

// AddInstance adds a new instance to the configuration
func (c *Config) AddInstance(instance Instance) error {
	// Validate instance
	if err := instance.Validate(); err != nil {
		return fmt.Errorf("invalid instance: %w", err)
	}

	// Check for duplicate name
	for i := range c.Instances {
		if c.Instances[i].Name == instance.Name {
			return fmt.Errorf("instance '%s' already exists", instance.Name)
		}
	}

	// If this is the first instance or marked as default, make it default
	if len(c.Instances) == 0 || instance.Default {
		// Clear other defaults
		for i := range c.Instances {
			c.Instances[i].Default = false
		}
		instance.Default = true
	}

	c.Instances = append(c.Instances, instance)
	return nil
}

// RemoveInstance removes an instance by name
func (c *Config) RemoveInstance(name string) error {
	for i := range c.Instances {
		if c.Instances[i].Name == name {
			wasDefault := c.Instances[i].Default

			// Remove instance
			c.Instances = append(c.Instances[:i], c.Instances[i+1:]...)

			// If it was default, make the first instance default
			if wasDefault && len(c.Instances) > 0 {
				c.Instances[0].Default = true
			}

			return nil
		}
	}
	return fmt.Errorf("instance '%s' not found", name)
}

// GetInstance gets an instance by name
func (c *Config) GetInstance(name string) (*Instance, error) {
	for i := range c.Instances {
		if c.Instances[i].Name == name {
			return &c.Instances[i], nil
		}
	}
	return nil, fmt.Errorf("instance '%s' not found", name)
}

// UpdateInstanceToken updates the token for an instance
func (c *Config) UpdateInstanceToken(name, token string) error {
	instance, err := c.GetInstance(name)
	if err != nil {
		return err
	}

	if token == "" {
		return errors.New("token cannot be empty")
	}

	instance.Token = token
	return nil
}

// ListInstances returns all instances
func (c *Config) ListInstances() []Instance {
	return c.Instances
}

// Validate validates the entire config
func (c *Config) Validate() error {
	if len(c.Instances) == 0 {
		return errors.New("no instances configured")
	}

	// Validate each instance
	for i, instance := range c.Instances {
		if err := instance.Validate(); err != nil {
			return fmt.Errorf("instance %d (%s) is invalid: %w", i, instance.Name, err)
		}
	}

	// Check for duplicate names
	names := make(map[string]bool)
	for _, instance := range c.Instances {
		if names[instance.Name] {
			return fmt.Errorf("duplicate instance name: %s", instance.Name)
		}
		names[instance.Name] = true
	}

	return nil
}

// Path returns the default config file path
// Linux/macOS: ~/.config/coolify/config.json
// Windows: %APPDATA%\coolify\config.json (e.g., C:\Users\username\AppData\Roaming\coolify\config.json)
func Path() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to xdg if home dir fails
		return filepath.Join(xdg.ConfigHome, "coolify", "config.json")
	}

	// Windows uses AppData/Roaming
	if filepath.Separator == '\\' {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "coolify", "config.json")
		}
		// Fallback for Windows if APPDATA not set
		return filepath.Join(homeDir, "AppData", "Roaming", "coolify", "config.json")
	}

	// Unix-like systems (Linux, macOS, BSD, etc.)
	return filepath.Join(homeDir, ".config", "coolify", "config.json")
}

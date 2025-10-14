package config

import (
	"errors"
	"fmt"
	"strings"
)

// Instance represents a Coolify instance configuration
type Instance struct {
	Name    string `json:"name"`
	FQDN    string `json:"fqdn"`
	Token   string `json:"token"`
	Default bool   `json:"default,omitempty"`
}

// Validate validates the instance configuration
func (i *Instance) Validate() error {
	if strings.TrimSpace(i.Name) == "" {
		return errors.New("instance name cannot be empty")
	}

	if strings.TrimSpace(i.FQDN) == "" {
		return errors.New("instance FQDN cannot be empty")
	}

	if !strings.HasPrefix(i.FQDN, "http://") && !strings.HasPrefix(i.FQDN, "https://") {
		return fmt.Errorf("instance FQDN must start with http:// or https://")
	}

	if strings.TrimSpace(i.Token) == "" {
		return errors.New("instance token cannot be empty")
	}

	return nil
}

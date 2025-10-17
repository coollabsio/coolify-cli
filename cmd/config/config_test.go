package config

import (
	"strings"
	"testing"

	"github.com/coollabsio/coolify-cli/internal/config"
)

func TestNewConfigCommand(t *testing.T) {
	cmd := NewConfigCommand()

	if cmd == nil {
		t.Fatal("NewConfigCommand() returned nil")
	}

	if cmd.Use != "config" {
		t.Errorf("Expected Use to be 'config', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if cmd.Run == nil {
		t.Error("Run function should not be nil")
	}
}

func TestConfigCommand_Output(t *testing.T) {
	// Test that the command returns the expected config path
	expectedPath := config.Path()

	// The path should not be empty
	if expectedPath == "" {
		t.Error("Expected config path to not be empty")
	}

	// The path should end with config.json
	if !strings.HasSuffix(expectedPath, "config.json") {
		t.Errorf("Expected path to end with 'config.json', got '%s'", expectedPath)
	}

	// The path should contain the coolify directory
	if !strings.Contains(expectedPath, "coolify") {
		t.Errorf("Expected path to contain 'coolify', got '%s'", expectedPath)
	}
}

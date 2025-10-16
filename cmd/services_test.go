package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServicesCmd_Structure(t *testing.T) {
	cmd := servicesCmd

	assert.Equal(t, "services", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	// Check that subcommands are registered
	hasListCmd := false
	hasGetCmd := false
	hasStartCmd := false
	hasStopCmd := false
	hasRestartCmd := false

	for _, subCmd := range cmd.Commands() {
		switch subCmd.Use {
		case "list":
			hasListCmd = true
		case "get <uuid>":
			hasGetCmd = true
		case "start <uuid>":
			hasStartCmd = true
		case "stop <uuid>":
			hasStopCmd = true
		case "restart <uuid>":
			hasRestartCmd = true
		}
	}

	assert.True(t, hasListCmd, "list subcommand should be registered")
	assert.True(t, hasGetCmd, "get subcommand should be registered")
	assert.True(t, hasStartCmd, "start subcommand should be registered")
	assert.True(t, hasStopCmd, "stop subcommand should be registered")
	assert.True(t, hasRestartCmd, "restart subcommand should be registered")
}

func TestServicesListCmd_Structure(t *testing.T) {
	cmd := listServicesCmd

	assert.Equal(t, "list", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.Nil(t, cmd.Args) // list takes no arguments
}

func TestServicesGetCmd_Args(t *testing.T) {
	cmd := getServiceCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"service-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestServicesGetCmd_Structure(t *testing.T) {
	cmd := getServiceCmd

	assert.Equal(t, "get <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestServicesStartCmd_Args(t *testing.T) {
	cmd := startServiceCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"service-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestServicesStartCmd_Structure(t *testing.T) {
	cmd := startServiceCmd

	assert.Equal(t, "start <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestServicesStopCmd_Args(t *testing.T) {
	cmd := stopServiceCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"service-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestServicesStopCmd_Structure(t *testing.T) {
	cmd := stopServiceCmd

	assert.Equal(t, "stop <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestServicesRestartCmd_Args(t *testing.T) {
	cmd := restartServiceCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"service-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestServicesRestartCmd_Structure(t *testing.T) {
	cmd := restartServiceCmd

	assert.Equal(t, "restart <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestServicesDeleteCmd_Args(t *testing.T) {
	cmd := deleteServiceCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"service-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestServicesDeleteCmd_Structure(t *testing.T) {
	cmd := deleteServiceCmd

	assert.Equal(t, "delete <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Verify flags exist
	assert.NotNil(t, cmd.Flags().Lookup("force"))
	assert.NotNil(t, cmd.Flags().Lookup("delete-configurations"))
	assert.NotNil(t, cmd.Flags().Lookup("delete-volumes"))
	assert.NotNil(t, cmd.Flags().Lookup("docker-cleanup"))
	assert.NotNil(t, cmd.Flags().Lookup("delete-connected-networks"))
}

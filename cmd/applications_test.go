package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplicationsListCmd_Flags(t *testing.T) {
	cmd := listApplicationsCmd

	// Verify command structure
	assert.Equal(t, "list", cmd.Use)
	assert.NotNil(t, cmd.RunE)
}

func TestApplicationsGetCmd_Args(t *testing.T) {
	cmd := getApplicationCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestApplicationsGetCmd_Flags(t *testing.T) {
	cmd := getApplicationCmd

	// Verify command structure
	assert.Equal(t, "get <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestApplicationsUpdateCmd_Args(t *testing.T) {
	cmd := updateApplicationCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestApplicationsUpdateCmd_Flags(t *testing.T) {
	cmd := updateApplicationCmd

	// Verify command structure
	assert.Equal(t, "update <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Verify key flags exist
	assert.NotNil(t, cmd.Flags().Lookup("name"))
	assert.NotNil(t, cmd.Flags().Lookup("description"))
	assert.NotNil(t, cmd.Flags().Lookup("git-branch"))
	assert.NotNil(t, cmd.Flags().Lookup("domains"))
	assert.NotNil(t, cmd.Flags().Lookup("build-command"))
	assert.NotNil(t, cmd.Flags().Lookup("start-command"))
	assert.NotNil(t, cmd.Flags().Lookup("docker-image"))
	assert.NotNil(t, cmd.Flags().Lookup("health-check-enabled"))
}

func TestApplicationsCmd_Structure(t *testing.T) {
	// Verify parent command exists
	assert.Equal(t, "applications", applicationsCmd.Use)
	assert.NotEmpty(t, applicationsCmd.Short)

	// Verify subcommands are registered
	hasListCmd := false
	hasGetCmd := false
	hasUpdateCmd := false

	for _, cmd := range applicationsCmd.Commands() {
		if cmd.Use == "list" {
			hasListCmd = true
		}
		if cmd.Use == "get <uuid>" {
			hasGetCmd = true
		}
		if cmd.Use == "update <uuid>" {
			hasUpdateCmd = true
		}
	}

	assert.True(t, hasListCmd, "list subcommand should be registered")
	assert.True(t, hasGetCmd, "get subcommand should be registered")
	assert.True(t, hasUpdateCmd, "update subcommand should be registered")
}

func TestApplicationsDeleteCmd_Args(t *testing.T) {
	cmd := deleteApplicationCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestApplicationsDeleteCmd_Flags(t *testing.T) {
	cmd := deleteApplicationCmd

	// Verify command structure
	assert.Equal(t, "delete <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Verify force flag exists
	forceFlag := cmd.Flags().Lookup("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "false", forceFlag.DefValue)
}

func TestApplicationsCmd_AllSubcommands(t *testing.T) {
	// Verify all subcommands are registered
	hasListCmd := false
	hasGetCmd := false
	hasUpdateCmd := false
	hasDeleteCmd := false

	for _, cmd := range applicationsCmd.Commands() {
		switch cmd.Use {
		case "list":
			hasListCmd = true
		case "get <uuid>":
			hasGetCmd = true
		case "update <uuid>":
			hasUpdateCmd = true
		case "delete <uuid>":
			hasDeleteCmd = true
		}
	}

	assert.True(t, hasListCmd, "list subcommand should be registered")
	assert.True(t, hasGetCmd, "get subcommand should be registered")
	assert.True(t, hasUpdateCmd, "update subcommand should be registered")
	assert.True(t, hasDeleteCmd, "delete subcommand should be registered")
}

func TestApplicationsStartCmd_Args(t *testing.T) {
	cmd := startApplicationCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestApplicationsStartCmd_Structure(t *testing.T) {
	cmd := startApplicationCmd

	assert.Equal(t, "start <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Verify aliases exist
	assert.Contains(t, cmd.Aliases, "deploy")

	// Verify flags exist
	assert.NotNil(t, cmd.Flags().Lookup("force"))
	assert.NotNil(t, cmd.Flags().Lookup("instant-deploy"))
}

func TestApplicationsStopCmd_Args(t *testing.T) {
	cmd := stopApplicationCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestApplicationsStopCmd_Structure(t *testing.T) {
	cmd := stopApplicationCmd

	assert.Equal(t, "stop <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestApplicationsRestartCmd_Args(t *testing.T) {
	cmd := restartApplicationCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestApplicationsRestartCmd_Structure(t *testing.T) {
	cmd := restartApplicationCmd

	assert.Equal(t, "restart <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestApplicationsLogsCmd_Args(t *testing.T) {
	cmd := logsApplicationCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestApplicationsLogsCmd_Structure(t *testing.T) {
	cmd := logsApplicationCmd

	assert.Equal(t, "logs <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Verify flags exist
	assert.NotNil(t, cmd.Flags().Lookup("lines"))
	assert.NotNil(t, cmd.Flags().Lookup("follow"))
}

func TestApplicationsCmd_AllLifecycleCommands(t *testing.T) {
	// Verify all lifecycle subcommands are registered
	hasStartCmd := false
	hasStopCmd := false
	hasRestartCmd := false
	hasLogsCmd := false

	for _, cmd := range applicationsCmd.Commands() {
		switch cmd.Use {
		case "start <uuid>":
			hasStartCmd = true
		case "stop <uuid>":
			hasStopCmd = true
		case "restart <uuid>":
			hasRestartCmd = true
		case "logs <uuid>":
			hasLogsCmd = true
		}
	}

	assert.True(t, hasStartCmd, "start subcommand should be registered")
	assert.True(t, hasStopCmd, "stop subcommand should be registered")
	assert.True(t, hasRestartCmd, "restart subcommand should be registered")
	assert.True(t, hasLogsCmd, "logs subcommand should be registered")
}

func TestApplicationsEnvsCmd_Structure(t *testing.T) {
	cmd := envsApplicationCmd

	assert.Equal(t, "envs", cmd.Use)
	assert.NotNil(t, cmd.Commands())
	assert.Greater(t, len(cmd.Commands()), 0, "envs should have subcommands")
}

func TestApplicationsEnvsListCmd_Args(t *testing.T) {
	cmd := listEnvsCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestApplicationsEnvsListCmd_Structure(t *testing.T) {
	cmd := listEnvsCmd

	assert.Equal(t, "list <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)
}

func TestApplicationsCmd_HasEnvsSubcommand(t *testing.T) {
	// Verify envs subcommand is registered
	hasEnvsCmd := false

	for _, cmd := range applicationsCmd.Commands() {
		if cmd.Use == "envs" {
			hasEnvsCmd = true
			break
		}
	}

	assert.True(t, hasEnvsCmd, "envs subcommand should be registered")
}

func TestApplicationsEnvsCreateCmd_Args(t *testing.T) {
	cmd := createEnvCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestApplicationsEnvsCreateCmd_Structure(t *testing.T) {
	cmd := createEnvCmd

	assert.Equal(t, "create <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Verify flags exist
	assert.NotNil(t, cmd.Flags().Lookup("key"))
	assert.NotNil(t, cmd.Flags().Lookup("value"))
	assert.NotNil(t, cmd.Flags().Lookup("build-time"))
	assert.NotNil(t, cmd.Flags().Lookup("preview"))
	assert.NotNil(t, cmd.Flags().Lookup("is-literal"))
	assert.NotNil(t, cmd.Flags().Lookup("is-multiline"))
}

func TestApplicationsEnvsUpdateCmd_Args(t *testing.T) {
	cmd := updateEnvCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 2 arguments")

	// Test with 1 argument - should fail
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.Error(t, err, "should require exactly 2 arguments")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123", "env-uuid-456"})
	assert.NoError(t, err, "should accept 2 arguments")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2", "uuid3"})
	assert.Error(t, err, "should not accept more than 2 arguments")
}

func TestApplicationsEnvsUpdateCmd_Structure(t *testing.T) {
	cmd := updateEnvCmd

	assert.Equal(t, "update <app_uuid> <env_uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Verify flags exist
	assert.NotNil(t, cmd.Flags().Lookup("key"))
	assert.NotNil(t, cmd.Flags().Lookup("value"))
	assert.NotNil(t, cmd.Flags().Lookup("build-time"))
	assert.NotNil(t, cmd.Flags().Lookup("preview"))
	assert.NotNil(t, cmd.Flags().Lookup("is-literal"))
	assert.NotNil(t, cmd.Flags().Lookup("is-multiline"))
}

func TestApplicationsEnvsDeleteCmd_Args(t *testing.T) {
	cmd := deleteEnvCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 2 arguments")

	// Test with 1 argument - should fail
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.Error(t, err, "should require exactly 2 arguments")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123", "env-uuid-456"})
	assert.NoError(t, err, "should accept 2 arguments")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2", "uuid3"})
	assert.Error(t, err, "should not accept more than 2 arguments")
}

func TestApplicationsEnvsDeleteCmd_Structure(t *testing.T) {
	cmd := deleteEnvCmd

	assert.Equal(t, "delete <app_uuid> <env_uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Verify flags exist
	assert.NotNil(t, cmd.Flags().Lookup("force"))
}

func TestApplicationsEnvsImportCmd_Args(t *testing.T) {
	cmd := importEnvCmd

	// Test with no arguments - should fail
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require exactly 1 argument")

	// Test with correct number of arguments - should pass
	err = cmd.Args(cmd, []string{"app-uuid-123"})
	assert.NoError(t, err, "should accept 1 argument")

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"uuid1", "uuid2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestApplicationsEnvsImportCmd_Structure(t *testing.T) {
	cmd := importEnvCmd

	assert.Equal(t, "import <uuid>", cmd.Use)
	assert.NotNil(t, cmd.RunE)
	assert.NotNil(t, cmd.Args)

	// Verify flags exist
	assert.NotNil(t, cmd.Flags().Lookup("file"))
	assert.NotNil(t, cmd.Flags().Lookup("build-time"))
	assert.NotNil(t, cmd.Flags().Lookup("preview"))
	assert.NotNil(t, cmd.Flags().Lookup("is-literal"))
}

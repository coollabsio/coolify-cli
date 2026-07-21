package application

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppCommand_RegistersMoveAndTagCommands(t *testing.T) {
	cmd := NewAppCommand()
	commands := map[string]*cobra.Command{}
	for _, subcommand := range cmd.Commands() {
		commands[subcommand.Name()] = subcommand
	}

	require.Contains(t, commands, "move")
	require.Contains(t, commands, "tag")
	tagCommands := map[string]*cobra.Command{}
	for _, subcommand := range commands["tag"].Commands() {
		tagCommands[subcommand.Name()] = subcommand
	}
	assert.Contains(t, tagCommands, "list")
	assert.Contains(t, tagCommands, "add")
	assert.Contains(t, tagCommands, "remove")
}

func TestNewApplicationCommands_ExposeParityFlags(t *testing.T) {
	assert.NotNil(t, NewLogsCommand().Flags().Lookup("show-timestamps"))
	assert.NotNil(t, NewLogsCommand().Flags().Lookup("service"))
	assert.NotNil(t, NewMoveCommand().Flags().Lookup("environment-uuid"))
	assert.NotNil(t, NewUpdateCommand().Flags().Lookup("compose-domain"))

	for _, flag := range []string{
		"disable-build-cache", "docker-images-to-keep", "include-source-commit-in-build",
		"inject-build-args-to-dockerfile", "is-env-sorting-enabled", "is-git-lfs-enabled",
		"is-git-shallow-clone-enabled", "is-git-submodules-enabled", "is-gzip-enabled",
		"is-pr-deployments-public-enabled", "is-preview-deployments-enabled",
		"is-raw-compose-deployment-enabled", "is-stripprefix-enabled", "stop-grace-period",
		"use-build-secrets",
	} {
		assert.NotNil(t, NewUpdateCommand().Flags().Lookup(flag), "missing --%s", flag)
	}
}

func TestUpdateCommand_RejectsDomainsWithComposeDomain(t *testing.T) {
	cmd := NewUpdateCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{
		"app-uuid",
		"--domains", "https://app.example.com",
		"--compose-domain", "app=https://app.example.com",
	})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "group [domains compose-domain]")
}

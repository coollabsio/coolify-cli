package application

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/application/env"
)

// NewAppCommand creates the app parent command
func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app",
		Aliases: []string{"apps", "application", "applications"},
		Short:   "Application related commands",
		Long:    `Manage Coolify applications - list, get, create, update, delete, and control application lifecycle.`,
	}

	// Add main subcommands
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewUpdateCommand())
	cmd.AddCommand(NewDeleteCommand())
	cmd.AddCommand(NewStartCommand())
	cmd.AddCommand(NewStopCommand())
	cmd.AddCommand(NewRestartCommand())
	cmd.AddCommand(NewLogsCommand())

	// Add env subcommand with its children
	envCmd := &cobra.Command{
		Use:     "env",
		Aliases: []string{"envs", "environment"},
		Short:   "Manage application environment variables",
		Long:    `List and manage environment variables for applications. All commands require the application UUID first to establish context.`,
	}
	envCmd.AddCommand(env.NewListEnvCommand())
	envCmd.AddCommand(env.NewGetEnvCommand())
	envCmd.AddCommand(env.NewCreateEnvCommand())
	envCmd.AddCommand(env.NewUpdateEnvCommand())
	envCmd.AddCommand(env.NewDeleteEnvCommand())
	envCmd.AddCommand(env.NewSyncEnvCommand())
	cmd.AddCommand(envCmd)

	return cmd
}

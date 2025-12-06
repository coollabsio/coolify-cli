package service

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/service/env"
)

// NewServiceCommand creates the service parent command with all subcommands
func NewServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service",
		Aliases: []string{"services", "svc"},
		Short:   "Service related commands",
		Long:    `Manage Coolify one-click services (databases, Redis, PostgreSQL, etc.).`,
	}

	// Add main service commands
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewStartCommand())
	cmd.AddCommand(NewStopCommand())
	cmd.AddCommand(NewRestartCommand())
	cmd.AddCommand(NewDeleteCommand())

	// Add env subcommand
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage service environment variables",
	}
	envCmd.AddCommand(env.NewListCommand())
	envCmd.AddCommand(env.NewGetCommand())
	envCmd.AddCommand(env.NewCreateCommand())
	envCmd.AddCommand(env.NewUpdateCommand())
	envCmd.AddCommand(env.NewDeleteCommand())
	envCmd.AddCommand(env.NewSyncCommand())
	cmd.AddCommand(envCmd)

	return cmd
}

package server

import (
	"github.com/spf13/cobra"
)

// NewServerCommand creates the server parent command
func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"servers"},
		Short:   "Server related commands",
		Long:    `Manage Coolify servers - list, get details, add new servers, validate connections, and remove servers.`,
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewAddCommand())
	cmd.AddCommand(NewRemoveCommand())
	cmd.AddCommand(NewValidateCommand())

	return cmd
}

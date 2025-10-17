package context

import (
	"github.com/spf13/cobra"
)

// NewContextCommand creates the context parent command
func NewContextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage Coolify contexts",
		Long:  `Manage Coolify contexts. A context contains the configuration (URL and token) for connecting to Coolify.`,
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewAddCommand())
	cmd.AddCommand(NewDeleteCommand())
	cmd.AddCommand(NewUseCommand())
	cmd.AddCommand(NewUpdateCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewSetTokenCommand())
	cmd.AddCommand(NewSetDefaultCommand())
	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(NewVerifyCommand())

	return cmd
}

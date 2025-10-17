package resources

import (
	"github.com/spf13/cobra"
)

// NewResourceCommand creates the resource parent command with all subcommands
func NewResourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resource",
		Aliases: []string{"resources"},
		Short:   "Resource related commands",
		Long:    `List all resources (applications, services, databases) in Coolify.`,
	}

	// Add all resource subcommands
	cmd.AddCommand(NewListCommand())

	return cmd
}

package domains

import (
	"github.com/spf13/cobra"
)

// NewDomainsCommand creates the domains parent command
func NewDomainsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "domain",
		Aliases: []string{"domains"},
		Short:   "Domain related commands",
		Long:    `List all domains configured across your Coolify resources.`,
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand())

	return cmd
}

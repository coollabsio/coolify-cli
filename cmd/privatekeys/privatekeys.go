package privatekeys

import (
	"github.com/spf13/cobra"
)

// NewDomainsCommand creates the domains parent command
func NewPrivateKeysCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "private-key",
		Aliases: []string{"private-keys", "key", "keys"},
		Short:   "Private key related commands",
		Long:    `Manage SSH private keys for server authentication - list, add, and remove keys.`,
	}

	// Add subcommands
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewDeleteCommand())

	return cmd
}

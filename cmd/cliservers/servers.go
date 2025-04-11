package cliservers

import (
	"github.com/coollabsio/cli-coolify/cmd/runtime"
	"github.com/spf13/cobra"
)

type cliServers struct {
	coolify runtime.Getter
}

func New(c runtime.Getter) *cliServers {
	return &cliServers{
		coolify: c,
	}
}

// NewCommand creates and returns the servers command
func (c *cliServers) NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "servers",
		Short: "Manage Coolify servers",
		Long: `
Manage servers in your Coolify instance.
This command allows you to list, add, remove, and manage servers.`,
	}

	// Add subcommands
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newGetCommand())
	cmd.AddCommand(c.newAddCommand())
	cmd.AddCommand(c.newRemoveCommand())
	cmd.AddCommand(c.newValidateCommand())

	return cmd
}

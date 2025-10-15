package privatekeys

import (
	"github.com/coollabsio/cli-coolify/internal/config"
	"github.com/spf13/cobra"
)

type cliPrivateKeys struct {
	coolify config.Getter
}

func New(c config.Getter) *cliPrivateKeys {
	return &cliPrivateKeys{
		coolify: c,
	}
}

func (c *cliPrivateKeys) NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "private-keys",
		Short: "Manage SSH private keys",
		Long:  `Manage SSH private keys for your Coolify instance.`,
	}

	// Add subcommands
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newGetCommand())
	cmd.AddCommand(c.newAddCommand())
	cmd.AddCommand(c.newRemoveCommand())

	return cmd
}

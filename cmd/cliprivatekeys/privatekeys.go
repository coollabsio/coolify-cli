package cliprivatekeys

import (
	"github.com/coollabsio/cli-coolify/cmd/runtime"
	"github.com/spf13/cobra"
)

// PrivateKey represents a private key in the Coolify API
type PrivateKey struct {
	ID         int    `json:"id"`
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type cliPrivateKeys struct {
	coolify runtime.Getter
}

func New(c runtime.Getter) *cliPrivateKeys {
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

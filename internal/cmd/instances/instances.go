package instances

import (
	cliinstancesset "github.com/coollabsio/cli-coolify/internal/cmd/instances/set"
	"github.com/coollabsio/cli-coolify/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cliInstances struct {
	coolify   config.Getter
	instances []config.Instance
}

func (c *cliInstances) runtime() *config.Coolify {
	return c.coolify()
}

func New(c config.Getter) *cliInstances {
	return &cliInstances{
		coolify: c,
	}
}

func (c *cliInstances) NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "instances",
		Short:   "Manage CLI instances",
		Aliases: []string{"instance"},
		Long: `
Manage CLI instances by adding, removing or setting options for the instance.
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if parent := cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
				if err := parent.PersistentPreRunE(parent, args); err != nil {
					return err
				}
			}
			if instances := viper.Get("instances"); instances != nil {
				return viper.UnmarshalKey("instances", &c.instances)
			}
			c.coolify().Logger.Info("No instances found in configuration file. Please add an instance using the 'instances add' command.")
			return nil
		},
	}

	cmd.AddCommand(c.newAddCommand())
	cmd.AddCommand(c.newRemoveCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(cliinstancesset.New(c.runtime).NewCommand())

	return cmd
}

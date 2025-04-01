package cliinstances

import (
	cliinstancesset "github.com/coollabsio/cli-coolify/cmd/cliinstances/set"
	"github.com/coollabsio/cli-coolify/cmd/coolTypes"
	"github.com/coollabsio/cli-coolify/cmd/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cliInstances struct {
	coolify   runtime.Getter
	instances []coolTypes.Instance
}

func (c *cliInstances) runtime() *runtime.Coolify {
	return c.coolify()
}

func New(c runtime.Getter) *cliInstances {
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
			if instances := viper.Get("instances"); instances != nil {
				return viper.UnmarshalKey("instances", &c.instances)
			}
			return nil
		},
	}

	cmd.AddCommand(c.newAddCommand())
	cmd.AddCommand(c.newRemoveCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(cliinstancesset.New(c.runtime).NewCommand())

	return cmd
}

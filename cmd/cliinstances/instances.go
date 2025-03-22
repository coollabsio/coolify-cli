package cliinstances

import (
	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
	"github.com/coollabsio/coolify-cli/cmd/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type runtimeGetter func() *runtime.Coolify

type cliInstances struct {
	coolify   runtimeGetter
	instances []coolTypes.Instance
}

func New(c runtimeGetter) *cliInstances {
	return &cliInstances{
		coolify: c,
	}
}

func (c *cliInstances) NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances",
		Short: "Manage CLI instances",
		Long: `
Manage CLI instances by adding, removing or altering options for the instance.
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if instances := viper.Get("instances"); instances != nil {
				return viper.UnmarshalKey("instances", &c.instances)
			}
			return nil
		},
	}

	cmd.AddCommand(c.newAddCommand())

	return cmd
}

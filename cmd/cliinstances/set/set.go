package cliinstancesset

import (
	"errors"

	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
	"github.com/coollabsio/coolify-cli/cmd/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cliInstancesSet struct {
	coolify   runtime.Getter
	instances []coolTypes.Instance
}

func New(c runtime.Getter) *cliInstancesSet {
	return &cliInstancesSet{
		coolify: c,
	}
}

// Set command modifies property on a instance. Pre and Post run functions validate all children commands and save the configuration file after the child commands sets a property.
// TLDR; children commands dont need to save the configuration file or do any validation "if instances exists".
func (c *cliInstancesSet) NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [command] [args]",
		Short: "set a property on a instance",
		Long: `
set a property on a instance from CLI configuration file.
`,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if instances := viper.Get("instances"); instances != nil {
				viper.UnmarshalKey("instances", &c.instances)
			}
			// Validate all set commands have instance name as the first argument and is found in the configuration file.
			for _, instance := range c.instances {
				if instance.Name == args[0] {
					return nil
				}
			}
			return errors.New("Instance name is not found in the configuration file.")
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Save the configuration file after setting the property.
			return c.coolify().Config.Save()
		},
	}

	cmd.AddCommand(c.newSetDefaultCommand())
	cmd.AddCommand(c.newSetTokenCommand())

	return cmd
}

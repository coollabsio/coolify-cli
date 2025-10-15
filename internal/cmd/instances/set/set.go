package cliinstancesset

import (
	"errors"

	"github.com/coollabsio/cli-coolify/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cliInstancesSet struct {
	coolify   config.Getter
	instances []config.Instance
}

func New(c config.Getter) *cliInstancesSet {
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
			if parent := cmd.Parent(); parent != nil && parent.PersistentPreRunE != nil {
				if err := parent.PersistentPreRunE(parent, args); err != nil {
					return err
				}
			}
			if instances := viper.Get("instances"); instances != nil {
				err := viper.UnmarshalKey("instances", &c.instances)
				if err != nil {
					return err
				}
			}
			// Validate all set commands have instance name as the first argument and is found in the configuration file.
			for _, instance := range c.instances {
				if instance.Name == args[0] {
					return nil
				}
			}
			return errors.New("instance name is not found in the configuration file")
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Save the configuration file after setting the property.
			return c.coolify().Save()
		},
	}

	cmd.AddCommand(c.newSetDefaultCommand())
	cmd.AddCommand(c.newSetTokenCommand())

	return cmd
}

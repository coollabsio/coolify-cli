package cliinstancesset

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *cliInstancesSet) newSetDefaultCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default [name]",
		Short: "set a instance as default",
		Long: `
set a instance as default from CLI configuration file.
`,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for i := range c.instances {
				c.instances[i].Default = c.instances[i].Name == args[0]
			}
			viper.Set("instances", c.instances)
			c.coolify().Logger.Info("Default instance set successfully to " + args[0])
		},
	}

	return cmd
}

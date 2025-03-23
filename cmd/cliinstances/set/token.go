package cliinstancesset

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *cliInstancesSet) newSetTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token [name] [token]",
		Short: "set a instance token",
		Long: `
set a instance token from CLI configuration file.
`,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			for i := range c.instances {
				if c.instances[i].Name == args[0] {
					c.instances[i].Token = args[1]
					break
				}
			}
			viper.Set("instances", c.instances)
		},
	}

	return cmd
}

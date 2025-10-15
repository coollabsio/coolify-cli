package instances

import (
	"errors"
	"slices"

	"github.com/coollabsio/cli-coolify/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *cliInstances) newRemoveCommand() *cobra.Command {
	force := false
	indexToRemove := -1
	cmd := &cobra.Command{
		Use: "remove [name]",
		Example: utils.GetCommandExample(`
%[1]s instances remove MyInstance
%[1]s instances remove localhost --force
`),
		Short: "remove a instance",
		Long: `
remove a instance from CLI configuration file.
`,
		Aliases:      []string{"delete"},
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			for i, instance := range c.instances {
				if instance.Name == args[0] {
					if !force && instance.Default {
						return errors.New("instance is set as default. Please set another instance as default before removing this instance or provide the force flag")
					}
					indexToRemove = i
					return nil
				}
			}
			return errors.New("instance name is not found in the configuration file")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			c.instances = slices.Delete(c.instances, indexToRemove, indexToRemove+1)
			viper.Set("instances", c.instances)
			return c.coolify().Save()
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&force, "force", "f", false, "Force remove instance if set as default")

	return cmd
}

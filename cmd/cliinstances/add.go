package cliinstances

import (
	"errors"
	"slices"

	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func (c *cliInstances) newAddCommand() *cobra.Command {
	force := false
	isNewDefault := false
	cmd := &cobra.Command{
		Use: "add [name] [fqdn] [token]",
		Example: `
coolify-cli instances add MyInstance https://my.instance.tld 1234
coolify-cli instances add AnotherInstance https://another.instance.tld 5678 --default
coolify-cli instances add MyInstance https://my.instance.tld 91011 --force
`,
		Short: "Add a new instance",
		Long: `
Add a new instance to the CLI configuration file.
`,
		Aliases:      []string{"create"},
		SilenceUsage: true,
		Args:         cobra.ExactArgs(3),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			for i, instance := range c.instances {
				if instance.Name == args[0] {
					if !force {
						return errors.New("Instance with the same name already exists. Use the force flag to overwrite or instances set to modify individual attributes.")
					}
					c.instances = slices.Delete(c.instances, i, i+1)
					break
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			newInstance := coolTypes.Instance{
				Name:    args[0],
				Fqdn:    args[1],
				Token:   args[2],
				Default: isNewDefault,
			}
			if isNewDefault {
				for i := range c.instances {
					c.instances[i].Default = false
				}
			}
			c.instances = append(c.instances, newInstance)
			viper.Set("instances", c.instances)
			return c.coolify().Config.Save()
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&force, "force", "f", false, "Force overwrite existing instance with the same name")
	flags.BoolVarP(&isNewDefault, "default", "d", false, "Set this instance as the default instance")

	return cmd
}

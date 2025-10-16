package context

import (
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewSetTokenCommand creates the set-token command
func NewSetDefaultCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "set-default <name>",
		Example: `context set-default myserver`,
		Args:    cli.ExactArgs(1, "<name>"),
		Short:   "Set a context as the default",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			instances := viper.Get("instances").([]interface{})

			// Check if instance exists
			var found bool
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == name {
					found = true
					instanceMap["default"] = true
				}
			}

			if !found {
				return fmt.Errorf("%s instance is not found", name)
			} else {
				// Only unset other defaults if we found the target instance
				for _, instance := range instances {
					instanceMap := instance.(map[string]interface{})
					if instanceMap["name"] != name {
						instanceMap["default"] = false
					}
				}
			}

			viper.Set("instances", instances)
			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			// Show the list after updating
			return NewListCommand().RunE(cmd, args)
		},
	}
}

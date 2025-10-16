package context

import (
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewUseCommand creates the use command
func NewUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "use <name>",
		Example: `context use myserver`,
		Args:    cli.ExactArgs(1, "<name>"),
		Short:   "Switch to a different context (set as default)",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			instances := viper.Get("instances").([]interface{})

			// Check if instance exists
			var found bool
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == name {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("%s not found", name)
			}

			// Update default
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == name {
					instanceMap["default"] = true
				} else {
					delete(instanceMap, "default")
				}
			}

			viper.Set("instances", instances)
			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			// Show the list after switching
			return NewListCommand().RunE(cmd, args)
		},
	}
}

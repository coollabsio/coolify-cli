package context

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/coollabsio/coolify-cli/internal/cli"
)

// NewSetTokenCommand creates the set-token command
func NewSetDefaultCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "set-default <context_name>",
		Example: `context set-default myserver`,
		Args:    cli.ExactArgs(1, "<context_name>"),
		Short:   "Set a context as the default",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			raw := viper.Get("instances")
			instances, ok := raw.([]interface{})
			if !ok {
				return fmt.Errorf("invalid instances configuration")
			}

			// Check if instance exists
			var found bool
			for _, instance := range instances {
				instanceMap, ok := instance.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid instance configuration")
				}
				if val, ok := instanceMap["name"].(string); ok && val == name {
					found = true
					instanceMap["default"] = true
				}
			}

			if !found {
				return fmt.Errorf("Context '%s' not found", name)
			}

			// Only unset other defaults if we found the target instance
			for _, instance := range instances {
				instanceMap, ok := instance.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid instance configuration")
				}

				if val, ok := instanceMap["name"].(string); ok && val != name {
					instanceMap["default"] = false
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

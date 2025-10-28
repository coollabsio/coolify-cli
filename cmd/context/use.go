package context

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/coollabsio/coolify-cli/internal/cli"
)

// NewUseCommand creates the use command
func NewUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "use <context_name>",
		Example: `context use myserver`,
		Args:    cli.ExactArgs(1, "<context_name>"),
		Short:   "Switch to a different context (set as default)",
		RunE: func(_ *cobra.Command, args []string) error {
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
					break
				}
			}

			if !found {
				return fmt.Errorf("Context '%s' not found", name)
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

			fmt.Printf("Switched to context '%s'.\n", name)
			return nil
		},
	}
}

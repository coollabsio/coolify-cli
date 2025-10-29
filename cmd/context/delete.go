package context

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/coollabsio/coolify-cli/internal/cli"
)

// NewDeleteCommand creates the delete command
func NewDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <context_name>",
		Example: `context delete myserver`,
		Args:    cli.ExactArgs(1, "<context_name>"),
		Short:   "Delete a context",

		RunE: func(_ *cobra.Command, args []string) error {
			Name := args[0]
			instances := viper.Get("instances").([]interface{})
			for i, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == Name {
					instances = slices.Delete(instances, i, i+1)
					viper.Set("instances", instances)
					if err := viper.WriteConfig(); err != nil {
						return fmt.Errorf("failed to write config: %w", err)
					}

					if instanceMap["default"] == true {
						if len(instances) > 0 {
							instances[0].(map[string]interface{})["default"] = true
							viper.Set("instances", instances)
							if err := viper.WriteConfig(); err != nil {
								return fmt.Errorf("failed to write config: %w", err)
							}
							newDefaultName := instances[0].(map[string]interface{})["name"]
							fmt.Printf("Context '%s' deleted. '%s' is now the default context.\n", Name, newDefaultName)
						} else {
							fmt.Printf("Context '%s' deleted. No contexts remaining.\n", Name)
						}
					} else {
						fmt.Printf("Context '%s' deleted.\n", Name)
					}
					return nil
				}
			}
			return fmt.Errorf("context '%s' not found", Name)
		},
	}
}

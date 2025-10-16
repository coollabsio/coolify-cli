package context

import (
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewDeleteCommand creates the delete command
func NewDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Example: `context delete myserver`,
		Args:    cli.ExactArgs(1, "<name>"),
		Short:   "Delete a context",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			instances := viper.Get("instances").([]interface{})

			for i, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == name {
					// Remove the instance
					instances = append(instances[:i], instances[i+1:]...)
					viper.Set("instances", instances)

					if err := viper.WriteConfig(); err != nil {
						return fmt.Errorf("failed to write config: %w", err)
					}

					fmt.Printf("%s removed.\n", name)

					// If it was the default, set first instance as new default
					if instanceMap["default"] == true {
						fmt.Println("Note: The default instance has been removed.")
						if len(instances) > 0 {
							instances[0].(map[string]interface{})["default"] = true
							viper.Set("instances", instances)
							if err := viper.WriteConfig(); err != nil {
								return fmt.Errorf("failed to write config: %w", err)
							}
							fmt.Printf("%s set as default.\n", instances[0].(map[string]interface{})["fqdn"])
						}
					}
					return nil
				}
			}

			return fmt.Errorf("%s not found", name)
		},
	}
}

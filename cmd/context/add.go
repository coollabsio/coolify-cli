package context

import (
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewAddCommand creates the add command
func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name> <url> <token>",
		Example: `context add myserver https://coolify.example.com your-api-token`,
		Args:    cli.ExactArgs(3, "<name> <url> <token>"),
		Short:   "Add a new context",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			host := args[1]
			token := args[2]

			force, _ := cmd.Flags().GetBool("force")
			setDefault, _ := cmd.Flags().GetBool("default")

			instances := viper.Get("instances").([]interface{})

			// Check if instance already exists
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == name {
					if force {
						instanceMap["token"] = token
						if setDefault {
							// Remove default from all instances
							for _, inst := range instances {
								instMap := inst.(map[string]interface{})
								delete(instMap, "default")
							}
							instanceMap["default"] = true
							fmt.Printf("%s already exists. Force overwriting. Setting it as default.\n", name)
						} else {
							fmt.Printf("%s already exists. Force overwriting.\n", name)
						}
						viper.Set("instances", instances)
						viper.WriteConfig()
						return nil
					}
					fmt.Printf("%s already exists.\n", name)
					fmt.Println("\nNote: Use --force to force overwrite.")
					return nil
				}
			}

			// Add new instance
			newInstance := map[string]interface{}{
				"name":  name,
				"fqdn":  host,
				"token": token,
			}

			instances = append(instances, newInstance)

			if setDefault {
				// Remove default from all instances
				for _, inst := range instances {
					instMap := inst.(map[string]interface{})
					delete(instMap, "default")
				}
				// Set new instance as default
				newInstance["default"] = true
			}

			viper.Set("instances", instances)
			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			// Show the list after adding
			return NewListCommand().RunE(cmd, args)
		},
	}

	cmd.Flags().BoolP("default", "d", false, "Set as default context")
	cmd.Flags().BoolP("force", "f", false, "Force overwrite if context already exists")

	return cmd
}

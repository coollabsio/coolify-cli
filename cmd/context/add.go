package context

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/config"
)

// NewAddCommand creates the add command
func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <context_name> <url> <token>",
		Example: `context add myserver https://coolify.example.com your-api-token`,
		Args:    cli.ExactArgs(3, "<context_name> <url> <token>"),
		Short:   "Add a new context",
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			host := args[1]
			token := args[2]

			force, _ := cmd.Flags().GetBool("force")
			setDefault, _ := cmd.Flags().GetBool("default")

			instances := viper.Get("instances").([]any)

			// Check if instance already exists
			for _, instance := range instances {
				instanceMap := instance.(map[string]any)
				if instanceMap["name"] == name {
					if force {
						instanceMap["token"] = token
						if setDefault {
							// Remove default from all instances
							for _, inst := range instances {
								instMap := inst.(map[string]any)
								instMap["default"] = false
							}
							instanceMap["default"] = true
							fmt.Printf("%s already exists. Force overwriting. Setting it as default.\n", name)
						} else {
							fmt.Printf("%s already exists. Force overwriting.\n", name)
						}
						viper.Set("instances", instances)
						if err := viper.WriteConfig(); err != nil {
							fmt.Printf("failed to write config: %v\n", err)
							return
						}
						return
					}
					fmt.Printf("%s already exists.\n", name)
					fmt.Println("\nNote: Use --force to force overwrite.")
					return
				}
			}

			// Add new instance
			newInstance := config.Instance{
				Name:    name,
				FQDN:    host,
				Token:   token,
				Default: false,
			}

			if setDefault {
				// Remove default from all instances
				for _, inst := range instances {
					instMap := inst.(map[string]any)
					instMap["default"] = false
				}
				newInstance.Default = true
				fmt.Printf("Context '%s' added and set as default.\n", newInstance.Name)
			} else {
				fmt.Printf("Context '%s' added successfully.\n", newInstance.Name)
			}

			instances = append(instances, newInstance)

			viper.Set("instances", instances)
			if err := viper.WriteConfig(); err != nil {
				fmt.Printf("failed to write config: %v\n", err)
			}
		},
	}

	cmd.Flags().BoolP("default", "d", false, "Set as default context")
	cmd.Flags().BoolP("force", "f", false, "Force overwrite if context already exists")

	return cmd
}

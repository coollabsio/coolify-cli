package context

import (
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewSetTokenCommand creates the set-token command
func NewSetTokenCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "set-token <name> <token>",
		Example: `context set-token myserver your-new-api-token`,
		Args:    cli.ExactArgs(2, "<name> <token>"),
		Short:   "Update the API token for a context",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			token := args[1]

			instances := viper.Get("instances").([]interface{})

			// Check if instance exists
			var found bool
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == name {
					found = true
					instanceMap["token"] = token
					break
				}
			}

			if !found {
				return fmt.Errorf("%s instance is not found", name)
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

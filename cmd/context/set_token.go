package context

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/coollabsio/coolify-cli/internal/cli"
)

// NewSetTokenCommand creates the set-token command
func NewSetTokenCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "set-token <context_name> <token>",
		Example: `context set-token myserver your-new-api-token`,
		Args:    cli.ExactArgs(2, "<context_name> <token>"),
		Short:   "Update the API token for a context",
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			token := args[1]
			var found interface{}
			for _, instance := range viper.Get("instances").([]interface{}) {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == name {
					found = instanceMap
					break
				}
			}
			if found == nil {
				return fmt.Errorf("context '%s' not found", name)
			}
			instances := viper.Get("instances").([]interface{})
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == name {
					instanceMap["token"] = token
				}
			}
			viper.Set("instances", instances)
			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("failed to update token for context '%s': %w", name, err)
			}
			fmt.Printf("Token updated for context '%s'.\n", name)
			return nil
		},
	}
}

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
		Use:     "set-token <context_name> <token>",
		Example: `context set-token myserver your-new-api-token`,
		Args:    cli.ExactArgs(2, "<context_name> <token>"),
		Short:   "Update the API token for a context",
		Run: func(cmd *cobra.Command, args []string) {
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
				fmt.Printf("%s instance is not found. \n", name)
				return
			}
			instances := viper.Get("instances").([]interface{})
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == name {
					instanceMap["token"] = token
				}
			}
			viper.Set("instances", instances)
			viper.WriteConfig()
			fmt.Printf("Token updated for context '%s'.\n", name)
		},
	}
}

package context

import (
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"slices"
)

// NewDeleteCommand creates the delete command
func NewDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <context_name>",
		Example: `context delete myserver`,
		Args:    cli.ExactArgs(1, "<context_name>"),
		Short:   "Delete a context",

		Run: func(cmd *cobra.Command, args []string) {
			Name := args[0]
			instances := viper.Get("instances").([]interface{})
			for i, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == Name {
					instances = slices.Delete(instances, i, i+1)
					viper.Set("instances", instances)
					viper.WriteConfig()

					if instanceMap["default"] == true {
						if len(instances) > 0 {
							instances[0].(map[string]interface{})["default"] = true
							viper.Set("instances", instances)
							viper.WriteConfig()
							newDefaultName := instances[0].(map[string]interface{})["name"]
							fmt.Printf("Context '%s' deleted. '%s' is now the default context.\n", Name, newDefaultName)
						} else {
							fmt.Printf("Context '%s' deleted. No contexts remaining.\n", Name)
						}
					} else {
						fmt.Printf("Context '%s' deleted.\n", Name)
					}
					return
				}
			}
			fmt.Printf("Context '%s' not found.\n", Name)
		},
	}
}

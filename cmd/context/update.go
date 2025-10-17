package context

import (
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewUpdateCommand creates the update command
func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <context_name>",
		Example: `context update myserver --name newname --url https://new.coolify.com --token newtoken`,
		Args:    cli.ExactArgs(1, "<context_name>"),
		Short:   "Update a context's properties (name, URL, token)",
		Run: func(cmd *cobra.Command, args []string) {
			oldName := args[0]
			instances := viper.Get("instances").([]interface{})

			// Get flags
			newName, _ := cmd.Flags().GetString("name")
			newURL, _ := cmd.Flags().GetString("url")
			newToken, _ := cmd.Flags().GetString("token")

			// Check if at least one flag is provided
			if newName == "" && newURL == "" && newToken == "" {
				fmt.Println("Error: At least one of --name, --url, or --token must be provided")
				fmt.Println("\nUsage: coolify context update <context_name> [--name <new_name>] [--url <new_url>] [--token <new_token>]")
				return
			}

			// Find the context
			var found bool
			var contextToUpdate map[string]interface{}
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == oldName {
					found = true
					contextToUpdate = instanceMap
					break
				}
			}

			if !found {
				fmt.Printf("Context '%s' not found.\n", oldName)
				return
			}

			// If renaming, check if new name already exists
			if newName != "" && newName != oldName {
				for _, instance := range instances {
					instanceMap := instance.(map[string]interface{})
					if instanceMap["name"] == newName {
						fmt.Printf("Error: Context with name '%s' already exists.\n", newName)
						return
					}
				}
				contextToUpdate["name"] = newName
			}

			// Update URL if provided
			if newURL != "" {
				contextToUpdate["fqdn"] = newURL
			}

			// Update token if provided
			if newToken != "" {
				contextToUpdate["token"] = newToken
			}

			// Save changes
			viper.Set("instances", instances)
			err := viper.WriteConfig()
			if err != nil {
				fmt.Printf("Error saving config: %v\n", err)
				return
			}

			// Use the new name if renamed, otherwise use old name
			finalName := oldName
			if newName != "" {
				finalName = newName
			}
			fmt.Printf("Context '%s' updated successfully.\n", finalName)
		},
	}

	cmd.Flags().StringP("name", "n", "", "New name for the context")
	cmd.Flags().StringP("url", "u", "", "New URL for the context")
	cmd.Flags().StringP("token", "t", "", "New token for the context")

	return cmd
}

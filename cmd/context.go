package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage Coolify contexts (instance configurations)",
	Long:  `Manage Coolify contexts. A context contains the configuration (URL and token) for a Coolify instance.`,
}

var contextVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get current context's Coolify version",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get API client
		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		// Get version using API client
		version, err := client.GetVersion(ctx)
		if err != nil {
			return fmt.Errorf("failed to get version: %w", err)
		}

		fmt.Println(version)
		return nil
	},
}
var listContextsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured contexts",
	Run: func(cmd *cobra.Command, args []string) {
		instances := viper.Get("instances").([]interface{})

		if PrettyMode {
			var prettyJSON bytes.Buffer
			instancesBytes, err := json.Marshal(instances)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = json.Indent(&prettyJSON, instancesBytes, "", "\t")
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(prettyJSON.String())
			return
		}
		if JsonMode {
			instancesBytes, err := json.Marshal(instances)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(instancesBytes))
			return
		}
		fmt.Fprintln(w, "#\tName\tFqdn\tToken\tDefault")
		for index, entry := range instances {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				fmt.Println("Error")
				return
			}
			if ShowSensitive {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", index+1, entryMap["name"], entryMap["fqdn"], entryMap["token"], map[bool]string{true: "true", false: ""}[entryMap["default"] == true])
			} else {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", index+1, entryMap["name"], entryMap["fqdn"], SensitiveInformationOverlay, map[bool]string{true: "true", false: ""}[entryMap["default"] == true])
			}

		}
		w.Flush()
		fmt.Println("\nNote: Use -s to show sensitive information.")
	},
}
var addContextCmd = &cobra.Command{
	Use:     "add <context_name> <url> <token>",
	Example: `context add myserver https://coolify.example.com your-api-token`,
	Args:    exactArgs(3, "<context_name> <url> <token>"),
	Short:   "Add a new context",
	Run: func(cmd *cobra.Command, args []string) {
		Name := args[0]
		Host := args[1]
		Token := args[2]
		force, _ := cmd.Flags().GetBool("force")
		instances := viper.Get("instances").([]interface{})
		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["name"] == Name {
				if force {
					instanceMap["token"] = Token
					if SetDefaultInstance {
						for _, instance := range instances {
							instanceMap := instance.(map[string]interface{})
							delete(instanceMap, "default")
						}
						instanceMap["default"] = true
						fmt.Printf("%s already exists. Force overwriting. Setting it as default. \n", Name)
					} else {
						fmt.Printf("%s already exists. Force overwriting. \n", Name)
					}
					viper.Set("instances", instances)
					viper.WriteConfig()
					return
				}
				fmt.Printf("%s already exists. \n", Name)
				fmt.Println("\nNote: Use --force to force overwrite.")
				return
			}
		}

		instances = append(instances, map[string]interface{}{
			"name":  Name,
			"fqdn":  Host,
			"token": Token,
		})

		if SetDefaultInstance {
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				delete(instanceMap, "default")
			}
			instances[len(instances)-1].(map[string]interface{})["default"] = true
			fmt.Printf("Context '%s' added and set as default.\n", Name)
		} else {
			fmt.Printf("Context '%s' added successfully.\n", Name)
		}
		viper.Set("instances", instances)
		viper.WriteConfig()
	},
}
var deleteContextCmd = &cobra.Command{
	Use:     "delete <context_name>",
	Example: `context delete myserver`,
	Args:    exactArgs(1, "<context_name>"),
	Short:   "Delete a context",

	Run: func(cmd *cobra.Command, args []string) {
		Name := args[0]
		instances := viper.Get("instances").([]interface{})
		for i, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["name"] == Name {
				instances = append(instances[:i], instances[i+1:]...)
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
var setTokenCmd = &cobra.Command{
	Use:     "set-token <context_name> <token>",
	Example: `context set-token myserver your-new-api-token`,
	Args:    exactArgs(2, "<context_name> <token>"),
	Short:   "Update the API token for a context",
	Run: func(cmd *cobra.Command, args []string) {
		Name = args[0]
		Token = args[1]
		var found interface{}
		for _, instance := range viper.Get("instances").([]interface{}) {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["name"] == Name {
				found = instanceMap
				break
			}
		}
		if found == nil {
			fmt.Printf("Context '%s' not found.\n", Name)
			return
		}
		instances := viper.Get("instances").([]interface{})
		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["name"] == Name {
				instanceMap["token"] = Token
			}
		}
		viper.Set("instances", instances)
		viper.WriteConfig()
		fmt.Printf("Token updated for context '%s'.\n", Name)
	},
}
var useContextCmd = &cobra.Command{
	Use:     "use <context_name>",
	Example: `context use myserver`,
	Args:    exactArgs(1, "<context_name>"),
	Short:   "Switch to a different context (set as default)",

	Run: func(cmd *cobra.Command, args []string) {
		Name := args[0]
		instances := viper.Get("instances").([]interface{})
		var found interface{}
		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["name"] == Name {
				found = instanceMap
				break
			}
		}
		if found == nil {
			fmt.Printf("Context '%s' not found.\n", Name)
			return
		}
		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["name"] == Name {
				instanceMap["default"] = true
			} else {
				delete(instanceMap, "default")
			}
		}
		viper.Set("instances", instances)
		viper.WriteConfig()
		fmt.Printf("Switched to context '%s'.\n", Name)
	},
}
var getContextCmd = &cobra.Command{
	Use:     "get <context_name>",
	Example: `context get myserver`,
	Args:    exactArgs(1, "<context_name>"),
	Short:   "Get details of a specific context",

	Run: func(cmd *cobra.Command, args []string) {
		Name := args[0]
		instances := viper.Get("instances").([]interface{})
		if PrettyMode {
			var prettyJSON bytes.Buffer
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				instanceMap["token"] = SensitiveInformationOverlay
			}
			instancesBytes, err := json.Marshal(instances)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = json.Indent(&prettyJSON, instancesBytes, "", "\t")
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(prettyJSON.String())
			return
		}
		if JsonMode {
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				instanceMap["token"] = SensitiveInformationOverlay
			}
			instancesBytes, err := json.Marshal(instances)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(instancesBytes))
			return
		}
		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["name"] == Name {
				fmt.Fprintln(w, "Name\tHost\tToken")
				if ShowSensitive {
					fmt.Fprintf(w, "%s\t%s\t%s\n", Name, instanceMap["fqdn"], instanceMap["token"])
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\n", Name, instanceMap["fqdn"], SensitiveInformationOverlay)
				}
				w.Flush()
				fmt.Println("\nNote: Use -s to show sensitive information.")
				return
			}
		}
		fmt.Printf("Context '%s' not found.\n", Name)
	},
}

var updateContextCmd = &cobra.Command{
	Use:     "update <context_name>",
	Example: `context update myserver --name newname --url https://new.coolify.com --token newtoken`,
	Args:    exactArgs(1, "<context_name>"),
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

func init() {
	addContextCmd.Flags().BoolVarP(&SetDefaultInstance, "default", "d", false, "Set as default context")
	addContextCmd.Flags().BoolP("force", "f", false, "Force overwrite if context already exists")

	updateContextCmd.Flags().StringP("name", "n", "", "New name for the context")
	updateContextCmd.Flags().StringP("url", "u", "", "New URL for the context")
	updateContextCmd.Flags().StringP("token", "t", "", "New token for the context")

	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextVersionCmd)
	contextCmd.AddCommand(listContextsCmd)
	contextCmd.AddCommand(addContextCmd)
	contextCmd.AddCommand(deleteContextCmd)
	contextCmd.AddCommand(setTokenCmd)
	contextCmd.AddCommand(updateContextCmd)
	contextCmd.AddCommand(useContextCmd)
	contextCmd.AddCommand(getContextCmd)
}

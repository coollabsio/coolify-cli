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
	Use:     "add <name> <url> <token>",
	Example: `context add myserver https://coolify.example.com your-api-token`,
	Args:    exactArgs(3, "<name> <url> <token>"),
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
		}
		viper.Set("instances", instances)
		viper.WriteConfig()
		listContextsCmd.Run(cmd, args)
	},
}
var deleteContextCmd = &cobra.Command{
	Use:     "delete <name>",
	Example: `context delete myserver`,
	Args:    exactArgs(1, "<name>"),
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
				fmt.Printf("%s removed. \n", Name)
				if instanceMap["default"] == true {
					fmt.Println("Note: The default instance has been removed.")
					if len(instances) > 0 {
						instances[0].(map[string]interface{})["default"] = true
						viper.Set("instances", instances)
						viper.WriteConfig()
						fmt.Printf("%s set as default. \n", instances[0].(map[string]interface{})["fqdn"])
					}
				}
				return
			}
		}
		fmt.Printf("%s not found. \n", Name)
	},
}
var setTokenCmd = &cobra.Command{
	Use:     "set-token <name> <token>",
	Example: `context set-token myserver your-new-api-token`,
	Args:    exactArgs(2, "<name> <token>"),
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
			fmt.Printf("%s instance is not found. \n", Name)
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
		listContextsCmd.Run(cmd, args)
	},
}
var useContextCmd = &cobra.Command{
	Use:     "use <name>",
	Example: `context use myserver`,
	Args:    exactArgs(1, "<name>"),
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
			fmt.Printf("%s not found. \n", Name)
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
		listContextsCmd.Run(cmd, args)
	},
}
var getContextCmd = &cobra.Command{
	Use:     "get <name>",
	Example: `context get myserver`,
	Args:    exactArgs(1, "<name>"),
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
		fmt.Printf("%s not found. \n", Name)
	},
}

func init() {
	addContextCmd.Flags().BoolVarP(&SetDefaultInstance, "default", "d", false, "Set as default context")
	addContextCmd.Flags().BoolP("force", "f", false, "Force overwrite if context already exists")

	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextVersionCmd)
	contextCmd.AddCommand(listContextsCmd)
	contextCmd.AddCommand(addContextCmd)
	contextCmd.AddCommand(deleteContextCmd)
	contextCmd.AddCommand(setTokenCmd)
	contextCmd.AddCommand(useContextCmd)
	contextCmd.AddCommand(getContextCmd)
}

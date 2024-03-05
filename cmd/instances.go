package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Coolify instance related commands.",
}

var instanceVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get instance version.",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := Fetch("version")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(data)
	},
}
var listInstancesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Coolify instances.",
	Run: func(cmd *cobra.Command, args []string) {
		instances := viper.Get("instances").([]interface{})

		if JsonMode {
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
			instancesBytes, err := json.Marshal(instances)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(instancesBytes))
			return
		}
		var defaultEntry map[string]interface{}
		nonDefaultEntries := make([]map[string]interface{}, 0)
		fmt.Fprintln(w, "Fqdn\tToken\tDefault")
		for _, entry := range instances {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				fmt.Println("Error")
				return
			}
			if isDefault, ok := entryMap["default"].(bool); ok && isDefault {
				defaultEntry = entryMap
			} else {
				nonDefaultEntries = append(nonDefaultEntries, entryMap)
			}
		}
		if defaultEntry != nil {
			if defaultEntry["token"] == "" {
				fmt.Fprintf(w, "%s\t%s\t%s\n", defaultEntry["fqdn"], "", "true")
			} else {
				if ShowSensitive {
					fmt.Fprintf(w, "%s\t%s\t%s\n", defaultEntry["fqdn"], defaultEntry["token"], "true")
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\n", defaultEntry["fqdn"], SensitiveInformationOverlay, "true")
				}
			}
		}
		for _, entryMap := range nonDefaultEntries {
			if entryMap["token"] == "" {
				fmt.Fprintf(w, "%s\t%s\t%s\n", entryMap["fqdn"], "", "")
			} else {
				if ShowSensitive {
					fmt.Fprintf(w, "%s\t%s\t%s\n", entryMap["fqdn"], entryMap["token"], "")
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\n", entryMap["fqdn"], SensitiveInformationOverlay, "")
				}
			}
		}
		w.Flush()
		fmt.Println("\nNote: Use -s to show sensitive information.")
	},
}
var addInstanceCmd = &cobra.Command{
	Use:     "add",
	Example: `add <host> <token>`,
	Args:    cobra.ExactArgs(2),
	Short:   "Add a Coolify instance.",

	Run: func(cmd *cobra.Command, args []string) {
		Host := args[0]
		Token := args[1]
		instances := viper.Get("instances").([]interface{})
		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["fqdn"] == Host {
				if Force {
					instanceMap["token"] = Token
					if SetDefaultInstance {
						for _, instance := range instances {
							instanceMap := instance.(map[string]interface{})
							delete(instanceMap, "default")
						}
						instanceMap["default"] = true
						fmt.Printf("%s already exists. Force overwriting. Setting it as default. \n", Host)
					} else {
						fmt.Printf("%s already exists. Force overwriting. \n", Host)
					}
					viper.Set("instances", instances)
					viper.WriteConfig()
					return
				}
				fmt.Printf("%s already exists. \n", Host)
				fmt.Println("\nNote: Use -f to force overwrite.")
				return
			}
		}

		instances = append(instances, map[string]interface{}{
			"fqdn":  Host,
			"token": Token,
		})

		if SetDefaultInstance {
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				delete(instanceMap, "default")
			}
			instances[len(instances)-1].(map[string]interface{})["default"] = true
			fmt.Printf("%s added and set as default.\n", Host)
		} else {
			fmt.Printf("%s added. \n", Host)
		}

		viper.Set("instances", instances)
		viper.WriteConfig()
	},
}
var removeInstanceCmd = &cobra.Command{
	Use:     "remove",
	Example: `remove <host>`,
	Args:    cobra.ExactArgs(1),
	Short:   "Remove a Coolify instance.",

	Run: func(cmd *cobra.Command, args []string) {
		Host := args[0]
		instances := viper.Get("instances").([]interface{})
		for i, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["fqdn"] == Host {
				instances = append(instances[:i], instances[i+1:]...)
				viper.Set("instances", instances)
				viper.WriteConfig()
				fmt.Printf("%s removed. \n", Host)
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
		fmt.Printf("%s not found. \n", Host)
	},
}
var setCmd = &cobra.Command{
	Use:   "set",
	Args:  cobra.ExactArgs(2),
	Short: "Set default instance or token.",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
var setTokenCmd = &cobra.Command{
	Use:     "token",
	Example: `set token "<token>" "<host>"`,
	Args:    cobra.ExactArgs(2),
	Short:   "Set token for the given Coolify instance.",
	Run: func(cmd *cobra.Command, args []string) {
		Token = args[0]
		Fqdn = args[1]
		var found bool
		for _, instance := range viper.Get("instances").([]interface{}) {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["fqdn"] == Fqdn {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("%s instance is not found. \n", Fqdn)
			return
		}

		instances := viper.Get("instances").([]interface{})
		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["fqdn"] == Fqdn {
				instanceMap["token"] = Token
			}
		}
		viper.Set("instances", instances)
		viper.WriteConfig()
		fmt.Printf("%s set with the given token. \n", Fqdn)
	},
}
var setDefaultCmd = &cobra.Command{
	Use:     "default",
	Example: `set default <host>`,
	Args:    cobra.ExactArgs(1),
	Short:   "Set the default Coolify instance.",

	Run: func(cmd *cobra.Command, args []string) {
		DefaultHost := args[0]
		instances := viper.Get("instances").([]interface{})
		var found bool
		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["fqdn"] == DefaultHost {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("%s not found. \n", DefaultHost)
			return
		}

		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["fqdn"] == DefaultHost {
				instanceMap["default"] = true
			} else {
				delete(instanceMap, "default")
			}
		}
		viper.Set("instances", instances)
		viper.WriteConfig()
		fmt.Printf("%s set as default. \n", DefaultHost)
	},
}
var getInstanceCmd = &cobra.Command{
	Use:     "get",
	Example: `config get <host>`,
	Args:    cobra.ExactArgs(1),
	Short:   "Get a Coolify instance.",

	Run: func(cmd *cobra.Command, args []string) {
		Host := args[0]
		instances := viper.Get("instances").([]interface{})
		if JsonMode {
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
			if instanceMap["fqdn"] == Host {
				fmt.Fprintln(w, "Host\tToken")
				if ShowSensitive {
					fmt.Fprintf(w, "%s\t%s\n", Host, instanceMap["token"])
				} else {
					fmt.Fprintf(w, "%s\t%s\n", Host, SensitiveInformationOverlay)
				}
				w.Flush()
				fmt.Println("\nNote: Use -s to show sensitive information.")
				return
			}
		}
		fmt.Printf("%s not found. \n", Host)
	},
}

func init() {
	addInstanceCmd.Flags().BoolVarP(&SetDefaultInstance, "default", "d", false, "Set default instance")

	rootCmd.AddCommand(instancesCmd)
	instancesCmd.AddCommand(instanceVersionCmd)
	instancesCmd.AddCommand(listInstancesCmd)
	instancesCmd.AddCommand(addInstanceCmd)
	instancesCmd.AddCommand(removeInstanceCmd)
	instancesCmd.AddCommand(setCmd)
	instancesCmd.AddCommand(getInstanceCmd)
	setCmd.AddCommand(setTokenCmd)
	setCmd.AddCommand(setDefaultCmd)

}
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Coolify instances and tokens.",
}

var listInstancesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Coolify instances",
	Run: func(cmd *cobra.Command, args []string) {
		instances := viper.Get("instances").([]interface{})

		var defaultEntry map[string]interface{}
		nonDefaultEntries := make([]map[string]interface{}, 0)
		fmt.Fprintln(w, "Instance\tToken\tDefault")
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

var setInstanceCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the default Coolify instance or update a token.",
}
var setTokenCmd = &cobra.Command{
	Use:     "token",
	Example: `config set token "<token>" "<host>"`,
	Args:    cobra.ExactArgs(2),
	Short:   "Set token for the given Coolify instance.",
	Run: func(cmd *cobra.Command, args []string) {
		Token = args[0]
		Fqdn = args[1]
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
	Example: `config set default <host>`,
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

var addInstanceCmd = &cobra.Command{
	Use:     "add",
	Example: `config add <host> <token>`,
	Args:    cobra.ExactArgs(2),
	Short:   "Add a Coolify instance",

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
	Example: `config remove <host>`,
	Args:    cobra.ExactArgs(1),
	Short:   "Remove a Coolify instance",

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
var getInstanceCmd = &cobra.Command{
	Use:     "get",
	Example: `config get <host>`,
	Args:    cobra.ExactArgs(1),
	Short:   "Get a Coolify instance",

	Run: func(cmd *cobra.Command, args []string) {
		Host := args[0]
		instances := viper.Get("instances").([]interface{})
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
	listInstancesCmd.Flags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")
	getInstanceCmd.Flags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")
	addInstanceCmd.Flags().BoolVarP(&Force, "force", "f", false, "Force")
	addInstanceCmd.Flags().BoolVarP(&SetDefaultInstance, "default", "d", false, "Set default instance")

	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(listInstancesCmd)
	configCmd.AddCommand(addInstanceCmd)
	configCmd.AddCommand(setInstanceCmd)
	configCmd.AddCommand(getInstanceCmd)
	configCmd.AddCommand(removeInstanceCmd)

	setInstanceCmd.AddCommand(setTokenCmd)
	setInstanceCmd.AddCommand(setDefaultCmd)
}

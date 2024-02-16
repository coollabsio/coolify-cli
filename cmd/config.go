package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure tokens and instances",
}
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured instances",
	Run: func(cmd *cobra.Command, args []string) {
		instances := viper.Get("instances")
		Json, _ := json.MarshalIndent(instances, "", " ")
		fmt.Println(string(Json))
	},
}

var defaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Get default instance",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Fqdn)
	},
}
var setCmd = &cobra.Command{
	Use: "set [default|<host>]",
	Example: `
config set default http://localhost:8000
config set http://localhost:8000 <token>`,
	Args:  cobra.ExactArgs(2),
	Short: "Update an existing or set the default instance",
	Long:  "Use 'default' as the second argument to set the default instance or",
	Run: func(cmd *cobra.Command, args []string) {
		Fqdn = args[0]
		if Fqdn[len(Fqdn)-1:] == "/" {
			Fqdn = Fqdn[:len(Fqdn)-1]
		}
		DefaultHost := ""
		if len(args) == 2 {
			if Fqdn == "default" {
				DefaultHost = args[1]
			} else {
				Token = args[1]
			}
		}

		instances := viper.Get("instances").([]interface{})

		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if Fqdn == "default" {
				if instanceMap["fqdn"] == DefaultHost {
					instanceMap["default"] = true
				} else {
					delete(instanceMap, "default")
				}
			} else {
				if instanceMap["fqdn"] == Fqdn && Token != "" {
					instanceMap["token"] = Token
				}
			}
		}
		if Fqdn != "default" {
			exists := false
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["fqdn"] == Fqdn {
					exists = true
				}
			}
			if !exists {
				instances = append(instances, map[string]interface{}{
					"fqdn":  Fqdn,
					"token": Token,
				})
			}
		}

		viper.Set("instances", instances)
		viper.WriteConfig()
		if Fqdn == "default" {
			fmt.Printf("%s set as default. \n", DefaultHost)
		} else {
			fmt.Printf("%s set with the given token. \n", Fqdn)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(listCmd)
	configCmd.AddCommand(defaultCmd)
}

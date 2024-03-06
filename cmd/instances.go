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
var addInstanceCmd = &cobra.Command{
	Use:     "add",
	Example: `add <instanceName> <fqdn> <token>`,
	Args:    cobra.ExactArgs(3),
	Short:   "Add a Coolify instance.",
	Run: func(cmd *cobra.Command, args []string) {
		Name := args[0]
		Host := args[1]
		Token := args[2]
		instances := viper.Get("instances").([]interface{})
		for _, instance := range instances {
			instanceMap := instance.(map[string]interface{})
			if instanceMap["name"] == Name {
				if Force {
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
				fmt.Println("\nNote: Use -f to force overwrite.")
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
		listInstancesCmd.Run(cmd, args)
	},
}
var removeInstanceCmd = &cobra.Command{
	Use:     "remove",
	Example: `remove <instanceName>`,
	Args:    cobra.ExactArgs(1),
	Short:   "Remove a Coolify instance.",

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
var setCmd = &cobra.Command{
	Use:   "set",
	Args:  cobra.ExactArgs(2),
	Short: "Set default instance or token.",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
var setTokenCmd = &cobra.Command{
	Use:     "token",
	Example: `set token <instanceName> "<token>"`,
	Args:    cobra.ExactArgs(2),
	Short:   "Set token for the given Coolify instance.",
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
		listInstancesCmd.Run(cmd, args)
	},
}
var setDefaultCmd = &cobra.Command{
	Use:     "default",
	Example: `set default <instanceName>`,
	Args:    cobra.ExactArgs(1),
	Short:   "Set the default Coolify instance.",

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
		listInstancesCmd.Run(cmd, args)
	},
}
var getInstanceCmd = &cobra.Command{
	Use:     "get",
	Example: `config get <instanceName>`,
	Args:    cobra.ExactArgs(1),
	Short:   "Get a Coolify instance.",

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

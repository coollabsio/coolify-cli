package cmd

import (
	"fmt"
	"os"

	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure tokens and instances.",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all instances.",
	Run: func(cmd *cobra.Command, args []string) {
		instances := viper.Get("instances").([]interface{})

		var defaultEntry map[string]interface{}
		nonDefaultEntries := make([]map[string]interface{}, 0)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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
				fmt.Fprintf(w, "%s\t%s\t%s\n", defaultEntry["fqdn"], "(empty)", "true")
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
				fmt.Fprintf(w, "%s\t%s\t%s\n", entryMap["fqdn"], "(empty)", "true")
			} else {
				if ShowSensitive {
					fmt.Fprintf(w, "%s\t%s\t%s\n", entryMap["fqdn"], entryMap["token"], "true")
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\n", entryMap["fqdn"], SensitiveInformationOverlay, "true")
				}
			}
		}
		w.Flush()
		fmt.Println("\nNote: -s to show sensitive information.")

	},
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the default instance or update a token.",
}
var setTokenCmd = &cobra.Command{
	Use:     "token",
	Example: `config set token "<token>" "<host>"`,
	Args:    cobra.ExactArgs(2),
	Short:   "Set token for the given instance.",
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
	Short:   "Set the default instance.",

	Run: func(cmd *cobra.Command, args []string) {
		DefaultHost := args[0]
		instances := viper.Get("instances").([]interface{})
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

func init() {
	listCmd.Flags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")

	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(listCmd)
	setCmd.AddCommand(setTokenCmd)
	setCmd.AddCommand(setDefaultCmd)
}

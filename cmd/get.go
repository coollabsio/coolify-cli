package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var WithResources bool

type Resource struct {
	ID     int    `json:"id"`
	Uuid   string `json:"uuid"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type Resources struct {
	Resources []Resource `json:"resources"`
}

type Server struct {
	ID       int    `json:"id"`
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	User     string `json:"user"`
	Port     int    `json:"port"`
	Settings struct {
		Reachable bool `json:"is_reachable"`
		Usable    bool `json:"is_usable"`
	} `json:"settings"`
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Query resources from the server",
}
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get instance version",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := Fetch("version")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(data)
	},
}
var serversCmd = &cobra.Command{
	Use:   "servers",
	Short: "Get all servers",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := Fetch("servers")
		if err != nil {
			fmt.Println(err)
			return
		}
		if JsonMode {
			if PrettyMode {
				var prettyJSON bytes.Buffer
				err := json.Indent(&prettyJSON, []byte(data), "", "\t")
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(string(prettyJSON.String()))
				return
			}
			fmt.Println(data)
			return
		}
		var jsondata []Server
		err = json.Unmarshal([]byte(data), &jsondata)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Fprintln(w, "Uuid\tName\tIP Address\tUser\tPort\tReachable\tUsable")
		for _, resource := range jsondata {
			if ShowSensitive {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%t\t%t\n", resource.UUID, resource.Name, resource.IP, resource.User, resource.Port, resource.Settings.Reachable, resource.Settings.Usable)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\t%t\n", resource.UUID, resource.Name, SensitiveInformationOverlay, SensitiveInformationOverlay, SensitiveInformationOverlay, resource.Settings.Reachable, resource.Settings.Usable)
			}
		}
		w.Flush()
		fmt.Println("\nNote: Use -s to show sensitive information.")
	},
}
var oneServerCmd = &cobra.Command{
	Use:   "server [uuid]",
	Args:  cobra.ExactArgs(1),
	Short: "Get server details by uuid",
	Run: func(cmd *cobra.Command, args []string) {
		uuid := args[0]
		var url = "server/" + uuid
		if WithResources {
			url = "server/" + uuid + "?resources=true"
		}

		data, err := Fetch(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		if JsonMode {
			if PrettyMode {
				var prettyJSON bytes.Buffer
				err := json.Indent(&prettyJSON, []byte(data), "", "\t")
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(string(prettyJSON.String()))
				return
			}
			fmt.Println(data)
			return
		}
		if WithResources {
			var jsondata Resources
			err = json.Unmarshal([]byte(data), &jsondata)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Fprintln(w, "Uuid\tName\tType\tStatus")
			for _, resource := range jsondata.Resources {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", resource.Uuid, resource.Name, resource.Type, resource.Status)
			}
			w.Flush()
		} else {
			var jsondata Server
			err = json.Unmarshal([]byte(data), &jsondata)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Fprintln(w, "Uuid\tName\tIP Address\tUser\tPort\tReachable\tUsable")
			if ShowSensitive {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%t\t%t\n", jsondata.UUID, jsondata.Name, jsondata.IP, jsondata.User, jsondata.Port, jsondata.Settings.Reachable, jsondata.Settings.Usable)

			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\t%t\n", jsondata.UUID, jsondata.Name, SensitiveInformationOverlay, SensitiveInformationOverlay, SensitiveInformationOverlay, jsondata.Settings.Reachable, jsondata.Settings.Usable)
			}
			w.Flush()
			fmt.Println("\nNote: Use -s to show sensitive information.")
		}

	},
}

func init() {
	serversCmd.Flags().BoolVarP(&JsonMode, "json", "", false, "Json mode")
	serversCmd.Flags().BoolVarP(&PrettyMode, "pretty", "", false, "Pretty mode")

	oneServerCmd.Flags().BoolVarP(&JsonMode, "json", "", false, "Json mode")
	oneServerCmd.Flags().BoolVarP(&PrettyMode, "pretty", "", false, "Pretty mode")
	oneServerCmd.Flags().BoolVarP(&WithResources, "resources", "", false, "With resources")

	serversCmd.Flags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")
	oneServerCmd.Flags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")

	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(versionCmd)
	getCmd.AddCommand(serversCmd)
	getCmd.AddCommand(oneServerCmd)

}

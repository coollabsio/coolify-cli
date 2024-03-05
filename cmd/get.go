package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var JsonMode bool

type Resource struct {
	ID     int    `json:"id"`
	Uuid   string `json:"uuid"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type Data struct {
	Resources []Resource `json:"resources"`
}
type Server struct {
	ID        int    `json:"id"`
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	IP        string `json:"ip"`
	User      string `json:"user"`
	Port      int    `json:"port"`
	Reachable bool   `json:"is_reachable"`
	Usable    bool   `json:"is_usable"`
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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		var jsondata []Server
		err = json.Unmarshal([]byte(data), &jsondata)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Fprintln(w, "Uuid\tName\tIP Address\tUser\tPort\tReachable\tUsable")
		for _, resource := range jsondata {
			if ShowSensitive {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%t\t%t\n", resource.UUID, resource.Name, resource.IP, resource.User, resource.Port, resource.Reachable, resource.Usable)

			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\t%t\n", resource.UUID, resource.Name, SensitiveInformationOverlay, SensitiveInformationOverlay, SensitiveInformationOverlay, resource.Reachable, resource.Usable)
			}
		}
		w.Flush()
		fmt.Println("\nNote: -s to show sensitive information.")
	},
}

func init() {
	serversCmd.Flags().BoolVarP(&JsonMode, "json", "", false, "Json mode")
	serversCmd.Flags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(versionCmd)
	getCmd.AddCommand(serversCmd)
}

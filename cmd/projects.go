package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

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

type Project struct {
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

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Project related commands",
}

var listProjectsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Run: func(cmd *cobra.Command, args []string) {
		CheckMinimumVersion("4.0.0-beta.235")
		data, err := Fetch("projects")
		if err != nil {
			log.Println(err)
			return
		}
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
		if JsonMode {
			fmt.Println(data)
			return
		}
		var jsondata []Project
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
var oneProjectCmd = &cobra.Command{
	Use:   "get [uuid]",
	Args:  cobra.ExactArgs(1),
	Short: "Get server details by uuid",
	Run: func(cmd *cobra.Command, args []string) {
		CheckMinimumVersion("4.0.0-beta.235")
		uuid := args[0]
		var url = "projects/" + uuid
		if WithResources {
			url = "projects/" + uuid + "?resources=true"
		}

		data, err := Fetch(url)
		if err != nil {
			fmt.Println(err)
			return
		}
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
		if JsonMode {
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
			var jsondata Project
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
	
	oneProjectCmd.Flags().BoolVarP(&WithResources, "resources", "", false, "With resources")
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(listProjectsCmd)
	projectsCmd.AddCommand(oneProjectCmd)
}

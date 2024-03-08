package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Resource related commands",
}

var listResourcesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all resources",
	Run: func(cmd *cobra.Command, args []string) {
		CheckMinimumVersion("4.0.0-beta.237")
		data, err := Fetch("resources")
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
		var jsondata []Resource
		err = json.Unmarshal([]byte(data), &jsondata)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintln(w, "Uuid\tName\tType\tStatus")
		for _, resource := range jsondata {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", resource.Uuid, resource.Name, resource.Type, resource.Status)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(resourcesCmd)
	resourcesCmd.AddCommand(listResourcesCmd)
}

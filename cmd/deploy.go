package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

type Deploy struct {
	Deployments []Deployment `json:"deployments"`
}

type Deployment struct {
	Message        string `json:"message"`
	ResourceUuid   string `json:"resource_uuid"`
	DeploymentUuid string `json:"deployment_uuid"`
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy related commands",
}

var deployByUuidCmd = &cobra.Command{
	Use:   "uuid <uuid>",
	Short: "Deploy by uuid",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CheckMinimumVersion("4.0.0-beta.237")
		var CsvUuids = ""
		for _, uuid := range args {
			CsvUuids += uuid + ","
		}
		CsvUuids = CsvUuids[:len(CsvUuids)-1]
		data, err := Fetch("deploy?uuid=" + CsvUuids)
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
		var jsondata Deploy
		err = json.Unmarshal([]byte(data), &jsondata)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Fprintln(w, "Message\tResource Uuid\tDeployment Uuid")
		for _, resource := range jsondata.Deployments {
			fmt.Fprintf(w, "%s\t%s\t%s\n", resource.Message, resource.ResourceUuid, resource.DeploymentUuid)
		}
		w.Flush()

	},
}

// TODO deployByTagCmd

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.AddCommand(deployByUuidCmd)
}

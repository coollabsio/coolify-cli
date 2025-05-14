package cmd

import (
	"fmt"
	"github.com/coollabsio/coolify-cli/pkg/client"
	"github.com/coollabsio/coolify-cli/pkg/config"
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
		// TODO - pull this apart once we are refactored to the client
		CheckDefaultThings(nil)
		c := client.New(config.GetBaseUrl(), config.GetToken())

		resources, err := c.ListResources()
		if err != nil {
			fmt.Println(cmd.ErrOrStderr(), err)
		}

		// Format as JSON. TODO: Is this needed or should people prefer cURL/other at this stage?
		if JsonMode || PrettyMode {
			json, err := prettyJson(resources)
			if err != nil {
				fmt.Println(cmd.ErrOrStderr(), err)
				return
			}
			fmt.Println(json)
			return
		}

		fmt.Fprintln(w, "Uuid\tName\tType\tStatus")
		for _, resource := range resources {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", resource.Uuid, resource.Name, resource.Type, resource.Status)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(resourcesCmd)
	resourcesCmd.AddCommand(listResourcesCmd)
}

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

type Domain struct {
	IP      string   `json:"ip"`
	Domains []string `json:"domains"`
}

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Domain related commands",
}

var listDomainsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all domains",
	Run: func(cmd *cobra.Command, args []string) {
		CheckMinimumVersion("4.0.0-beta.237")
		data, err := Fetch("domains")
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
		var jsondata []Domain
		err = json.Unmarshal([]byte(data), &jsondata)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Fprintln(w, "IP Address\tDomains")
		for _, resource := range jsondata {
			for _, domain := range resource.Domains {
				fmt.Fprintf(w, "%s\t%s\n", resource.IP, domain)
			}

		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(domainsCmd)
	domainsCmd.AddCommand(listDomainsCmd)

}

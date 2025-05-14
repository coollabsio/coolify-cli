package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/coollabsio/coolify-cli/pkg/client"
	"github.com/spf13/cobra"
)

var WithResources bool

type Resources struct {
	// Transitional with refactor
	Resources []client.Resource `json:"resources"`
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

var serversCmd = &cobra.Command{
	Use:   "servers",
	Short: "Server related commands",
}

var listServersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all servers",
	Run: func(cmd *cobra.Command, args []string) {
		CheckDefaultThings(nil)

		baseUrl := "servers"
		data, err := Fetch(baseUrl)
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
	Use:   "get [uuid]",
	Args:  cobra.ExactArgs(1),
	Short: "Get server details by uuid",
	Run: func(cmd *cobra.Command, args []string) {
		CheckDefaultThings(nil)
		baseUrl := "servers/"

		uuid := args[0]
		var url = baseUrl + uuid
		if WithResources {
			url = baseUrl + uuid + "?resources=true"
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

var removeServerCmd = &cobra.Command{
	Use:   "remove [uuid]",
	Short: "Remove a server",
	Run: func(cmd *cobra.Command, args []string) {
		CheckDefaultThings(nil)
		baseUrl := "servers/"
		uuid := args[0]
		response, err := Delete(baseUrl + uuid)
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := map[string]string{}
		json.Unmarshal([]byte(response), &msg)
		fmt.Println(msg["message"])
	},
}

var addServerCmd = &cobra.Command{
	Use:   "add [name] [ip] [private_key_uuid]",
	Short: "Add a server",
	Run: func(cmd *cobra.Command, args []string) {
		CheckDefaultThings(nil)
		baseUrl := "servers"
		name := args[0]
		ip := args[1]
		privateKeyUuid := args[2]
		port, _ := cmd.Flags().GetInt("port")
		user, _ := cmd.Flags().GetString("user")
		validate, _ := cmd.Flags().GetBool("validate")
		jsonData, err := json.Marshal(map[string]interface{}{
			"name":             name,
			"ip":               ip,
			"port":             port,
			"user":             user,
			"private_key_uuid": privateKeyUuid,
			"instant_validate": validate,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		response, err := Post(baseUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := map[string]string{}
		json.Unmarshal([]byte(response), &msg)
		if validate {
			fmt.Println("Server added successfully with uuid " + msg["uuid"])
		} else {
			fmt.Println("Server added successfully with uuid " + msg["uuid"] + ". Server is not validated. Use 'servers validate " + msg["uuid"] + "' to validate the server.")
		}
	},
}

var validateServerCmd = &cobra.Command{
	Use:   "validate [uuid]",
	Short: "Validate a server",
	Run: func(cmd *cobra.Command, args []string) {
		CheckDefaultThings(nil)
		baseUrl := "servers/"
		uuid := args[0]
		var url = baseUrl + uuid + "/validate"
		response, err := Fetch(url)
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := map[string]string{}
		json.Unmarshal([]byte(response), &msg)
		fmt.Println(msg["message"])
	},
}

func init() {
	oneServerCmd.Flags().BoolVarP(&WithResources, "resources", "", false, "With resources")
	rootCmd.AddCommand(serversCmd)
	serversCmd.AddCommand(listServersCmd)
	serversCmd.AddCommand(oneServerCmd)

	addServerCmd.Flags().IntP("port", "p", 22, "Port")
	addServerCmd.Flags().StringP("user", "u", "root", "User")
	addServerCmd.Flags().BoolP("validate", "", false, "Validate the server")
	serversCmd.AddCommand(addServerCmd)
	serversCmd.AddCommand(validateServerCmd)
	serversCmd.AddCommand(removeServerCmd)
}

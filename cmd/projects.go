// package cmd

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"log"

// 	"github.com/spf13/cobra"
// )

// var WithPrograms bool

// type Program struct {
// 	ID     int    `json:"id"`
// 	Uuid   string `json:"uuid"`
// 	Name   string `json:"name"`
// 	Type   string `json:"type"`
// 	Status string `json:"status"`
// }

// type Programs struct {
// 	Programs []Program `json:"programs"`
// }

// type Project struct {
// 	ID       int    `json:"id"`
// 	UUID     string `json:"uuid"`
// 	Name     string `json:"name"`
// 	IP       string `json:"ip"`
// 	User     string `json:"user"`
// 	Port     int    `json:"port"`
// 	Settings struct {
// 		Reachable bool `json:"is_reachable"`
// 		Usable    bool `json:"is_usable"`
// 	} `json:"settings"`
// }

// var projectsCmd = &cobra.Command{
// 	Use:   "projects",
// 	Short: "Project related commands",
// }

// var listProjectsCmd = &cobra.Command{
// 	Use:   "list",
// 	Short: "List all projects",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		CheckMinimumVersion("4.0.0-beta.235")
// 		data, err := Fetch("projects")
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		if PrettyMode {
// 			var prettyJSON bytes.Buffer
// 			err := json.Indent(&prettyJSON, []byte(data), "", "\t")
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 			fmt.Println(string(prettyJSON.String()))
// 			return
// 		}
// 		if JsonMode {
// 			fmt.Println(data)
// 			return
// 		}
// 		var jsondata []Project
// 		err = json.Unmarshal([]byte(data), &jsondata)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}

// 		fmt.Fprintln(w, "Uuid\tName\tIP Address\tUser\tPort\tReachable\tUsable")
// 		for _, program := range jsondata {
// 			if ShowSensitive {
// 				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%t\t%t\n", program.UUID, program.Name, program.IP, program.User, program.Port, program.Settings.Reachable, program.Settings.Usable)
// 			} else {
// 				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\t%t\n", program.UUID, program.Name, SensitiveInformationOverlay, SensitiveInformationOverlay, SensitiveInformationOverlay, program.Settings.Reachable, program.Settings.Usable)
// 			}
// 		}
// 		w.Flush()
// 		fmt.Println("\nNote: Use -s to show sensitive information.")
// 	},
// }
// var oneProjectCmd = &cobra.Command{
// 	Use:   "get [uuid]",
// 	Args:  cobra.ExactArgs(1),
// 	Short: "Get server details by uuid",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		CheckMinimumVersion("4.0.0-beta.235")
// 		uuid := args[0]
// 		var url = "projects/" + uuid
// 		if WithPrograms {
// 			url = "projects/" + uuid + "?programs=true"
// 		}

// 		data, err := Fetch(url)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		if PrettyMode {
// 			var prettyJSON bytes.Buffer
// 			err := json.Indent(&prettyJSON, []byte(data), "", "\t")
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 			fmt.Println(string(prettyJSON.String()))
// 			return
// 		}
// 		if JsonMode {
// 			fmt.Println(data)
// 			return
// 		}
// 		if WithPrograms {
// 			var jsondata Programs
// 			err = json.Unmarshal([]byte(data), &jsondata)
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 			fmt.Fprintln(w, "Uuid\tName\tType\tStatus")
// 			for _, program := range jsondata.Programs {
// 				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", program.Uuid, program.Name, program.Type, program.Status)
// 			}
// 			w.Flush()
// 		} else {
// 			var jsondata Project
// 			err = json.Unmarshal([]byte(data), &jsondata)
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 			fmt.Fprintln(w, "Uuid\tName\tIP Address\tUser\tPort\tReachable\tUsable")
// 			if ShowSensitive {
// 				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%t\t%t\n", jsondata.UUID, jsondata.Name, jsondata.IP, jsondata.User, jsondata.Port, jsondata.Settings.Reachable, jsondata.Settings.Usable)

// 			} else {
// 				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\t%t\n", jsondata.UUID, jsondata.Name, SensitiveInformationOverlay, SensitiveInformationOverlay, SensitiveInformationOverlay, jsondata.Settings.Reachable, jsondata.Settings.Usable)
// 			}
// 			w.Flush()
// 			fmt.Println("\nNote: Use -s to show sensitive information.")
// 		}

// 	},
// }

// func init() {

// 	oneProjectCmd.Flags().BoolVarP(&WithPrograms, "programs", "", false, "With programs")
// 	rootCmd.AddCommand(projectsCmd)
// 	projectsCmd.AddCommand(listProjectsCmd)
// 	projectsCmd.AddCommand(oneProjectCmd)
// }

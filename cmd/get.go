package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))
var JsonMode bool

type model struct {
	table table.Model
}

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

func (m model) Init() tea.Cmd { return nil }
func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
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
	},
}

// var serversCmd = &cobra.Command{
// 	Use:   "servers",
// 	Short: "Get all servers",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		data, err := Fetch("servers")
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		var jsondata []Server
// 		err = json.Unmarshal([]byte(data), &jsondata)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		if Raw {
// 			Json, _ := json.MarshalIndent(jsondata, "", " ")
// 			fmt.Println(string(Json))
// 		} else {
// 			var rows []table.Row
// 			for _, resource := range jsondata {
// 				rows = append(rows, table.Row{resource.UUID, resource.Name, resource.IP, resource.User, fmt.Sprint(resource.Port), fmt.Sprint(resource.Reachable), fmt.Sprint(resource.Usable)})
// 			}
// 			sort.SliceStable(rows, func(i, j int) bool {
// 				return rows[i][1] > rows[j][1]
// 			})
// 			columns := []table.Column{
// 				{Title: "Uuid", Width: 10},
// 				{Title: "Name", Width: 20},
// 				{Title: "IP", Width: 40},
// 				{Title: "User", Width: 20},
// 				{Title: "Port", Width: 20},
// 				{Title: "Reachable", Width: 10},
// 				{Title: "Usable", Width: 10},
// 			}
// 			s := table.DefaultStyles()
// 			s.Header = s.Header.
// 				BorderStyle(lipgloss.DoubleBorder()).
// 				BorderForeground(lipgloss.Color("240")).
// 				BorderBottom(true).
// 				Bold(false)
// 			s.Selected = s.Selected.
// 				Foreground(lipgloss.Color("229")).
// 				Background(lipgloss.Color("57")).
// 				Bold(false)

// 			t := table.New(
// 				table.WithColumns(columns),
// 				table.WithRows(rows),
// 				table.WithFocused(true),
// 				table.WithHeight(20),
// 			)
// 			t.SetStyles(s)
// 			m := model{t}
// 			if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
// 				fmt.Println("Error running program:", err)
// 				os.Exit(1)
// 			}
// 		}
// 	},
// }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit

			// case "enter":
			// 	selectedRow := fmt.Sprint(m.table.SelectedRow()[0])
			// 	oneServerCmd.Run(oneServerCmd, []string{selectedRow})
			// 	return m, nil
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// var oneServerCmd = &cobra.Command{
// 	Use:   "server [uuid]",
// 	Args:  cobra.ExactArgs(1),
// 	Short: "Get resources by server",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		uuid := args[0]
// 		data, err := Fetch("server/" + uuid)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		var jsondata Data
// 		err = json.Unmarshal([]byte(data), &jsondata)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		if Raw {
// 			Json, _ := json.MarshalIndent(jsondata, "", " ")
// 			fmt.Println(string(Json))
// 		} else {
// 			var rows []table.Row
// 			for _, resource := range jsondata.Resources {
// 				rows = append(rows, table.Row{fmt.Sprint(resource.Uuid), resource.Name, resource.Type, resource.Status})
// 			}
// 			sort.SliceStable(rows, func(i, j int) bool {
// 				return rows[i][3] > rows[j][3]
// 			})
// 			columns := []table.Column{
// 				{Title: "Uuid", Width: 10},
// 				{Title: "Name", Width: 60},
// 				{Title: "Type", Width: 20},
// 				{Title: "Status", Width: 20},
// 			}
// 			s := table.DefaultStyles()
// 			s.Header = s.Header.
// 				BorderStyle(lipgloss.DoubleBorder()).
// 				BorderForeground(lipgloss.Color("240")).
// 				BorderBottom(true).
// 				Bold(false)
// 			s.Selected = s.Selected.
// 				Foreground(lipgloss.Color("229")).
// 				Background(lipgloss.Color("57")).
// 				Bold(false)

// 			t := table.New(
// 				table.WithColumns(columns),
// 				table.WithRows(rows),
// 				table.WithFocused(true),
// 				table.WithHeight(20),
// 			)
// 			t.SetStyles(s)
// 			m := model{t}
// 			if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
// 				fmt.Println("Error running program:", err)
// 				os.Exit(1)
// 			}

// 		}
// 	},
// }

func init() {
	serversCmd.Flags().BoolVarP(&JsonMode, "json", "", false, "Json mode")
	serversCmd.Flags().BoolVarP(&ShowSensitive, "show-sensitive", "s", false, "Show sensitive information")
	// oneServerCmd.Flags().BoolVarP(&Raw, "raw", "r", false, "Raw (json) mode")

	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(versionCmd)
	getCmd.AddCommand(serversCmd)
	// getCmd.AddCommand(oneServerCmd)

}

package cliservers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coollabsio/cli-coolify/cmd/utils"
	"github.com/coollabsio/cli-coolify/pkg/gen/openapi"
	"github.com/spf13/cobra"
)

type getModel struct {
	server        *openapi.Server
	sensitive     bool
	withResources bool
	err           error
}

func (c *cliServers) newGetCommand() *cobra.Command {
	var withResources bool

	cmd := &cobra.Command{
		Use:   "get [uuid]",
		Short: "Get server details",
		Long: `
Get detailed information about a specific server.
Optionally show its resources and sensitive information.`,
		Example: utils.GetCommandExample(`
%[1]s servers get 123e4567-e89b-12d3-a456-426614174000
%[1]s servers get 123e4567-e89b-12d3-a456-426614174000 --resources
%[1]s servers get 123e4567-e89b-12d3-a456-426614174000 --sensitive
%[1]s servers get 123e4567-e89b-12d3-a456-426614174000 --format json`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid := args[0]

			// Fetch server details
			serverData, err := c.fetchServer(cmd.Context(), uuid, withResources)
			if err != nil {
				return fmt.Errorf("failed to fetch server details: %w", err)
			}

			outFormat, err := cmd.Flags().GetString("format")
			if err != nil {
				return fmt.Errorf("failed to get output format: %w", err)
			}
			// Handle JSON output format
			if outFormat == "json" {
				return json.NewEncoder(os.Stdout).Encode(serverData)
			}

			// Create and run Bubble Tea program for interactive display
			p := tea.NewProgram(initialGetModel(serverData))
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running detail view: %w", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&withResources, "resources", false, "Show server resources")

	return cmd
}

func initialGetModel(server *openapi.Server) getModel {
	return getModel{
		server: server,
	}
}

// Implement Bubble Tea Model interface
func (m getModel) Init() tea.Cmd { return nil }

func (m getModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m getModel) View() string {
	var s strings.Builder

	// Create styles
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Bold(true).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("60"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99"))

	// Server details section
	s.WriteString(titleStyle.Render("Server Details"))
	s.WriteString("\n")

	// Helper function to add a field
	addField := func(label, value string) {
		s.WriteString(fmt.Sprintf("%s: %s\n",
			labelStyle.Render(label),
			valueStyle.Render(value)))
	}

	addField("UUID", *m.server.Uuid)
	addField("Name", *m.server.Name)

	addField("IP Address", *m.server.Ip)
	addField("User", *m.server.User)

	addField("Port", fmt.Sprintf("%d", *m.server.Port))

	status := "Offline"
	if *m.server.Settings.IsReachable && *m.server.Settings.IsUsable {
		status = "Online"
	}
	addField("Status", status)

	return "\n" + s.String()
}

func (c *cliServers) fetchServer(ctx context.Context, uuid string, withResources bool) (*openapi.Server, error) {

	req, err := c.coolify().Client.GetServerByUuid(ctx, uuid, func(ctx context.Context, req *http.Request) error {
		if withResources {
			req.URL.RawQuery = url.Values{"resources": {"true"}}.Encode()
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	parsedResponse, err := openapi.ParseGetServerByUuidResponse(req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if parsedResponse.StatusCode() != http.StatusOK {
		switch parsedResponse.StatusCode() {
		case http.StatusNotFound:
			return nil, fmt.Errorf("failed to get server: %s", *parsedResponse.JSON404.Message)
		default:
			return nil, fmt.Errorf("failed to get server: %s", string(parsedResponse.Body))
		}
	}

	return parsedResponse.JSON200, nil
}

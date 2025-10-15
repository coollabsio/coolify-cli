package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/cli-coolify/internal/client"
	"github.com/coollabsio/cli-coolify/internal/tui"
	"github.com/coollabsio/cli-coolify/internal/utils"
	"github.com/spf13/cobra"
)

type listModel struct {
	filterableTable *tui.FilterableTable
	servers         *[]client.Server
	sensitive       bool
	filter          string
	err             error
}

func (c *cliServers) newListCommand() *cobra.Command {
	var showSensitive bool
	var initialFilter string

	cmd := &cobra.Command{
		Use:   "list [filter]",
		Short: "List all servers",
		Long: `
List all servers registered in your Coolify instance.
Use --sensitive to show sensitive information like IP addresses.`,
		Example: utils.GetCommandExample(`
%[1]s servers list
%[1]s servers list "my-server"
%[1]s servers list --format json
%[1]s servers list --sensitive`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				initialFilter = args[0]
			}

			// Fetch servers from API
			data, err := c.fetchServers(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to fetch servers: %w", err)
			}

			outputFormat, err := cmd.Flags().GetString("format")
			if err != nil {
				return fmt.Errorf("failed to get output format: %w", err)
			}

			// Handle JSON output format
			if outputFormat == "json" {
				return json.NewEncoder(os.Stdout).Encode(data)
			}

			// Create and run Bubble Tea program for interactive display
			p := tea.NewProgram(initialListModel(data, showSensitive, initialFilter))
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running list view: %w", err)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&showSensitive, "sensitive", "s", false, "Show sensitive information")

	return cmd
}

func initialListModel(servers *[]client.Server, sensitive bool, initialFilter string) listModel {
	columns := []table.Column{
		{Title: "UUID", Width: 36},
		{Title: "Name", Width: 30},
		{Title: "IP Address", Width: 15},
	}

	// Convert servers to FilterableItems
	items := make([]tui.FilterableItem, len(*servers))
	for i, s := range *servers {
		items[i] = &s
	}

	// Create row builder function
	rowBuilder := func(item tui.FilterableItem) table.Row {
		s := item.(*client.Server)

		return table.Row{
			*s.Uuid,
			*s.Name,
			*s.Ip,
		}
	}

	detailBuilder := func(item tui.FilterableItem, sensitive bool) string {
		s := item.(*client.Server)

		var builder strings.Builder
		addSection := func(title, value interface{}) {
			builder.WriteString(tui.FocusedStyle.Bold(true).Render(fmt.Sprintf("%s: ", title)))
			switch v := value.(type) {
			case *string:
				builder.WriteString(*v)
			case *int:
				builder.WriteString(fmt.Sprintf("%d", *v))
			case *client.ServerProxyType:
				if v != nil {
					builder.WriteString(string(*v))
				} else {
					builder.WriteString("N/A")
				}
			case string:
				builder.WriteString(v)
			case *bool:
				if v != nil {
					builder.WriteString(fmt.Sprintf("%t", *v))
				} else {
					builder.WriteString("N/A")
				}
			}
			builder.WriteString("\n\n")
		}

		addSection("UUID", s.Uuid)
		addSection("Name", s.Name)
		addSection("IP Address", s.Ip)
		addSection("User", s.User)
		addSection("Port", s.Port)
		addSection("Proxy Type", s.ProxyType)
		addSection("Settings", "")
		addSection("  Created At", s.Settings.CreatedAt)
		addSection("  Updated At", s.Settings.UpdatedAt)
		addSection("  Server ID", s.Settings.ServerId)
		addSection("  Concurrent Builds", s.Settings.ConcurrentBuilds)
		addSection("  Dynamic Timeout", s.Settings.DynamicTimeout)
		addSection("  Docker", "")
		addSection("    Delete Unused Networks", s.Settings.DeleteUnusedNetworks)
		addSection("    Delete Unused Volumes", s.Settings.DeleteUnusedVolumes)
		addSection("    Cleanup Frequency", s.Settings.DockerCleanupFrequency)
		addSection("    Cleanup Threshold", s.Settings.DockerCleanupThreshold)
		addSection("  Force Disabled", s.Settings.ForceDisabled)
		addSection("  Force Server Cleanup", s.Settings.ForceServerCleanup)
		addSection("  Is Build Server", s.Settings.IsBuildServer)
		addSection("  Is Cloudflare Tunnel", s.Settings.IsCloudflareTunnel)
		addSection("  Is Jump Server", s.Settings.IsJumpServer)
		if s.Settings.IsLogdrainAxiomEnabled != nil && *s.Settings.IsLogdrainAxiomEnabled {
			addSection("  Axiom", "")
			addSection("    API Key", s.Settings.LogdrainAxiomApiKey)
			addSection("    Dataset Name", s.Settings.LogdrainAxiomDatasetName)
		}
		if s.Settings.IsLogdrainCustomEnabled != nil && *s.Settings.IsLogdrainCustomEnabled {
			addSection("  Custom Drain", "")
			addSection("    Config", s.Settings.LogdrainCustomConfig)
			addSection("    Config Parser", s.Settings.LogdrainCustomConfigParser)
		}
		if s.Settings.IsLogdrainHighlightEnabled != nil && *s.Settings.IsLogdrainHighlightEnabled {
			addSection("  Highlight", "")
			addSection("    Project ID", s.Settings.LogdrainHighlightProjectId)
		}
		if s.Settings.IsLogdrainNewrelicEnabled != nil && *s.Settings.IsLogdrainNewrelicEnabled {
			addSection("  Newrelic", "")
			addSection("    Base URI", s.Settings.LogdrainNewrelicBaseUri)
			addSection("    License Key", s.Settings.LogdrainNewrelicLicenseKey)
		}
		addSection("  Metrics", "")
		addSection("    History Days", s.Settings.SentinelMetricsHistoryDays)
		addSection("    Refresh Rate", s.Settings.SentinelMetricsRefreshRateSeconds)
		addSection("    Token", s.Settings.SentinelToken)

		return builder.String()
	}

	ft := tui.NewTableFilter(items, columns, rowBuilder).
		WithInitialFilter(initialFilter).
		WithDetailView(detailBuilder)

	return listModel{
		filterableTable: ft,
		servers:         servers,
		sensitive:       sensitive,
		filter:          initialFilter,
	}
}

// Implement Bubble Tea Model interface
func (m listModel) Init() tea.Cmd { return nil }

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.filterableTable.Update(msg)
}

func (m listModel) View() string {
	return m.filterableTable.View()
}

func (c *cliServers) fetchServers(ctx context.Context) (*[]client.Server, error) {
	req, err := c.coolify().Client.ListServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	parsedResponse, err := client.ParseListServersResponse(req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return parsedResponse.JSON200, nil
}

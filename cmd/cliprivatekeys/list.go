package cliprivatekeys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/cli-coolify/cmd/utils"
	"github.com/coollabsio/cli-coolify/pkg/tui"
	"github.com/spf13/cobra"
)

type filterableListModel struct {
	FilterableTable *tui.FilterableTable
}

func newFilterableListModel(keys []PrivateKey, filter string) *filterableListModel {
	columns := []table.Column{
		{Title: "UUID", Width: 30},
		{Title: "Name", Width: 30},
		{Title: "Created At", Width: 30},
	}

	return &filterableListModel{
		FilterableTable: tui.NewTableFilter(wrapKeys(keys), columns, buildRow, filter).
			WithDetailView(buildDetailView).
			WithDetailHeader("Private Key Details"),
	}
}

type wrappedKey struct {
	key PrivateKey
}

func (k wrappedKey) GetFilterValue() string {
	return k.key.Name
}

func wrapKeys(keys []PrivateKey) []tui.FilterableItem {
	items := make([]tui.FilterableItem, len(keys))
	for i, key := range keys {
		items[i] = wrappedKey{key: key}
	}
	return items
}

func buildRow(item tui.FilterableItem) table.Row {
	key := item.(wrappedKey)
	return table.Row{
		key.key.UUID,
		key.key.Name,
		key.key.CreatedAt,
	}
}

func buildDetailView(item tui.FilterableItem, sensitive bool) string {
	key := item.(wrappedKey)
	var s strings.Builder
	addSection := func(title, value string) {
		s.WriteString(tui.FocusedStyle.Bold(true).Render(title + ": "))
		s.WriteString(value + "\n\n")
	}
	addSection("UUID", key.key.UUID)
	addSection("Name", key.key.Name)
	addSection("Description", key.key.Description)

	if sensitive {
		addSection("Private Key", "\n"+key.key.PrivateKey)
		addSection("Public Key", "\n"+key.key.PublicKey)
	} else {
		addSection("Private Key", "********")
		addSection("Public Key", "********")
	}

	addSection("Git Related", fmt.Sprintf("%v", key.key.IsGitRelated))
	addSection("Team ID", fmt.Sprintf("%d", key.key.TeamID))
	addSection("Created At", key.key.CreatedAt)
	addSection("Updated At", key.key.UpdatedAt)

	return s.String()
}

func (m *filterableListModel) Init() tea.Cmd {
	return nil
}

func (m *filterableListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.FilterableTable.Update(msg)
}

func (m *filterableListModel) View() string {
	return m.FilterableTable.View()
}

func (c *cliPrivateKeys) handleDelete(item tui.FilterableItem) error {
	key := item.(wrappedKey)
	deleteReq, err := c.coolify().NewRequest(context.Background(), http.MethodDelete, fmt.Sprintf("security/keys/%s", key.key.UUID), http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	_, err = c.coolify().DoRequest(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to delete private key: %w", err)
	}

	return nil
}

func (c *cliPrivateKeys) newListCommand() *cobra.Command {
	var filter string
	var showSensitive bool

	cmd := &cobra.Command{
		Use:   "list [filter]",
		Short: "List all private keys",
		Long:  `List all SSH private keys registered in your Coolify instance.`,
		Example: utils.GetCommandExample(`
%[1]s private-keys list --format json
%[1]s private-keys list "My Key"
%[1]s private-keys list --show-sensitive
%[1]s private-keys list # Interactive mode
`),
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				filter = args[0]
			}

			req, err := c.coolify().NewRequest(cmd.Context(), http.MethodGet, "security/keys", nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			data, err := c.coolify().DoRequest(req)
			if err != nil {
				return fmt.Errorf("failed to fetch private keys: %w", err)
			}

			// Parse the JSON data to get structured private keys
			var keys []PrivateKey
			if err := json.Unmarshal(data, &keys); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				// For JSON output, redact sensitive data if --show-sensitive is not set
				if !showSensitive {
					// Create a copy with redacted sensitive fields
					redactedKeys := make([]PrivateKey, len(keys))
					for i, key := range keys {
						redactedKeys[i] = key
						redactedKeys[i].PrivateKey = "********"
						redactedKeys[i].PublicKey = "********"
					}
					keys = redactedKeys
				}

				// For JSON output, directly encode to stdout
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(keys)
			}

			model := newFilterableListModel(keys, filter)
			model.FilterableTable.WithDeleteHandler(c.handleDelete)
			p := tea.NewProgram(model, tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}

	cmd.Flags().BoolVarP(&showSensitive, "show-sensitive", "s", false, "Show sensitive information like public keys")

	return cmd
}

package cliprivatekeys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/cli-coolify/cmd/coolTypes"
	"github.com/coollabsio/cli-coolify/cmd/utils"
	"github.com/coollabsio/cli-coolify/pkg/gen/openapi"
	"github.com/coollabsio/cli-coolify/pkg/tui"
	"github.com/spf13/cobra"
)

type filterableListModel struct {
	FilterableTable *tui.FilterableTable
}

func newFilterableListModel(keys []openapi.PrivateKey, filter string) *filterableListModel {
	columns := []table.Column{
		{Title: "UUID", Width: 30},
		{Title: "Name", Width: 30},
		{Title: "Created At", Width: 30},
	}

	return &filterableListModel{
		FilterableTable: tui.NewTableFilter(wrapKeys(keys), columns, buildRow).
			WithInitialFilter(filter).
			WithDetailView(buildDetailView).
			WithDetailHeader("Private Key Details"),
	}
}

func wrapKeys(keys []openapi.PrivateKey) []tui.FilterableItem {
	items := make([]tui.FilterableItem, len(keys))
	for i, key := range keys {
		items[i] = &key
	}
	return items
}

func buildRow(item tui.FilterableItem) table.Row {
	key := item.(*openapi.PrivateKey)
	return table.Row{
		*key.Uuid,
		*key.Name,
		*key.CreatedAt,
	}
}

func buildDetailView(item tui.FilterableItem, sensitive bool) string {
	key := item.(*openapi.PrivateKey)
	var s strings.Builder
	addSection := func(title string, value interface{}) {
		s.WriteString(tui.FocusedStyle.Bold(true).Render(title + ": "))
		if value != nil {
			switch v := value.(type) {
			case *string:
				if v != nil {
					s.WriteString(*v + "\n\n")
				}
			case *bool:
				if v != nil {
					s.WriteString(fmt.Sprintf("%v\n\n", *v))
				}
			case *int:
				if v != nil {
					s.WriteString(fmt.Sprintf("%d\n\n", *v))
				}
			}

		} else {
			s.WriteString("N/A\n\n")
		}
	}
	addSection("UUID", key.Uuid)
	addSection("Name", key.Name)
	addSection("Description", key.Description)

	if sensitive {
		addSection("Private Key", key.PrivateKey)
	} else {
		addSection("Private Key", &coolTypes.Redacted)
	}

	addSection("Git Related", key.IsGitRelated)
	addSection("Team ID", key.TeamId)
	addSection("Created At", key.CreatedAt)
	addSection("Updated At", key.UpdatedAt)

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
	key := item.(*openapi.PrivateKey)
	deleteReq, err := c.coolify().Client.DeletePrivateKeyByUuid(context.Background(), *key.Uuid)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	parsedResponse, err := openapi.ParseDeletePrivateKeyByUuidResponse(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if parsedResponse.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete private key: %s", string(parsedResponse.Body))
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

			response, err := c.coolify().Client.ListPrivateKeys(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			parsedResponse, err := openapi.ParseListPrivateKeysResponse(response)
			if err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if parsedResponse.StatusCode() != http.StatusOK {
				return fmt.Errorf("failed to fetch private keys: %s", string(parsedResponse.Body))
			}

			keys := *parsedResponse.JSON200

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				// For JSON output, redact sensitive data if --show-sensitive is not set
				if !showSensitive {
					// Create a copy with redacted sensitive fields
					redactedKeys := make([]openapi.PrivateKey, len(*parsedResponse.JSON200))
					for i, key := range *parsedResponse.JSON200 {
						redactedKeys[i] = key
						redactedKeys[i].PrivateKey = &coolTypes.Redacted
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
			p := tea.NewProgram(model)
			_, err = p.Run()
			return err
		},
	}

	cmd.Flags().BoolVarP(&showSensitive, "show-sensitive", "s", false, "Show sensitive information like public keys")

	return cmd
}

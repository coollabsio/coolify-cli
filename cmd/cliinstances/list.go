package cliinstances

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/cli-coolify/cmd/coolTypes"
	"github.com/coollabsio/cli-coolify/cmd/emoji"
	"github.com/coollabsio/cli-coolify/pkg/tui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// wrappedInstance implements the FilterableItem interface
type wrappedInstance struct {
	instance coolTypes.Instance
}

func (w wrappedInstance) GetFilterValue() string {
	return w.instance.Name
}

type filterableListModel struct {
	filterableTable *tui.FilterableTable
}

func (c *cliInstances) handleDelete(item tui.FilterableItem) error {
	instance := item.(wrappedInstance).instance

	// Don't allow deleting default instance without force flag
	if instance.Default {
		return fmt.Errorf("cannot delete default instance. Use 'instances remove %s --force' instead", instance.Name)
	}

	// Find and remove the instance from the slice
	for i, existing := range c.instances {
		if existing.Name == instance.Name {
			c.instances = append(c.instances[:i], c.instances[i+1:]...)
			break
		}
	}

	// Update viper and save
	viper.Set("instances", c.instances)
	return c.coolify().Save()
}

func newFilterableListModel(instances []coolTypes.Instance, sensitive bool, initialFilter string, deleteHandler func(tui.FilterableItem) error) *filterableListModel {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "URL", Width: 40},
		{Title: "Default", Width: 8},
	}

	// Convert instances to FilterableItems
	items := make([]tui.FilterableItem, len(instances))
	for i, instance := range instances {
		items[i] = wrappedInstance{instance: instance}
	}

	// Create row builder function
	rowBuilder := func(item tui.FilterableItem) table.Row {
		instance := item.(wrappedInstance).instance
		e := emoji.CrossMark
		if instance.Default {
			e = emoji.CheckMarkButton
		}

		return table.Row{
			instance.Name,
			instance.Fqdn,
			e,
		}
	}

	// Create detail view builder function
	detailBuilder := func(item tui.FilterableItem, sensitive bool) string {
		instance := item.(wrappedInstance).instance
		var s strings.Builder

		addSection := func(title, value string) {
			s.WriteString(tui.FocusedStyle.Bold(true).Render(title + ": "))
			s.WriteString(value + "\n\n")
		}

		addSection("Name", instance.Name)
		addSection("URL", instance.Fqdn)
		if sensitive {
			addSection("Token", instance.Token)
		} else {
			addSection("Token", "********")
		}
		addSection("Default", fmt.Sprintf("%v", instance.Default))

		return s.String()
	}

	ft := tui.NewTableFilter(items, columns, rowBuilder).
		WithInitialFilter(initialFilter).
		WithDetailView(detailBuilder).
		WithDetailHeader("Instance Details").
		WithDeleteHandler(deleteHandler)

	return &filterableListModel{
		filterableTable: ft,
	}
}

func (m *filterableListModel) Init() tea.Cmd {
	return nil
}

func (m *filterableListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.filterableTable.Update(msg)
}

func (m *filterableListModel) View() string {
	return m.filterableTable.View()
}

func (c *cliInstances) newListCommand() *cobra.Command {
	sensitive := false
	cmd := &cobra.Command{
		Use:   "list [name]",
		Short: "List all instances",
		Long: `
List all instances from the CLI configuration file.
If a name is provided, only instances matching that name will be shown.
`,
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			initialFilter := ""
			if len(args) > 0 {
				initialFilter = args[0]
			}

			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return fmt.Errorf("failed to get format: %v", err)
			}
			// If format is json, output JSON and exit
			if format == "json" {
				// Filter instances for JSON output
				filteredInstances := filterInstances(c.instances, initialFilter)

				// If not sensitive, redact tokens
				if !sensitive {
					filteredInstances = redactTokens(filteredInstances)
				}

				// Encode directly to JSON using the struct's annotations
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(filteredInstances)
			}

			// Run interactive UI
			p := tea.NewProgram(newFilterableListModel(c.instances, sensitive, initialFilter, c.handleDelete))
			_, err = p.Run()
			if err != nil {
				return fmt.Errorf("program error: %v", err)
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&sensitive, "sensitive", "s", false, "Show sensitive information such as tokens")

	return cmd
}

// filterInstances filters instances based on a name filter
func filterInstances(instances []coolTypes.Instance, filter string) []coolTypes.Instance {
	if filter == "" {
		return instances
	}

	filtered := make([]coolTypes.Instance, 0)
	for _, instance := range instances {
		if strings.Contains(strings.ToLower(instance.Name), strings.ToLower(filter)) {
			filtered = append(filtered, instance)
		}
	}
	return filtered
}

// redactTokens creates a copy of instances with redacted tokens
func redactTokens(instances []coolTypes.Instance) []coolTypes.Instance {
	redacted := make([]coolTypes.Instance, len(instances))
	for i, instance := range instances {
		// Create a copy to avoid modifying original
		redacted[i] = instance
		if instance.Token != "" {
			redacted[i].Token = "********"
		}
	}
	return redacted
}

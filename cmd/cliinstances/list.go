package cliinstances

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/coolify-cli/cmd/coolTypes"
	"github.com/spf13/cobra"
)

func (c *cliInstances) newListCommand() *cobra.Command {
	sensitive := false
	format := "table"
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
			p := tea.NewProgram(newListModel(c.instances, sensitive, initialFilter))
			_, err := p.Run()
			if err != nil {
				return fmt.Errorf("program error: %v", err)
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&sensitive, "sensitive", "s", false, "Show sensitive information such as tokens")
	flags.StringVar(&format, "format", "table", "Output format (table|json)")

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

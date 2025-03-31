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
				filteredInstances := c.instances
				if initialFilter != "" {
					filteredInstances = make([]coolTypes.Instance, 0)
					for _, instance := range c.instances {
						if strings.Contains(strings.ToLower(instance.Name), strings.ToLower(initialFilter)) {
							filteredInstances = append(filteredInstances, instance)
						}
					}
				}

				output := make([]map[string]interface{}, len(filteredInstances))
				for i, instance := range filteredInstances {
					token := instance.Token
					if !sensitive && token != "" {
						token = "********"
					}
					output[i] = map[string]interface{}{
						"name":    instance.Name,
						"fqdn":    instance.Fqdn,
						"token":   token,
						"default": instance.Default,
					}
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(output)
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

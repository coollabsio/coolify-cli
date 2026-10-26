package context

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
)

// NewGetCommand creates the get command
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <context_name>",
		Example: `context get myserver`,
		Args:    cli.ExactArgs(1, "<context_name>"),
		Short:   "Get details of a specific context",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			instancesRaw := viper.Get("instances")
			if instancesRaw == nil {
				instancesRaw = []any{}
			}
			instancesInterface := instancesRaw.([]any)

			format, _ := cmd.Flags().GetString("format")

			// Use the same credential-free view model as `context list`, so
			// get output can never expose the token, even with --show-sensitive.
			var results []contextSummary
			for _, item := range instancesInterface {
				itemMap := item.(map[string]any)
				if getString(itemMap, "name") != name {
					continue
				}
				results = append(results, contextSummary{
					Name:    name,
					FQDN:    getString(itemMap, "fqdn"),
					Default: getBool(itemMap, "default"),
				})
				break
			}

			if len(results) == 0 {
				return fmt.Errorf("Context '%s' not found", name)
			}

			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(results)
		},
	}
}

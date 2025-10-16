package context

import (
	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/config"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			// Convert interface{} to config.Instance structs
			var instances []config.Instance
			for _, item := range instancesInterface {
				itemMap := item.(map[string]any)

				instance := config.Instance{
					Name:    getString(itemMap, "name"),
					FQDN:    getString(itemMap, "fqdn"),
					Token:   getString(itemMap, "token"),
					Default: getBool(itemMap, "default"),
				}
				instances = append(instances, instance)
			}

			// If a name was provided, filter to that single instance
			var results []config.Instance
			for _, inst := range instances {
				if inst.Name == name {
					results = append(results, inst)
					break
				}
			}

			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			return formatter.Format(results)
		},
	}
}

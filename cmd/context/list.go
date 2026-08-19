package context

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/coollabsio/coolify-cli/internal/output"
)

type contextSummary struct {
	Name    string `json:"name"`
	FQDN    string `json:"fqdn"`
	Default bool   `json:"default"`
}

// NewListCommand creates the list command
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured contexts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get instances from viper (returns []interface{})
			instancesRaw := viper.Get("instances")
			if instancesRaw == nil {
				instancesRaw = []interface{}{}
			}
			instancesInterface := instancesRaw.([]interface{})

			format, _ := cmd.Flags().GetString("format")

			// Use a credential-free view model so list output can never expose tokens.
			var instances []contextSummary
			for _, item := range instancesInterface {
				itemMap := item.(map[string]any)

				instance := contextSummary{
					Name:    getString(itemMap, "name"),
					FQDN:    getString(itemMap, "fqdn"),
					Default: getBool(itemMap, "default"),
				}
				instances = append(instances, instance)
			}

			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(instances)
		},
	}
}

// Helper functions to safely extract values from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

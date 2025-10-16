package context

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewGetCommand creates the get command
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <name>",
		Example: `context get myserver`,
		Args:    cli.ExactArgs(1, "<name>"),
		Short:   "Get details of a specific context",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			instances := viper.Get("instances").([]interface{})

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			// Find the instance
			var targetInstance map[string]interface{}
			for _, instance := range instances {
				instanceMap := instance.(map[string]interface{})
				if instanceMap["name"] == name {
					targetInstance = instanceMap
					break
				}
			}

			if targetInstance == nil {
				return fmt.Errorf("%s not found", name)
			}

			// Mask sensitive info if needed
			if !showSensitive {
				targetInstance["token"] = cli.SensitiveInformationOverlay
			}

			if format == "pretty" {
				var prettyJSON bytes.Buffer
				instanceBytes, err := json.Marshal(targetInstance)
				if err != nil {
					return fmt.Errorf("failed to marshal instance: %w", err)
				}
				err = json.Indent(&prettyJSON, instanceBytes, "", "\t")
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}
				fmt.Println(prettyJSON.String())
				return nil
			}

			if format == "json" {
				instanceBytes, err := json.Marshal(targetInstance)
				if err != nil {
					return fmt.Errorf("failed to marshal instance: %w", err)
				}
				fmt.Println(string(instanceBytes))
				return nil
			}

			// Table format
			fmt.Println("Name\tHost\tToken")
			fmt.Printf("%s\t%s\t%s\n", name, targetInstance["fqdn"], targetInstance["token"])

			if !showSensitive {
				fmt.Println("\nNote: Use -s to show sensitive information.")
			}

			return nil
		},
	}
}

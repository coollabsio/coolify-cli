package cliprivatekeys

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/coollabsio/cli-coolify/pkg/gen/openapi"
	"github.com/coollabsio/cli-coolify/pkg/tui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func (c *cliPrivateKeys) newGetCommand() *cobra.Command {
	var showSensitive bool

	cmd := &cobra.Command{
		Use:   "get [uuid]",
		Short: "Get private key details",
		Long:  `Get the details of a specific SSH private key by its UUID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid := args[0]

			response, err := c.coolify().Client.GetPrivateKeyByUuid(cmd.Context(), uuid)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			parsedResponse, err := openapi.ParseGetPrivateKeyByUuidResponse(response)
			if err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			if parsedResponse.StatusCode() != http.StatusOK {
				return fmt.Errorf("failed to fetch private key: %s", string(parsedResponse.Body))
			}

			format, _ := cmd.Flags().GetString("format")
			key := *parsedResponse.JSON200

			if format == "json" {
				// Redact sensitive data if --show-sensitive is not set
				if !showSensitive {
					// Create a copy with redacted sensitive fields
					redactedKey := key
					redacted := "********"
					redactedKey.PrivateKey = &redacted
					key = redactedKey
				}

				// For JSON output, directly encode to stdout
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(key)
			}

			// Setup styled table
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)

			// Set title
			titleStyle := tui.FocusedStyle.Bold(true)
			fmt.Println(titleStyle.Render(fmt.Sprintf("Private Key: %s", *key.Name)))

			// Add rows
			t.AppendRow(table.Row{"UUID", *key.Uuid})
			t.AppendRow(table.Row{"Name", *key.Name})

			// Handle sensitive info
			if showSensitive {
				t.AppendRow(table.Row{"Private Key", *key.PrivateKey})
			} else {
				sensitiveOverlay := tui.BlurredStyle.Render("(hidden - use --show-sensitive to display)")
				t.AppendRow(table.Row{"Private Key", sensitiveOverlay})
			}

			// Apply styling
			t.SetStyle(table.StyleLight)
			t.Render()

			if !showSensitive {
				fmt.Println("\nNote: Use --show-sensitive to display the actual key data")
			}

			return nil
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.BoolVarP(&showSensitive, "show-sensitive", "s", false, "Show sensitive information like key contents")

	return cmd
}

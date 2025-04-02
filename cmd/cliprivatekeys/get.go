package cliprivatekeys

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

			req, err := c.coolify().NewRequest(cmd.Context(), http.MethodGet, fmt.Sprintf("security/keys/%s", uuid), nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			data, err := c.coolify().DoRequest(req)
			if err != nil {
				return fmt.Errorf("failed to fetch private key: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			var key PrivateKey
			if err := json.Unmarshal(data, &key); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if format == "json" {
				// Redact sensitive data if --show-sensitive is not set
				if !showSensitive {
					// Create a copy with redacted sensitive fields
					redactedKey := key
					redactedKey.PrivateKey = "********"
					redactedKey.PublicKey = "********"
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

			// Define styles
			focusedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
			blurredStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("60"))

			// Set title
			titleStyle := focusedStyle.Copy().Bold(true)
			fmt.Println(titleStyle.Render(fmt.Sprintf("Private Key: %s", key.Name)))

			// Add rows
			t.AppendRow(table.Row{"UUID", key.UUID})
			t.AppendRow(table.Row{"Name", key.Name})

			// Handle sensitive info
			if showSensitive {
				t.AppendRow(table.Row{"Public Key", key.PublicKey})
				// Format private key for display
				formattedKey := strings.ReplaceAll(key.PrivateKey, "\n", "\\n")
				t.AppendRow(table.Row{"Private Key", formattedKey})
			} else {
				sensitiveOverlay := blurredStyle.Render("(hidden - use --show-sensitive to display)")
				t.AppendRow(table.Row{"Public Key", sensitiveOverlay})
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

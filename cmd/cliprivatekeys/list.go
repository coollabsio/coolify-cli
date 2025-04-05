package cliprivatekeys

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coollabsio/cli-coolify/cmd/utils"
	"github.com/spf13/cobra"
)

func (c *cliPrivateKeys) newListCommand() *cobra.Command {
	var showSensitive bool
	var initialFilter string

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
				initialFilter = args[0]
			}

			req, err := c.coolify().NewRequest(cmd.Context(), http.MethodGet, "security/keys", nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			data, err := c.coolify().DoRequest(req)
			if err != nil {
				return fmt.Errorf("failed to fetch private keys: %w", err)
			}

			// Parse the JSON data to get structured private keys
			var keys []PrivateKey
			if err := json.Unmarshal(data, &keys); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				// For JSON output, redact sensitive data if --show-sensitive is not set
				if !showSensitive {
					// Create a copy with redacted sensitive fields
					redactedKeys := make([]PrivateKey, len(keys))
					for i, key := range keys {
						redactedKeys[i] = key
						redactedKeys[i].PrivateKey = "********"
						redactedKeys[i].PublicKey = "********"
					}
					keys = redactedKeys
				}

				// For JSON output, directly encode to stdout
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(keys)
			}

			// Define the delete function to be passed to the model
			deleteFunc := func(uuid string) error {
				deleteReq, err := c.coolify().NewRequest(cmd.Context(), http.MethodDelete, fmt.Sprintf("security/keys/%s", uuid), nil)
				if err != nil {
					return fmt.Errorf("failed to create delete request: %w", err)
				}

				_, err = c.coolify().DoRequest(deleteReq)
				if err != nil {
					return fmt.Errorf("failed to delete private key: %w", err)
				}

				return nil
			}

			// Run the interactive BubbleTea model
			model := newListModel(keys, showSensitive, initialFilter, deleteFunc)
			p := tea.NewProgram(&model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running UI: %w", err)
			}

			return nil
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.BoolVarP(&showSensitive, "show-sensitive", "s", false, "Show sensitive information like public keys")

	return cmd
}

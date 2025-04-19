package cliprivatekeys

import (
	"fmt"
	"net/http"

	"github.com/coollabsio/cli-coolify/pkg/gen/openapi"
	"github.com/coollabsio/cli-coolify/pkg/tui"
	"github.com/spf13/cobra"
)

func (c *cliPrivateKeys) newRemoveCommand() *cobra.Command {
	var forceRemove bool

	cmd := &cobra.Command{
		Use:          "remove [uuid]",
		Short:        "Remove a private key",
		Long:         `Remove an private key from your Coolify instance.`,
		SilenceUsage: true,
		Aliases:      []string{"delete", "rm"},
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uuid := args[0]

			if !forceRemove {
				fmt.Printf("Are you sure you want to remove the private key with UUID '%s'? [y/N] ", uuid)
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Operation canceled")
					return nil
				}
			}

			req, err := c.coolify().Client.DeletePrivateKeyByUuid(cmd.Context(), uuid)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			parsedResponse, err := openapi.ParseDeletePrivateKeyByUuidResponse(req)
			if err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			if parsedResponse.StatusCode() != http.StatusOK {
				errorMessage := "failed to remove private key"
				switch parsedResponse.StatusCode() {
				case http.StatusBadRequest:
					errorMessage = fmt.Sprintf("%s: %s", errorMessage, *parsedResponse.JSON400.Message)
				case http.StatusUnprocessableEntity:
					errorMessage = fmt.Sprintf("%s: %s", errorMessage, *parsedResponse.JSON422.Message)
				default:
					errorMessage = fmt.Sprintf("%s: %s", errorMessage, string(parsedResponse.Body))
				}
				return fmt.Errorf("%s", errorMessage)
			}

			fmt.Println(tui.SuccessStyle.Render("Private key removed successfully"))
			return nil
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.BoolVarP(&forceRemove, "force", "f", false, "Attempt to remove without confirmation prompt")

	return cmd
}

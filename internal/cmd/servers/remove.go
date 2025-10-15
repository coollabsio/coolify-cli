package servers

import (
	"fmt"
	"net/http"

	"github.com/coollabsio/cli-coolify/internal/client"
	"github.com/coollabsio/cli-coolify/internal/tui"
	"github.com/coollabsio/cli-coolify/internal/utils"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func (c *cliServers) newRemoveCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove [uuid]",
		Short: "Remove a server",
		Long: `
Remove a server from your Coolify instance.
This action cannot be undone.`,
		Example: utils.GetCommandExample(`
%[1]s servers remove [uuid]
%[1]s servers remove [uuid] --force`),
		Aliases: []string{"delete", "rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			toRemove := args[0]

			if !force {
				fmt.Printf("Are you sure you want to remove the server with UUID '%s'? [y/N] ", toRemove)
				var confirm string
				_, err := fmt.Scanln(&confirm)
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Operation cancelled")
					return nil
				}
			}
			parsedUUID, err := uuid.Parse(toRemove)
			if err != nil {
				return fmt.Errorf("invalid UUID format: %w", err)
			}

			response, err := c.coolify().Client.DeleteServerByUuid(cmd.Context(), parsedUUID)
			if err != nil {
				return fmt.Errorf("failed to remove server: %w", err)
			}
			parsedResponse, err := client.ParseDeleteServerByUuidResponse(response)
			if err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			if parsedResponse.StatusCode() != http.StatusOK {
				switch parsedResponse.StatusCode() {
				case http.StatusNotFound:
					return fmt.Errorf("failed to remove server: %s", *parsedResponse.JSON404.Message)
				default:
					return fmt.Errorf("failed to remove server: %s", string(parsedResponse.Body))
				}
			}
			fmt.Println(tui.SuccessStyle.Render(*parsedResponse.JSON200.Message))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

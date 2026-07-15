package database

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewMoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <uuid>",
		Short: "Move a database to another environment",
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			environmentUUID, _ := cmd.Flags().GetString("environment-uuid")
			if environmentUUID == "" {
				return fmt.Errorf("--environment-uuid is required")
			}
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			response, err := service.NewDatabaseService(client).Move(cmd.Context(), args[0], environmentUUID)
			if err != nil {
				return fmt.Errorf("failed to move database: %w", err)
			}
			fmt.Println(response.Message)
			return nil
		},
	}
	cmd.Flags().String("environment-uuid", "", "Target environment UUID (required)")
	return cmd
}

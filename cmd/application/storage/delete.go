package storage

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDeleteCommand returns the storage delete command
func NewDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <app_uuid> <storage_uuid>",
		Short: "Delete a storage from an application",
		Long: `Delete a persistent volume or file storage from an application.

Examples:
  coolify app storage delete <app_uuid> <storage_uuid>`,
		Args: cli.ExactArgs(2, "<app_uuid> <storage_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			appSvc := service.NewApplicationService(client)
			if err := appSvc.DeleteStorage(ctx, args[0], args[1]); err != nil {
				return fmt.Errorf("failed to delete storage: %w", err)
			}

			fmt.Println("Storage deleted successfully.")
			return nil
		},
	}
}

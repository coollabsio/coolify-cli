package storage

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDeleteCommand returns the service storage delete command
func NewDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <service_uuid> <storage_uuid>",
		Short: "Delete a storage from a service",
		Long: `Delete a persistent volume or file storage from a service.

Examples:
  coolify svc storage delete <service_uuid> <storage_uuid>`,
		Args: cli.ExactArgs(2, "<service_uuid> <storage_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			svcSvc := service.NewService(client)
			if err := svcSvc.DeleteStorage(ctx, args[0], args[1]); err != nil {
				return fmt.Errorf("failed to delete storage: %w", err)
			}

			fmt.Println("Storage deleted successfully.")
			return nil
		},
	}
}

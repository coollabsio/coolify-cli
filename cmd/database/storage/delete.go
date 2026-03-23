package storage

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDeleteCommand returns the database storage delete command
func NewDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <db_uuid> <storage_uuid>",
		Short: "Delete a storage from a database",
		Long: `Delete a persistent volume or file storage from a database.

Examples:
  coolify db storage delete <db_uuid> <storage_uuid>`,
		Args: cli.ExactArgs(2, "<db_uuid> <storage_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := cli.CheckMinimumVersion(ctx, client, "4.0.0-beta.470"); err != nil {
				return err
			}

			dbSvc := service.NewDatabaseService(client)
			if err := dbSvc.DeleteStorage(ctx, args[0], args[1]); err != nil {
				return fmt.Errorf("failed to delete storage: %w", err)
			}

			fmt.Println("Storage deleted successfully.")
			return nil
		},
	}
}

package backup

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewTriggerCommand triggers a database backup
func NewTriggerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "trigger <database_uuid> <backup_uuid>",
		Short: "Trigger immediate backup",
		Long:  `Trigger an immediate backup for a specific backup configuration. First UUID is the database, second is the specific backup configuration to trigger.`,
		Args:  cli.ExactArgs(2, "<database_uuid> <backup_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			dbUUID := args[0]
			backupUUID := args[1]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)

			// Trigger immediate backup by updating with backup_now flag
			req := &models.DatabaseBackupUpdateRequest{
				BackupNow: cli.BoolPtr(true),
			}

			err = dbService.UpdateBackup(ctx, dbUUID, backupUUID, req)
			if err != nil {
				return fmt.Errorf("failed to trigger backup: %w", err)
			}

			fmt.Println("Immediate backup triggered successfully")
			return nil
		},
	}
}

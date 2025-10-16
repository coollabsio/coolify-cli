package backup

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewUpdateCommand updates a database
func NewUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update <database_uuid> <backup_uuid>",
		Short: "Update backup configuration",
		Long:  `Update a backup configuration settings (frequency, retention, S3, etc.). First UUID is the database, second is the specific backup configuration.`,
		Args:  cli.ExactArgs(2, "<database_uuid> <backup_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			dbUUID := args[0]
			backupUUID := args[1]

			req := &models.DatabaseBackupUpdateRequest{}
			hasChanges := false

			if cmd.Flags().Changed("enabled") {
				enabled, _ := cmd.Flags().GetBool("enabled")
				req.Enabled = &enabled
				hasChanges = true
			}
			if cmd.Flags().Changed("frequency") {
				freq, _ := cmd.Flags().GetString("frequency")
				req.Frequency = &freq
				hasChanges = true
			}
			if cmd.Flags().Changed("save-s3") {
				saveS3, _ := cmd.Flags().GetBool("save-s3")
				req.SaveS3 = &saveS3
				hasChanges = true
			}
			if cmd.Flags().Changed("s3-storage-uuid") {
				s3UUID, _ := cmd.Flags().GetString("s3-storage-uuid")
				req.S3StorageUUID = &s3UUID
				hasChanges = true
			}
			if cmd.Flags().Changed("databases-to-backup") {
				dbs, _ := cmd.Flags().GetString("databases-to-backup")
				req.DatabasesToBackup = &dbs
				hasChanges = true
			}
			if cmd.Flags().Changed("dump-all") {
				dumpAll, _ := cmd.Flags().GetBool("dump-all")
				req.DumpAll = &dumpAll
				hasChanges = true
			}

			// Retention settings
			if cmd.Flags().Changed("retention-amount-locally") {
				amount, _ := cmd.Flags().GetInt("retention-amount-locally")
				req.DatabaseBackupRetentionAmountLocally = &amount
				hasChanges = true
			}
			if cmd.Flags().Changed("retention-days-locally") {
				days, _ := cmd.Flags().GetInt("retention-days-locally")
				req.DatabaseBackupRetentionDaysLocally = &days
				hasChanges = true
			}
			if cmd.Flags().Changed("retention-max-storage-locally") {
				storage, _ := cmd.Flags().GetInt("retention-max-storage-locally")
				req.DatabaseBackupRetentionMaxStorageLocally = &storage
				hasChanges = true
			}
			if cmd.Flags().Changed("retention-amount-s3") {
				amount, _ := cmd.Flags().GetInt("retention-amount-s3")
				req.DatabaseBackupRetentionAmountS3 = &amount
				hasChanges = true
			}
			if cmd.Flags().Changed("retention-days-s3") {
				days, _ := cmd.Flags().GetInt("retention-days-s3")
				req.DatabaseBackupRetentionDaysS3 = &days
				hasChanges = true
			}
			if cmd.Flags().Changed("retention-max-storage-s3") {
				storage, _ := cmd.Flags().GetInt("retention-max-storage-s3")
				req.DatabaseBackupRetentionMaxStorageS3 = &storage
				hasChanges = true
			}

			if !hasChanges {
				return fmt.Errorf("no fields to update")
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)
			err = dbService.UpdateBackup(ctx, dbUUID, backupUUID, req)
			if err != nil {
				return fmt.Errorf("failed to update backup: %w", err)
			}

			fmt.Println("Backup configuration updated successfully")
			return nil
		},
	}
}

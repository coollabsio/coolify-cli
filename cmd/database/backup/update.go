package backup

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewUpdateCommand updates a database
func NewUpdateCommand() *cobra.Command {
	updateBackupCmd := &cobra.Command{
		Use:   "update <database_uuid> <backup_uuid>",
		Short: "Update backup configuration",
		Long:  `Update a backup configuration settings (frequency, retention, S3, etc.). First UUID is the database, second is the specific backup configuration.`,
		Args:  cli.ExactArgs(2, "<database_uuid> <backup_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			dbUUID := args[0]
			backupUUID := args[1]

			req, hasChanges := buildUpdateBackupRequest(cmd)

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

	updateBackupCmd.Flags().Bool("enabled", false, "Enable or disable backup")
	updateBackupCmd.Flags().String("frequency", "", "Backup frequency (cron expression)")
	updateBackupCmd.Flags().Bool("save-s3", false, "Save backups to S3")
	updateBackupCmd.Flags().String("s3-storage-uuid", "", "S3 storage UUID")
	updateBackupCmd.Flags().String("databases-to-backup", "", "Comma-separated list of databases to backup")
	updateBackupCmd.Flags().Bool("dump-all", false, "Dump all databases")
	updateBackupCmd.Flags().Int("retention-amount-locally", 0, "Number of backups to retain locally")
	updateBackupCmd.Flags().Int("retention-days-locally", 0, "Days to retain backups locally")
	updateBackupCmd.Flags().Float64("retention-max-storage-locally", 0, "Max storage for local backups (GB)")
	updateBackupCmd.Flags().Int("retention-amount-s3", 0, "Number of backups to retain in S3")
	updateBackupCmd.Flags().Int("retention-days-s3", 0, "Days to retain backups in S3")
	updateBackupCmd.Flags().Float64("retention-max-storage-s3", 0, "Max storage for S3 backups (GB)")

	return updateBackupCmd
}

func buildUpdateBackupRequest(cmd *cobra.Command) (*models.DatabaseBackupUpdateRequest, bool) {
	req := &models.DatabaseBackupUpdateRequest{}
	hasChanges := false
	for _, name := range []string{
		"enabled", "frequency", "save-s3", "s3-storage-uuid", "databases-to-backup", "dump-all",
		"retention-amount-locally", "retention-days-locally", "retention-max-storage-locally",
		"retention-amount-s3", "retention-days-s3", "retention-max-storage-s3",
	} {
		if cmd.Flags().Changed(name) {
			hasChanges = true
			break
		}
	}
	if cmd.Flags().Changed("enabled") {
		v, _ := cmd.Flags().GetBool("enabled")
		req.Enabled = &v
	}
	if cmd.Flags().Changed("frequency") {
		v, _ := cmd.Flags().GetString("frequency")
		req.Frequency = &v
	}
	if cmd.Flags().Changed("save-s3") {
		v, _ := cmd.Flags().GetBool("save-s3")
		req.SaveS3 = &v
	}
	if cmd.Flags().Changed("s3-storage-uuid") {
		v, _ := cmd.Flags().GetString("s3-storage-uuid")
		req.S3StorageUUID = &v
	}
	if cmd.Flags().Changed("databases-to-backup") {
		v, _ := cmd.Flags().GetString("databases-to-backup")
		req.DatabasesToBackup = &v
	}
	if cmd.Flags().Changed("dump-all") {
		v, _ := cmd.Flags().GetBool("dump-all")
		req.DumpAll = &v
	}
	if cmd.Flags().Changed("retention-amount-locally") {
		v, _ := cmd.Flags().GetInt("retention-amount-locally")
		req.DatabaseBackupRetentionAmountLocally = &v
	}
	if cmd.Flags().Changed("retention-days-locally") {
		v, _ := cmd.Flags().GetInt("retention-days-locally")
		req.DatabaseBackupRetentionDaysLocally = &v
	}
	if cmd.Flags().Changed("retention-max-storage-locally") {
		v, _ := cmd.Flags().GetFloat64("retention-max-storage-locally")
		req.DatabaseBackupRetentionMaxStorageLocally = &v
	}
	if cmd.Flags().Changed("retention-amount-s3") {
		v, _ := cmd.Flags().GetInt("retention-amount-s3")
		req.DatabaseBackupRetentionAmountS3 = &v
	}
	if cmd.Flags().Changed("retention-days-s3") {
		v, _ := cmd.Flags().GetInt("retention-days-s3")
		req.DatabaseBackupRetentionDaysS3 = &v
	}
	if cmd.Flags().Changed("retention-max-storage-s3") {
		v, _ := cmd.Flags().GetFloat64("retention-max-storage-s3")
		req.DatabaseBackupRetentionMaxStorageS3 = &v
	}
	return req, hasChanges
}

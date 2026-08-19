package backup

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewCreateCommand creates a new database
func NewCreateCommand() *cobra.Command {
	createBackupCmd := &cobra.Command{
		Use:   "create <database_uuid>",
		Short: "Create a new scheduled backup configuration",
		Long: `Create a new scheduled backup configuration for a database. Configure frequency, retention, S3 storage, and other backup options.

Example: coolify database backup create abc123 --frequency "0 0 * * *" --enabled`,
		Args: cli.ExactArgs(1, "<database_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			dbUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Check minimum version requirement
			if err := cli.CheckMinimumVersion(ctx, client, "4.0.0-beta.436"); err != nil {
				return err
			}

			req := buildCreateBackupRequest(cmd)

			dbService := service.NewDatabaseService(client)
			backup, err := dbService.CreateBackup(ctx, dbUUID, req)
			if err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(backup)
		},
	}

	createBackupCmd.Flags().String("frequency", "", "Backup frequency (cron expression, e.g., '0 0 * * *' for daily)")
	createBackupCmd.Flags().Bool("enabled", false, "Enable backup schedule")
	createBackupCmd.Flags().Bool("save-s3", false, "Save backups to S3")
	createBackupCmd.Flags().String("s3-storage-uuid", "", "S3 storage UUID")
	createBackupCmd.Flags().String("databases-to-backup", "", "Comma-separated list of databases to backup")
	createBackupCmd.Flags().Bool("dump-all", false, "Dump all databases")
	createBackupCmd.Flags().Int("retention-amount-locally", 0, "Number of backups to retain locally")
	createBackupCmd.Flags().Int("retention-days-locally", 0, "Days to retain backups locally")
	createBackupCmd.Flags().Float64("retention-max-storage-locally", 0, "Max storage for local backups (GB)")
	createBackupCmd.Flags().Int("retention-amount-s3", 0, "Number of backups to retain in S3")
	createBackupCmd.Flags().Int("retention-days-s3", 0, "Days to retain backups in S3")
	createBackupCmd.Flags().Float64("retention-max-storage-s3", 0, "Max storage for S3 backups (GB)")
	createBackupCmd.Flags().Int("timeout", 0, "Backup timeout in seconds")
	createBackupCmd.Flags().Bool("disable-local-backup", false, "Disable local backup storage")

	return createBackupCmd
}

func buildCreateBackupRequest(cmd *cobra.Command) *models.DatabaseBackupCreateRequest {
	req := &models.DatabaseBackupCreateRequest{}
	if cmd.Flags().Changed("frequency") {
		v, _ := cmd.Flags().GetString("frequency")
		req.Frequency = &v
	}
	if cmd.Flags().Changed("enabled") {
		v, _ := cmd.Flags().GetBool("enabled")
		req.Enabled = &v
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
	if cmd.Flags().Changed("timeout") {
		v, _ := cmd.Flags().GetInt("timeout")
		req.Timeout = &v
	}
	if cmd.Flags().Changed("disable-local-backup") {
		v, _ := cmd.Flags().GetBool("disable-local-backup")
		req.DisableLocalBackup = &v
	}
	return req
}

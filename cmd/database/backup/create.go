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

			req := &models.DatabaseBackupCreateRequest{}

			// Apply flags if provided
			if cmd.Flags().Changed("frequency") {
				frequency, _ := cmd.Flags().GetString("frequency")
				req.Frequency = &frequency
			}
			if cmd.Flags().Changed("enabled") {
				enabled, _ := cmd.Flags().GetBool("enabled")
				req.Enabled = &enabled
			}
			if cmd.Flags().Changed("save-s3") {
				saveS3, _ := cmd.Flags().GetBool("save-s3")
				req.SaveS3 = &saveS3
			}
			if cmd.Flags().Changed("s3-storage-uuid") {
				s3UUID, _ := cmd.Flags().GetString("s3-storage-uuid")
				req.S3StorageUUID = &s3UUID
			}
			if cmd.Flags().Changed("databases") {
				databases, _ := cmd.Flags().GetString("databases")
				req.DatabasesToBackup = &databases
			}
			if cmd.Flags().Changed("dump-all") {
				dumpAll, _ := cmd.Flags().GetBool("dump-all")
				req.DumpAll = &dumpAll
			}
			if cmd.Flags().Changed("retention-amount-locally") {
				amount, _ := cmd.Flags().GetInt("retention-amount-locally")
				req.DatabaseBackupRetentionAmountLocally = &amount
			}
			if cmd.Flags().Changed("retention-days-locally") {
				days, _ := cmd.Flags().GetInt("retention-days-locally")
				req.DatabaseBackupRetentionDaysLocally = &days
			}
			if cmd.Flags().Changed("retention-storage-locally") {
				storage, _ := cmd.Flags().GetString("retention-storage-locally")
				req.DatabaseBackupRetentionMaxStorageLocally = &storage
			}
			if cmd.Flags().Changed("retention-amount-s3") {
				amount, _ := cmd.Flags().GetInt("retention-amount-s3")
				req.DatabaseBackupRetentionAmountS3 = &amount
			}
			if cmd.Flags().Changed("retention-days-s3") {
				days, _ := cmd.Flags().GetInt("retention-days-s3")
				req.DatabaseBackupRetentionDaysS3 = &days
			}
			if cmd.Flags().Changed("retention-storage-s3") {
				storage, _ := cmd.Flags().GetString("retention-storage-s3")
				req.DatabaseBackupRetentionMaxStorageS3 = &storage
			}
			if cmd.Flags().Changed("timeout") {
				timeout, _ := cmd.Flags().GetInt("timeout")
				req.Timeout = &timeout
			}
			if cmd.Flags().Changed("disable-local") {
				disableLocal, _ := cmd.Flags().GetBool("disable-local")
				req.DisableLocalBackup = &disableLocal
			}

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
	createBackupCmd.Flags().String("retention-max-storage-locally", "", "Max storage for local backups (e.g., '1GB', '500MB')")
	createBackupCmd.Flags().Int("retention-amount-s3", 0, "Number of backups to retain in S3")
	createBackupCmd.Flags().Int("retention-days-s3", 0, "Days to retain backups in S3")
	createBackupCmd.Flags().String("retention-max-storage-s3", "", "Max storage for S3 backups (e.g., '1GB', '500MB')")
	createBackupCmd.Flags().Int("timeout", 0, "Backup timeout in seconds")
	createBackupCmd.Flags().Bool("disable-local-backup", false, "Disable local backup storage")

	return createBackupCmd
}

package backup

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewCreateCommand creates a new database
func NewCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <database_uuid>",
		Short: "Create a new scheduled backup configuration",
		Long: `Create a new scheduled backup configuration for a database. Configure frequency, retention, S3 storage, and other backup options.

Example: coolify database backup create abc123 --frequency "0 0 * * *" --enabled`,
		Args: cli.ExactArgs(1, "<database_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			dbUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
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
			if cmd.Flags().Changed("retention-amount-local") {
				amount, _ := cmd.Flags().GetInt("retention-amount-local")
				req.DatabaseBackupRetentionAmountLocally = &amount
			}
			if cmd.Flags().Changed("retention-days-local") {
				days, _ := cmd.Flags().GetInt("retention-days-local")
				req.DatabaseBackupRetentionDaysLocally = &days
			}
			if cmd.Flags().Changed("retention-storage-local") {
				storage, _ := cmd.Flags().GetString("retention-storage-local")
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
}

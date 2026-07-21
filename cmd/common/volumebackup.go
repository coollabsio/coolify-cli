package common

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/models"
)

// BindVolumeBackupScheduleFlags registers flags shared by storage backup schedule set commands.
func BindVolumeBackupScheduleFlags(cmd *cobra.Command) {
	cmd.Flags().String("frequency", "", "Cron frequency (required)")
	cmd.Flags().Bool("enabled", true, "Enable the schedule")
	cmd.Flags().Bool("save-s3", false, "Also save backups to S3")
	cmd.Flags().Bool("disable-local-backup", false, "Disable local backup retention")
	cmd.Flags().Bool("stop-during-backup", false, "Stop resource during backup")
	cmd.Flags().String("s3-storage-uuid", "", "S3 storage UUID when --save-s3 is set")
	cmd.Flags().Int("retention-amount-locally", 7, "Local retention count")
	cmd.Flags().Int("retention-days-locally", 0, "Local retention days")
	cmd.Flags().Float64("retention-max-storage-locally", 0, "Local max storage")
	cmd.Flags().Int("retention-amount-s3", 7, "S3 retention count")
	cmd.Flags().Int("retention-days-s3", 0, "S3 retention days")
	cmd.Flags().Float64("retention-max-storage-s3", 0, "S3 max storage")
	cmd.Flags().Int("timeout", 3600, "Backup timeout seconds")
}

// BuildVolumeBackupScheduleRequest builds a PUT body from flags.
func BuildVolumeBackupScheduleRequest(cmd *cobra.Command) (models.VolumeBackupScheduleRequest, error) {
	frequency, _ := cmd.Flags().GetString("frequency")
	if frequency == "" {
		return models.VolumeBackupScheduleRequest{}, fmt.Errorf("--frequency is required")
	}
	req := models.VolumeBackupScheduleRequest{Frequency: frequency}
	if cmd.Flags().Changed("enabled") {
		v, _ := cmd.Flags().GetBool("enabled")
		req.Enabled = &v
	}
	if cmd.Flags().Changed("save-s3") {
		v, _ := cmd.Flags().GetBool("save-s3")
		req.SaveS3 = &v
	}
	if cmd.Flags().Changed("disable-local-backup") {
		v, _ := cmd.Flags().GetBool("disable-local-backup")
		req.DisableLocalBackup = &v
	}
	if cmd.Flags().Changed("stop-during-backup") {
		v, _ := cmd.Flags().GetBool("stop-during-backup")
		req.StopDuringBackup = &v
	}
	if cmd.Flags().Changed("s3-storage-uuid") {
		v, _ := cmd.Flags().GetString("s3-storage-uuid")
		req.S3StorageUUID = &v
	}
	if cmd.Flags().Changed("retention-amount-locally") {
		v, _ := cmd.Flags().GetInt("retention-amount-locally")
		req.RetentionAmountLocally = &v
	}
	if cmd.Flags().Changed("retention-days-locally") {
		v, _ := cmd.Flags().GetInt("retention-days-locally")
		req.RetentionDaysLocally = &v
	}
	if cmd.Flags().Changed("retention-max-storage-locally") {
		v, _ := cmd.Flags().GetFloat64("retention-max-storage-locally")
		req.RetentionMaxStorageLocally = &v
	}
	if cmd.Flags().Changed("retention-amount-s3") {
		v, _ := cmd.Flags().GetInt("retention-amount-s3")
		req.RetentionAmountS3 = &v
	}
	if cmd.Flags().Changed("retention-days-s3") {
		v, _ := cmd.Flags().GetInt("retention-days-s3")
		req.RetentionDaysS3 = &v
	}
	if cmd.Flags().Changed("retention-max-storage-s3") {
		v, _ := cmd.Flags().GetFloat64("retention-max-storage-s3")
		req.RetentionMaxStorageS3 = &v
	}
	if cmd.Flags().Changed("timeout") {
		v, _ := cmd.Flags().GetInt("timeout")
		req.Timeout = &v
	}
	return req, nil
}

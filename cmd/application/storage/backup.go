package storage

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/common"
	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewSetBackupScheduleCommand sets or replaces a volume backup schedule (PUT).
func NewSetBackupScheduleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-backup <app_uuid> <storage_uuid>",
		Args:  cli.ExactArgs(2, "<app_uuid> <storage_uuid>"),
		Short: "Create or replace storage volume backup schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := common.BuildVolumeBackupScheduleRequest(cmd)
			if err != nil {
				return err
			}
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := service.NewApplicationService(client).SetStorageBackupSchedule(cmd.Context(), args[0], args[1], req)
			if err != nil {
				return err
			}
			formatName, _ := cmd.Flags().GetString("format")
			if formatName == "table" {
				formatName = "pretty"
			}
			formatter, err := output.NewFormatter(formatName, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(resp)
		},
	}
	common.BindVolumeBackupScheduleFlags(cmd)
	return cmd
}

// NewDeleteBackupScheduleCommand removes a volume backup schedule.
func NewDeleteBackupScheduleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-backup <app_uuid> <storage_uuid>",
		Args:  cli.ExactArgs(2, "<app_uuid> <storage_uuid>"),
		Short: "Delete storage volume backup schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			if err := service.NewApplicationService(client).DeleteStorageBackupSchedule(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}
			fmt.Println("Storage backup schedule deleted.")
			return nil
		},
	}
}

// NewRunBackupCommand runs a volume backup schedule now.
func NewRunBackupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run-backup <app_uuid> <storage_uuid>",
		Args:  cli.ExactArgs(2, "<app_uuid> <storage_uuid>"),
		Short: "Run storage volume backup now",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := service.NewApplicationService(client).RunStorageBackup(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			formatName, _ := cmd.Flags().GetString("format")
			if formatName == "table" {
				formatName = "pretty"
			}
			formatter, err := output.NewFormatter(formatName, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(resp)
		},
	}
}

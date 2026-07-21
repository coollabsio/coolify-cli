package storage

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/common"
	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewSetBackupScheduleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-backup <service_uuid> <storage_uuid>",
		Args:  cli.ExactArgs(2, "<service_uuid> <storage_uuid>"),
		Short: "Create or replace service storage volume backup schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := common.BuildVolumeBackupScheduleRequest(cmd)
			if err != nil {
				return err
			}
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := service.NewService(client).SetStorageBackupSchedule(cmd.Context(), args[0], args[1], req)
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

func NewDeleteBackupScheduleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-backup <service_uuid> <storage_uuid>",
		Args:  cli.ExactArgs(2, "<service_uuid> <storage_uuid>"),
		Short: "Delete service storage volume backup schedule",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			if err := service.NewService(client).DeleteStorageBackupSchedule(cmd.Context(), args[0], args[1]); err != nil {
				return err
			}
			fmt.Println("Storage backup schedule deleted.")
			return nil
		},
	}
}

func NewRunBackupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run-backup <service_uuid> <storage_uuid>",
		Args:  cli.ExactArgs(2, "<service_uuid> <storage_uuid>"),
		Short: "Run service storage volume backup now",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			resp, err := service.NewService(client).RunStorageBackup(cmd.Context(), args[0], args[1])
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

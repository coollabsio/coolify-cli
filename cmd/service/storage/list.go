package storage

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewListCommand returns the service storage list command
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <service_uuid>",
		Short: "List all storages for a service",
		Long:  `List all persistent volumes and file storages for a specific service.`,
		Args:  cli.ExactArgs(1, "<service_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			svcSvc := service.NewService(client)
			storages, err := svcSvc.ListStorages(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to list storages: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			return formatter.Format(storages)
		},
	}
}

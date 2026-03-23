package storage

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewListCommand returns the database storage list command
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <db_uuid>",
		Short: "List all storages for a database",
		Long:  `List all persistent volumes and file storages for a specific database.`,
		Args:  cli.ExactArgs(1, "<db_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := cli.CheckMinimumVersion(ctx, client, "4.0.0-beta.470"); err != nil {
				return err
			}

			dbSvc := service.NewDatabaseService(client)
			storages, err := dbSvc.ListStorages(ctx, args[0])
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

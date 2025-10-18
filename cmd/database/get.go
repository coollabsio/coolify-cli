package database

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewGetCommand gets database details
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <uuid>",
		Short: "Get database details",
		Long:  `Get detailed information about a specific database by UUID.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)
			database, err := dbService.Get(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to get database: %w", err)
			}

			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
			formatter, err := output.NewFormatter("table", output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(database)
		},
	}
}

package backup

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewListCommand lists all databases
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <database_uuid>",
		Short: "List all backup configurations for a database",
		Long:  `List all backup configurations for a specific database.`,
		Args:  cli.ExactArgs(1, "<database_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			dbUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)
			backups, err := dbService.ListBackups(ctx, dbUUID)
			if err != nil {
				return fmt.Errorf("failed to list backups: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(backups)
		},
	}
}

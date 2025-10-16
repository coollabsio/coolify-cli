package backup

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewExecutionCommand lists all databases
func NewExecutionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "executions <database_uuid> <backup_uuid>",
		Short: "List backup executions",
		Long:  `List all executions for a backup configuration. First UUID is the database, second is the specific backup configuration.`,
		Args:  cli.ExactArgs(2, "<database_uuid> <backup_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			dbUUID := args[0]
			backupUUID := args[1]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)
			executions, err := dbService.ListBackupExecutions(ctx, dbUUID, backupUUID)
			if err != nil {
				return fmt.Errorf("failed to list backup executions: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(executions)
		},
	}
}

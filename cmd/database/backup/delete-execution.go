package backup

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewDeleteExecutionCommand lists all databases
func NewDeleteExecutionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-execution <database_uuid> <backup_uuid> <execution_uuid>",
		Short: "Delete backup execution",
		Long:  `Delete a specific backup execution and optionally from S3. First UUID is the database, second is the backup configuration, third is the specific execution.`,
		Args:  cli.ExactArgs(3, "<database_uuid> <backup_uuid> <execution_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			dbUUID := args[0]
			backupUUID := args[1]
			executionUUID := args[2]

			force, _ := cmd.Flags().GetBool("force")
			deleteS3, _ := cmd.Flags().GetBool("delete-s3")

			if !force {
				fmt.Printf("Are you sure you want to delete backup execution %s? (y/N): ", executionUUID)
				reader := bufio.NewReader(os.Stdin)
				response, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("error reading input: %w", err)
				}
				response = strings.TrimSpace(strings.ToLower(response))
				if response != "y" && response != "yes" {
					fmt.Println("Delete cancelled")
					return nil
				}
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)
			err = dbService.DeleteBackupExecution(ctx, dbUUID, backupUUID, executionUUID, deleteS3)
			if err != nil {
				return fmt.Errorf("failed to delete backup execution: %w", err)
			}

			fmt.Println("Backup execution deleted successfully")
			return nil
		},
	}
}

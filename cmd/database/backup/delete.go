package backup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDeleteCommand deletes a database
func NewDeleteCommand() *cobra.Command {
	deleteBackupCmd := &cobra.Command{
		Use:   "delete <database_uuid> <backup_uuid>",
		Short: "Delete backup configuration",
		Long:  `Delete a backup configuration and optionally all its executions from S3. First UUID is the database, second is the specific backup configuration.`,
		Args:  cli.ExactArgs(2, "<database_uuid> <backup_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			dbUUID := args[0]
			backupUUID := args[1]

			force, _ := cmd.Flags().GetBool("force")
			deleteS3, _ := cmd.Flags().GetBool("delete-s3")

			if !force {
				fmt.Printf("Are you sure you want to delete backup configuration %s? (y/N): ", backupUUID)
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
			err = dbService.DeleteBackup(ctx, dbUUID, backupUUID, deleteS3)
			if err != nil {
				return fmt.Errorf("failed to delete backup: %w", err)
			}

			fmt.Println("Backup configuration deleted successfully")
			return nil
		},
	}

	deleteBackupCmd.Flags().Bool("delete-s3", false, "Delete backup files from S3")
	return deleteBackupCmd
}

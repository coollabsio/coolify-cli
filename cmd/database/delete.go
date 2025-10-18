package database

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDeleteCommand deletes a database
func NewDeleteCommand() *cobra.Command {
	deleteDatabaseCmd := &cobra.Command{
		Use:   "delete <uuid>",
		Short: "Delete a database",
		Long:  `Delete a database and optionally clean up its configurations, volumes, and networks.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			force, _ := cmd.Flags().GetBool("force")
			deleteConfigurations, _ := cmd.Flags().GetBool("delete-configurations")
			deleteVolumes, _ := cmd.Flags().GetBool("delete-volumes")
			dockerCleanup, _ := cmd.Flags().GetBool("docker-cleanup")
			deleteConnectedNetworks, _ := cmd.Flags().GetBool("delete-connected-networks")

			if !force {
				fmt.Printf("Are you sure you want to delete database %s? (y/N): ", uuid)
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
			err = dbService.Delete(ctx, uuid, deleteConfigurations, deleteVolumes, dockerCleanup, deleteConnectedNetworks)
			if err != nil {
				return fmt.Errorf("failed to delete database: %w", err)
			}

			fmt.Println("Database deleted successfully")
			return nil
		},
	}

	deleteDatabaseCmd.Flags().Bool("delete-configurations", true, "Delete configurations")
	deleteDatabaseCmd.Flags().Bool("delete-volumes", true, "Delete volumes")
	deleteDatabaseCmd.Flags().Bool("docker-cleanup", true, "Run docker cleanup")
	deleteDatabaseCmd.Flags().Bool("delete-connected-networks", true, "Delete connected networks")

	return deleteDatabaseCmd
}

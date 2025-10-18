package database

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewRestartCommand restarts a database
func NewRestartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restart <uuid>",
		Short: "Restart a database",
		Long:  `Restart a database by UUID.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)
			response, err := dbService.Restart(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to restart database: %w", err)
			}

			fmt.Println(response.Message)
			return nil
		},
	}
}

package database

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewStartCommand starts a database
func NewStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start <uuid>",
		Short: "Start a database",
		Long:  `Start a database by UUID.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)
			response, err := dbService.Start(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to start database: %w", err)
			}

			fmt.Println(response.Message)
			return nil
		},
	}
}

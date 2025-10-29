package database

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewStopCommand stops a database
func NewStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <uuid>",
		Short: "Stop a database",
		Long:  `Stop a database by UUID.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)
			response, err := dbService.Stop(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to stop database: %w", err)
			}

			fmt.Println(response.Message)
			return nil
		},
	}
}

package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewRestartCommand restarts a service
func NewRestartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restart <uuid>",
		Short: "Restart a service",
		Long:  `Restart a service (restart all containers).`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			serviceSvc := service.NewServiceService(client)
			resp, err := serviceSvc.Restart(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to restart service: %w", err)
			}

			fmt.Println(resp.Message)
			return nil
		},
	}
}

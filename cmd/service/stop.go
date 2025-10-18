package service

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewStopCommand stops a service
func NewStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <uuid>",
		Short: "Stop a service",
		Long:  `Stop a service (stop all containers).`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			serviceSvc := service.NewService(client)
			resp, err := serviceSvc.Stop(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to stop service: %w", err)
			}

			fmt.Println(resp.Message)
			return nil
		},
	}
}

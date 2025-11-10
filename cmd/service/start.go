package service

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewStartCommand starts a service
func NewStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start <uuid>",
		Short: "Start a service",
		Long:  `Start a service (deploy all containers).`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			serviceSvc := service.NewService(client)
			resp, err := serviceSvc.Start(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to start service: %w", err)
			}

			fmt.Println(resp.Message)
			return nil
		},
	}
}

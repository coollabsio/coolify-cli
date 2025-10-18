package application

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewRestartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restart <uuid>",
		Short: "Restart an application",
		Long:  `Restart a running application.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			appSvc := service.NewApplicationService(client)
			resp, err := appSvc.Restart(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to restart application: %w", err)
			}

			fmt.Println(resp.Message)
			return nil
		},
	}
}

package application

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

func NewStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <uuid>",
		Short: "Stop an application",
		Long:  `Stop a running application.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			appSvc := service.NewApplicationService(client)
			resp, err := appSvc.Stop(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to stop application: %w", err)
			}

			fmt.Println(resp.Message)
			return nil
		},
	}
}

package service

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewGetCommand gets service details
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <uuid>",
		Short: "Get service details",
		Long:  `Get detailed information about a specific service.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			serviceSvc := service.NewService(client)
			svc, err := serviceSvc.Get(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to get service: %w", err)
			}

			formatter, err := output.NewFormatter("table", output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(svc)
		},
	}
}

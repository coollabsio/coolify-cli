package service

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewListCommand lists all services
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all services",
		Long:  `List all services in Coolify.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			serviceSvc := service.NewService(client)
			services, err := serviceSvc.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list services: %w", err)
			}

			formatter, err := output.NewFormatter("table", output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(services)
		},
	}
}

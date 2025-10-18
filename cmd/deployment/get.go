package deployment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewGetCommand gets deployment details
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <uuid>",
		Short: "Get deployment details by UUID",
		Long:  `Get detailed information about a specific deployment by its UUID.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			deploySvc := service.NewDeploymentService(client)
			deployment, err := deploySvc.Get(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to get deployment: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(deployment)
		},
	}
}

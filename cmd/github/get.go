package github

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <app_uuid>",
		Short: "Get GitHub App details by UUID",
		Long:  `Get detailed information about a specific GitHub App integration.`,
		Args:  cli.ExactArgs(1, "<app_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			appUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			svc := service.NewGitHubAppService(client)
			app, err := svc.Get(ctx, appUUID)
			if err != nil {
				return fmt.Errorf("failed to get GitHub App: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(app)
		},
	}
}

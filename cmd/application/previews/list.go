package previews

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewListPreviewsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <app_uuid>",
		Short: "List preview deployments",
		Long:  `List the preview deployments of an application, newest pull request first.`,
		Args:  cli.ExactArgs(1, "<app_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := cli.CheckMinimumVersion(ctx, client, minimumVersion); err != nil {
				return err
			}

			appSvc := service.NewApplicationService(client)
			previews, err := appSvc.ListPreviews(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to list preview deployments: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(previews)
		},
	}
}

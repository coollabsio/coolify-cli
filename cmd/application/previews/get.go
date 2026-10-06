package previews

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewGetPreviewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <app_uuid> <pr_id>",
		Short: "Get a preview deployment",
		Long:  `Get a preview deployment of an application by pull request ID.`,
		Args:  cli.ExactArgs(2, "<app_uuid> <pr_id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			prID, err := parsePullRequestID(args[1])
			if err != nil {
				return err
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := cli.CheckMinimumVersion(ctx, client, minimumVersion); err != nil {
				return err
			}

			appSvc := service.NewApplicationService(client)
			preview, err := appSvc.GetPreview(ctx, args[0], prID)
			if err != nil {
				return fmt.Errorf("failed to get preview deployment: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(preview)
		},
	}
}

package gitlab

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewGetCommand gets a GitLab App by ID or UUID
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <app_id_or_uuid>",
		Short: "Get GitLab App details by ID or UUID",
		Long:  `Get detailed information about a specific GitLab App integration.`,
		Args:  cli.ExactArgs(1, "<app_id_or_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			idOrUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			svc := service.NewGitLabAppService(client)
			app, err := svc.Get(ctx, idOrUUID)
			if err != nil {
				return fmt.Errorf("failed to get GitLab App: %w", err)
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

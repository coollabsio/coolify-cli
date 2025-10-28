package github

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewListRepositoriesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "repos <app_uuid>",
		Short: "List repositories accessible by a GitHub App",
		Long:  `List all repositories that are accessible by the specified GitHub App.`,
		Args:  cli.ExactArgs(1, "<app_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			appUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			svc := service.NewGitHubAppService(client)
			repos, err := svc.ListRepositories(ctx, appUUID)
			if err != nil {
				return fmt.Errorf("failed to list repositories: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(repos)
		},
	}
}

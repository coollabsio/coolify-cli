package github

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewListBranchesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "branches <app_uuid> <owner/repo>",
		Short: "List branches for a repository",
		Long: `List all branches for a specific repository. Provide the app UUID and repository in owner/repo format.

Example: coolify github branches abc-123-def owner/repository`,
		Args: cli.ExactArgs(2, "<app_uuid> <owner/repo>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			appUUID := args[0]

			// Parse owner/repo
			ownerRepo := args[1]
			parts := cli.SplitOwnerRepo(ownerRepo)
			if len(parts) != 2 {
				return fmt.Errorf("invalid repository format. Expected 'owner/repo', got '%s'", ownerRepo)
			}
			owner, repo := parts[0], parts[1]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			svc := service.NewGitHubAppService(client)
			branches, err := svc.ListBranches(ctx, appUUID, owner, repo)
			if err != nil {
				return fmt.Errorf("failed to list branches: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(branches)
		},
	}
}

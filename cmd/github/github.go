package github

import (
	"github.com/spf13/cobra"
)

func NewGitHubCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "github",
		Aliases: []string{"gh", "github-app", "github-apps"},
		Short:   "Manage GitHub App integrations",
		Long:    `Manage GitHub App integrations for private repository deployments.`,
	}

	// Add main database commands
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewUpdateCommand())
	cmd.AddCommand(NewDeleteCommand())
	cmd.AddCommand(NewListRepositoriesCommand())
	cmd.AddCommand(NewListBranchesCommand())

	return cmd
}

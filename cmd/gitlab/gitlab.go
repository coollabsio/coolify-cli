package gitlab

import (
	"github.com/spf13/cobra"
)

// NewGitLabCommand creates the parent command for GitLab App management
func NewGitLabCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gitlab",
		Aliases: []string{"gl", "gitlab-app", "gitlab-apps"},
		Short:   "Manage GitLab App integrations",
		Long:    `Manage GitLab App (OAuth) sources for private repository deployments.`,
	}

	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewUpdateCommand())
	cmd.AddCommand(NewDeleteCommand())

	return cmd
}

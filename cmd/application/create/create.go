package create

import "github.com/spf13/cobra"

// NewCreateCommand creates the create parent command with all subcommands
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new application",
		Long: `Create a new application from various sources.

Available source types:
  public      Create from a public git repository
  github      Create from a private repository using GitHub App
  deploy-key  Create from a private repository using SSH deploy key
  dockerfile  Create from a custom Dockerfile
  dockerimage Create from a pre-built Docker image

Examples:
  coolify app create public --server-uuid <uuid> --project-uuid <uuid> --environment-name production \
    --git-repository "https://github.com/user/repo" --git-branch main --build-pack nixpacks --ports-exposes 3000

  coolify app create github --server-uuid <uuid> --project-uuid <uuid> --environment-name production \
    --github-app-uuid <uuid> --git-repository "user/repo" --git-branch main --build-pack nixpacks --ports-exposes 3000

  coolify app create dockerimage --server-uuid <uuid> --project-uuid <uuid> --environment-name production \
    --docker-registry-image-name "nginx:latest" --ports-exposes 80`,
	}

	// Add all create subcommands
	cmd.AddCommand(NewPublicCommand())
	cmd.AddCommand(NewGitHubCommand())
	cmd.AddCommand(NewDeployKeyCommand())
	cmd.AddCommand(NewDockerfileCommand())
	cmd.AddCommand(NewDockerImageCommand())

	return cmd
}

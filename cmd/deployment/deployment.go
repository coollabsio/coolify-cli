package deployment

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
)

const dockerTagMinVersion = "4.0.0-beta.471"

// NewDeploymentCommand creates the deployment parent command with all subcommands
func NewDeploymentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy related commands",
	}

	// Add all deployment subcommands
	cmd.AddCommand(NewUUIDCommand())
	cmd.AddCommand(NewNameCommand())
	cmd.AddCommand(NewBatchCommand())
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewCancelCommand())

	return cmd
}

func addDeployFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("force", false, "Force deployment")
	cmd.Flags().Int("pull-request-id", 0, "Pull request ID for preview deployments")
	cmd.Flags().String("docker-tag", "", "Docker image tag override for the deployment")
}

func getDeployRequest(cmd *cobra.Command, uuid string) models.DeployRequest {
	req := models.DeployRequest{
		UUID: uuid,
	}

	if cmd.Flags().Changed("force") {
		force, _ := cmd.Flags().GetBool("force")
		req.Force = &force
	}
	if cmd.Flags().Changed("pull-request-id") {
		pullRequestID, _ := cmd.Flags().GetInt("pull-request-id")
		req.PullRequestID = &pullRequestID
	}
	if cmd.Flags().Changed("docker-tag") {
		dockerTag, _ := cmd.Flags().GetString("docker-tag")
		req.DockerTag = &dockerTag
	}

	return req
}

func validateDeployFlags(ctx context.Context, cmd *cobra.Command, client *api.Client) error {
	if cmd.Flags().Changed("docker-tag") {
		return cli.CheckMinimumVersion(ctx, client, dockerTagMinVersion)
	}

	return nil
}

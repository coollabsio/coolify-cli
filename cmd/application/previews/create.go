package previews

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewCreatePreviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <app_uuid> <pr_id>",
		Aliases: []string{"open", "deploy"},
		Short:   "Open or redeploy a preview deployment",
		Long: `Open a preview deployment for a pull request and queue its deployment.
When the preview already exists, it is deployed again.

Git based applications need --git-type, unless they use a GitHub App or GitLab App source.
Bitbucket pull requests also need --commit. Docker Image applications need --docker-tag for a new preview.`,
		Example: `  coolify app previews create <app_uuid> 42 --git-type github
  coolify app previews create <app_uuid> 42 --git-type bitbucket --commit 1a2b3c4d
  coolify app previews create <app_uuid> 42 --docker-tag pr-42
  coolify app previews create <app_uuid> 42 --git-type gitlab --no-deploy`,
		Args: cli.ExactArgs(2, "<app_uuid> <pr_id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			prID, err := parsePullRequestID(args[1])
			if err != nil {
				return err
			}

			req := models.ApplicationPreviewCreateRequest{PullRequestID: prID}
			for flag, target := range map[string]**string{
				"url":        &req.PullRequestHTMLURL,
				"git-type":   &req.GitType,
				"commit":     &req.Commit,
				"docker-tag": &req.DockerTag,
			} {
				if cmd.Flags().Changed(flag) {
					value, _ := cmd.Flags().GetString(flag)
					*target = &value
				}
			}
			if cmd.Flags().Changed("force") {
				force, _ := cmd.Flags().GetBool("force")
				req.Force = &force
			}
			if noDeploy, _ := cmd.Flags().GetBool("no-deploy"); noDeploy {
				instantDeploy := false
				req.InstantDeploy = &instantDeploy
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := cli.CheckMinimumVersion(ctx, client, minimumVersion); err != nil {
				return err
			}

			appSvc := service.NewApplicationService(client)
			resp, err := appSvc.CreatePreview(ctx, args[0], req)
			if err != nil {
				return fmt.Errorf("failed to create preview deployment: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			if format != output.FormatTable {
				formatter, err := output.NewFormatter(format, output.Options{})
				if err != nil {
					return err
				}
				return formatter.Format(resp)
			}

			fmt.Println(resp.Message)
			if resp.DeploymentUUID != nil && *resp.DeploymentUUID != "" {
				fmt.Printf("Deployment UUID: %s\n", *resp.DeploymentUUID)
			}
			if resp.Preview.Domains != nil && *resp.Preview.Domains != "" {
				fmt.Printf("Domains: %s\n", *resp.Preview.Domains)
			}
			return nil
		},
	}

	cmd.Flags().String("url", "", "Pull request URL")
	cmd.Flags().String("git-type", "", "Git provider of the pull request: github, gitlab, gitea, or bitbucket")
	cmd.Flags().String("commit", "", "Commit SHA to deploy (required for Bitbucket)")
	cmd.Flags().String("docker-tag", "", "Docker image tag (Docker Image applications only)")
	cmd.Flags().Bool("force", false, "Rebuild without cache")
	cmd.Flags().Bool("no-deploy", false, "Only create the preview, do not queue a deployment")
	return cmd
}

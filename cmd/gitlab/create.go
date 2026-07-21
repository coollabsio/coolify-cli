package gitlab

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewCreateCommand creates a GitLab App integration
func NewCreateCommand() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a GitLab App integration",
		Long: `Create a new GitLab App (OAuth) source. This allows deploying private repositories from GitLab.com or self-hosted GitLab.

Required flags: --name, --html-url

Example:
  coolify gitlab create --name "My GitLab" --html-url "https://gitlab.com" --client-id "..." --client-secret "..."`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			name, _ := cmd.Flags().GetString("name")
			htmlURL, _ := cmd.Flags().GetString("html-url")

			req := &models.GitLabAppCreateRequest{
				Name:    name,
				HTMLURL: htmlURL,
			}

			if cmd.Flags().Changed("api-url") {
				apiURL, _ := cmd.Flags().GetString("api-url")
				req.APIURL = &apiURL
			}
			if cmd.Flags().Changed("custom-user") {
				user, _ := cmd.Flags().GetString("custom-user")
				req.CustomUser = &user
			}
			if cmd.Flags().Changed("custom-port") {
				port, _ := cmd.Flags().GetInt("custom-port")
				req.CustomPort = &port
			}
			if cmd.Flags().Changed("group-name") {
				group, _ := cmd.Flags().GetString("group-name")
				req.GroupName = &group
			}
			if cmd.Flags().Changed("client-id") {
				clientID, _ := cmd.Flags().GetString("client-id")
				req.ClientID = &clientID
			}
			if cmd.Flags().Changed("client-secret") {
				clientSecret, _ := cmd.Flags().GetString("client-secret")
				req.ClientSecret = &clientSecret
			}
			if cmd.Flags().Changed("webhook-token") {
				token, _ := cmd.Flags().GetString("webhook-token")
				req.WebhookToken = &token
			}
			if cmd.Flags().Changed("redirect-uri") {
				redirect, _ := cmd.Flags().GetString("redirect-uri")
				req.RedirectURI = &redirect
			}
			if cmd.Flags().Changed("system-wide") {
				systemWide, _ := cmd.Flags().GetBool("system-wide")
				req.IsSystemWide = &systemWide
			}

			svc := service.NewGitLabAppService(client)
			app, err := svc.Create(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create GitLab App: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(app)
		},
	}

	createCmd.Flags().String("name", "", "GitLab App name (required)")
	createCmd.Flags().String("html-url", "", "GitLab instance URL (required, e.g. https://gitlab.com)")
	createCmd.Flags().String("api-url", "", "GitLab API URL (defaults to {html-url}/api/v4)")
	createCmd.Flags().String("custom-user", "", "Custom SSH user (default: git)")
	createCmd.Flags().Int("custom-port", 0, "Custom SSH port (default: 22)")
	createCmd.Flags().String("group-name", "", "Optional comma-separated group filter")
	createCmd.Flags().String("client-id", "", "GitLab OAuth Application ID")
	createCmd.Flags().String("client-secret", "", "GitLab OAuth Application Secret")
	createCmd.Flags().String("webhook-token", "", "Webhook secret token (auto-generated when omitted)")
	createCmd.Flags().String("redirect-uri", "", "OAuth redirect URI registered in GitLab")
	createCmd.Flags().Bool("system-wide", false, "Make this app system-wide (non-cloud instances only)")

	_ = createCmd.MarkFlagRequired("name")
	_ = createCmd.MarkFlagRequired("html-url")

	return createCmd
}

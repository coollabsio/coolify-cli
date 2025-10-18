package github

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewCreateCommand() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a GitHub App integration",
		Long: `Create a new GitHub App integration. This allows you to deploy private repositories from GitHub.

Required flags: --name, --api-url, --html-url, --app-id, --installation-id, --client-id, --client-secret, --private-key-uuid

Example: coolify github create --name "My GitHub App" --api-url "https://api.github.com" --html-url "https://github.com" --app-id 123456 --installation-id 789012 --client-id "Iv1.abc123" --client-secret "secret123" --private-key-uuid "abc-123-def-456"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			name, _ := cmd.Flags().GetString("name")
			apiURL, _ := cmd.Flags().GetString("api-url")
			htmlURL, _ := cmd.Flags().GetString("html-url")
			appID, _ := cmd.Flags().GetInt("app-id")
			installationID, _ := cmd.Flags().GetInt("installation-id")
			clientID, _ := cmd.Flags().GetString("client-id")
			clientSecret, _ := cmd.Flags().GetString("client-secret")
			privateKeyUUID, _ := cmd.Flags().GetString("private-key-uuid")

			req := &models.GitHubAppCreateRequest{
				Name:           name,
				APIURL:         apiURL,
				HTMLURL:        htmlURL,
				AppID:          appID,
				InstallationID: installationID,
				ClientID:       clientID,
				ClientSecret:   clientSecret,
				PrivateKeyUUID: privateKeyUUID,
			}

			// Optional fields
			if cmd.Flags().Changed("organization") {
				org, _ := cmd.Flags().GetString("organization")
				req.Organization = &org
			}
			if cmd.Flags().Changed("custom-user") {
				user, _ := cmd.Flags().GetString("custom-user")
				req.CustomUser = &user
			}
			if cmd.Flags().Changed("custom-port") {
				port, _ := cmd.Flags().GetInt("custom-port")
				req.CustomPort = &port
			}
			if cmd.Flags().Changed("webhook-secret") {
				secret, _ := cmd.Flags().GetString("webhook-secret")
				req.WebhookSecret = &secret
			}
			if cmd.Flags().Changed("system-wide") {
				systemWide, _ := cmd.Flags().GetBool("system-wide")
				req.IsSystemWide = &systemWide
			}

			svc := service.NewGitHubAppService(client)
			app, err := svc.Create(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create GitHub App: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(app)
		},
	}

	createCmd.Flags().String("name", "", "GitHub App name (required)")
	createCmd.Flags().String("organization", "", "GitHub organization")
	createCmd.Flags().String("api-url", "", "GitHub API URL (required, e.g., https://api.github.com)")
	createCmd.Flags().String("html-url", "", "GitHub HTML URL (required, e.g., https://github.com)")
	createCmd.Flags().String("custom-user", "", "Custom user for SSH (default: git)")
	createCmd.Flags().Int("custom-port", 0, "Custom port for SSH (default: 22)")
	createCmd.Flags().Int("app-id", 0, "GitHub App ID (required)")
	createCmd.Flags().Int("installation-id", 0, "GitHub Installation ID (required)")
	createCmd.Flags().String("client-id", "", "GitHub OAuth Client ID (required)")
	createCmd.Flags().String("client-secret", "", "GitHub OAuth Client Secret (required)")
	createCmd.Flags().String("webhook-secret", "", "GitHub Webhook Secret")
	createCmd.Flags().String("private-key-uuid", "", "UUID of existing private key (required)")
	createCmd.Flags().Bool("system-wide", false, "Is this app system-wide (cloud only)")

	_ = createCmd.MarkFlagRequired("name")
	_ = createCmd.MarkFlagRequired("api-url")
	_ = createCmd.MarkFlagRequired("html-url")
	_ = createCmd.MarkFlagRequired("app-id")
	_ = createCmd.MarkFlagRequired("installation-id")
	_ = createCmd.MarkFlagRequired("client-id")
	_ = createCmd.MarkFlagRequired("client-secret")
	_ = createCmd.MarkFlagRequired("private-key-uuid")

	return createCmd
}

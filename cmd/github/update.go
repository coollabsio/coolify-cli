package github

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewUpdateCommand() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update <app_uuid>",
		Short: "Update a GitHub App integration",
		Long:  `Update an existing GitHub App integration. Provide the app UUID and the fields you want to update.`,
		Args:  cli.ExactArgs(1, "<app_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			appUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			req := &models.GitHubAppUpdateRequest{}

			// Update only fields that were explicitly provided
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = &name
			}
			if cmd.Flags().Changed("organization") {
				org, _ := cmd.Flags().GetString("organization")
				req.Organization = &org
			}
			if cmd.Flags().Changed("api-url") {
				apiURL, _ := cmd.Flags().GetString("api-url")
				req.APIURL = &apiURL
			}
			if cmd.Flags().Changed("html-url") {
				htmlURL, _ := cmd.Flags().GetString("html-url")
				req.HTMLURL = &htmlURL
			}
			if cmd.Flags().Changed("custom-user") {
				user, _ := cmd.Flags().GetString("custom-user")
				req.CustomUser = &user
			}
			if cmd.Flags().Changed("custom-port") {
				port, _ := cmd.Flags().GetInt("custom-port")
				req.CustomPort = &port
			}
			if cmd.Flags().Changed("app-id") {
				id, _ := cmd.Flags().GetInt("app-id")
				req.AppID = &id
			}
			if cmd.Flags().Changed("installation-id") {
				id, _ := cmd.Flags().GetInt("installation-id")
				req.InstallationID = &id
			}
			if cmd.Flags().Changed("client-id") {
				clientID, _ := cmd.Flags().GetString("client-id")
				req.ClientID = &clientID
			}
			if cmd.Flags().Changed("client-secret") {
				clientSecret, _ := cmd.Flags().GetString("client-secret")
				req.ClientSecret = &clientSecret
			}
			if cmd.Flags().Changed("webhook-secret") {
				secret, _ := cmd.Flags().GetString("webhook-secret")
				req.WebhookSecret = &secret
			}
			if cmd.Flags().Changed("private-key-uuid") {
				uuid, _ := cmd.Flags().GetString("private-key-uuid")
				req.PrivateKeyUUID = &uuid
			}
			if cmd.Flags().Changed("system-wide") {
				systemWide, _ := cmd.Flags().GetBool("system-wide")
				req.IsSystemWide = &systemWide
			}

			svc := service.NewGitHubAppService(client)
			err = svc.Update(ctx, appUUID, req)
			if err != nil {
				return fmt.Errorf("failed to update GitHub App: %w", err)
			}

			fmt.Println("GitHub App updated successfully")
			return nil
		},
	}

	updateCmd.Flags().String("name", "", "GitHub App name")
	updateCmd.Flags().String("organization", "", "GitHub organization")
	updateCmd.Flags().String("api-url", "", "GitHub API URL")
	updateCmd.Flags().String("html-url", "", "GitHub HTML URL")
	updateCmd.Flags().String("custom-user", "", "Custom user for SSH")
	updateCmd.Flags().Int("custom-port", 0, "Custom port for SSH")
	updateCmd.Flags().Int("app-id", 0, "GitHub App ID")
	updateCmd.Flags().Int("installation-id", 0, "GitHub Installation ID")
	updateCmd.Flags().String("client-id", "", "GitHub OAuth Client ID")
	updateCmd.Flags().String("client-secret", "", "GitHub OAuth Client Secret")
	updateCmd.Flags().String("webhook-secret", "", "GitHub Webhook Secret")
	updateCmd.Flags().String("private-key-uuid", "", "UUID of private key")
	updateCmd.Flags().Bool("system-wide", false, "Is this app system-wide")

	return updateCmd
}

package gitlab

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewUpdateCommand updates a GitLab App integration
func NewUpdateCommand() *cobra.Command {
	updateCmd := &cobra.Command{
		Use:   "update <app_id_or_uuid>",
		Short: "Update a GitLab App integration",
		Long:  `Update an existing GitLab App integration. Provide the app ID or UUID and the fields you want to update.`,
		Args:  cli.ExactArgs(1, "<app_id_or_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			idOrUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			req := &models.GitLabAppUpdateRequest{}

			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = &name
			}
			if cmd.Flags().Changed("html-url") {
				htmlURL, _ := cmd.Flags().GetString("html-url")
				req.HTMLURL = &htmlURL
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
			app, err := svc.Update(ctx, idOrUUID, req)
			if err != nil {
				return fmt.Errorf("failed to update GitLab App: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}

			return formatter.Format(app)
		},
	}

	updateCmd.Flags().String("name", "", "GitLab App name")
	updateCmd.Flags().String("html-url", "", "GitLab instance URL")
	updateCmd.Flags().String("api-url", "", "GitLab API URL")
	updateCmd.Flags().String("custom-user", "", "Custom SSH user")
	updateCmd.Flags().Int("custom-port", 0, "Custom SSH port")
	updateCmd.Flags().String("group-name", "", "Optional comma-separated group filter")
	updateCmd.Flags().String("client-id", "", "GitLab OAuth Application ID")
	updateCmd.Flags().String("client-secret", "", "GitLab OAuth Application Secret")
	updateCmd.Flags().String("webhook-token", "", "Webhook secret token")
	updateCmd.Flags().String("redirect-uri", "", "OAuth redirect URI")
	updateCmd.Flags().Bool("system-wide", false, "Is this app system-wide")

	return updateCmd
}

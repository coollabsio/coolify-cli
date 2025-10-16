package cmd

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

var githubCmd = &cobra.Command{
	Use:     "github",
	Aliases: []string{"gh", "github-app", "github-apps"},
	Short:   "Manage GitHub App integrations",
	Long:    `Manage GitHub App integrations for private repository deployments.`,
}

var listGitHubAppsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all GitHub App integrations",
	Long:  `List all GitHub App integrations configured in your Coolify instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		svc := service.NewGitHubAppService(client)
		apps, err := svc.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list GitHub Apps: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return err
		}

		return formatter.Format(apps)
	},
}

var getGitHubAppCmd = &cobra.Command{
	Use:   "get <app_uuid>",
	Short: "Get GitHub App details by UUID",
	Long:  `Get detailed information about a specific GitHub App integration.`,
	Args:  exactArgs(1, "<app_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		appUUID := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		svc := service.NewGitHubAppService(client)
		app, err := svc.Get(ctx, appUUID)
		if err != nil {
			return fmt.Errorf("failed to get GitHub App: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return err
		}

		return formatter.Format(app)
	},
}

var createGitHubAppCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a GitHub App integration",
	Long: `Create a new GitHub App integration. This allows you to deploy private repositories from GitHub.

Required flags: --name, --api-url, --html-url, --app-id, --installation-id, --client-id, --client-secret, --private-key-uuid

Example: coolify github create --name "My GitHub App" --api-url "https://api.github.com" --html-url "https://github.com" --app-id 123456 --installation-id 789012 --client-id "Iv1.abc123" --client-secret "secret123" --private-key-uuid "abc-123-def-456"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
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

var updateGitHubAppCmd = &cobra.Command{
	Use:   "update <app_uuid>",
	Short: "Update a GitHub App integration",
	Long:  `Update an existing GitHub App integration. Provide the app UUID and the fields you want to update.`,
	Args:  exactArgs(1, "<app_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		appUUID := args[0]

		client, err := getAPIClient(cmd)
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

var deleteGitHubAppCmd = &cobra.Command{
	Use:   "delete <app_uuid>",
	Short: "Delete a GitHub App integration",
	Long:  `Delete a GitHub App integration. The app must not be used by any applications.`,
	Args:  exactArgs(1, "<app_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		appUUID := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		force, _ := cmd.Flags().GetBool("force")

		// Prompt for confirmation unless --force is used
		if !force {
			var response string
			fmt.Printf("Are you sure you want to delete GitHub App %s? This cannot be undone. (yes/no): ", appUUID)
			fmt.Scanln(&response)

			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled.")
				return nil
			}
		}

		svc := service.NewGitHubAppService(client)
		err = svc.Delete(ctx, appUUID)
		if err != nil {
			return fmt.Errorf("failed to delete GitHub App: %w", err)
		}

		fmt.Println("GitHub App deleted successfully")
		return nil
	},
}

var listRepositoriesCmd = &cobra.Command{
	Use:   "repos <app_uuid>",
	Short: "List repositories accessible by a GitHub App",
	Long:  `List all repositories that are accessible by the specified GitHub App.`,
	Args:  exactArgs(1, "<app_uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		appUUID := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		svc := service.NewGitHubAppService(client)
		repos, err := svc.ListRepositories(ctx, appUUID)
		if err != nil {
			return fmt.Errorf("failed to list repositories: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return err
		}

		return formatter.Format(repos)
	},
}

var listBranchesCmd = &cobra.Command{
	Use:   "branches <app_uuid> <owner/repo>",
	Short: "List branches for a repository",
	Long: `List all branches for a specific repository. Provide the app UUID and repository in owner/repo format.

Example: coolify github branches abc-123-def owner/repository`,
	Args: exactArgs(2, "<app_uuid> <owner/repo>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		appUUID := args[0]

		// Parse owner/repo
		ownerRepo := args[1]
		parts := splitOwnerRepo(ownerRepo)
		if len(parts) != 2 {
			return fmt.Errorf("invalid repository format. Expected 'owner/repo', got '%s'", ownerRepo)
		}
		owner, repo := parts[0], parts[1]

		client, err := getAPIClient(cmd)
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

func init() {
	// Create command flags
	createGitHubAppCmd.Flags().String("name", "", "GitHub App name (required)")
	createGitHubAppCmd.Flags().String("organization", "", "GitHub organization")
	createGitHubAppCmd.Flags().String("api-url", "", "GitHub API URL (required, e.g., https://api.github.com)")
	createGitHubAppCmd.Flags().String("html-url", "", "GitHub HTML URL (required, e.g., https://github.com)")
	createGitHubAppCmd.Flags().String("custom-user", "", "Custom user for SSH (default: git)")
	createGitHubAppCmd.Flags().Int("custom-port", 0, "Custom port for SSH (default: 22)")
	createGitHubAppCmd.Flags().Int("app-id", 0, "GitHub App ID (required)")
	createGitHubAppCmd.Flags().Int("installation-id", 0, "GitHub Installation ID (required)")
	createGitHubAppCmd.Flags().String("client-id", "", "GitHub OAuth Client ID (required)")
	createGitHubAppCmd.Flags().String("client-secret", "", "GitHub OAuth Client Secret (required)")
	createGitHubAppCmd.Flags().String("webhook-secret", "", "GitHub Webhook Secret")
	createGitHubAppCmd.Flags().String("private-key-uuid", "", "UUID of existing private key (required)")
	createGitHubAppCmd.Flags().Bool("system-wide", false, "Is this app system-wide (cloud only)")

	createGitHubAppCmd.MarkFlagRequired("name")
	createGitHubAppCmd.MarkFlagRequired("api-url")
	createGitHubAppCmd.MarkFlagRequired("html-url")
	createGitHubAppCmd.MarkFlagRequired("app-id")
	createGitHubAppCmd.MarkFlagRequired("installation-id")
	createGitHubAppCmd.MarkFlagRequired("client-id")
	createGitHubAppCmd.MarkFlagRequired("client-secret")
	createGitHubAppCmd.MarkFlagRequired("private-key-uuid")

	// Update command flags (all optional)
	updateGitHubAppCmd.Flags().String("name", "", "GitHub App name")
	updateGitHubAppCmd.Flags().String("organization", "", "GitHub organization")
	updateGitHubAppCmd.Flags().String("api-url", "", "GitHub API URL")
	updateGitHubAppCmd.Flags().String("html-url", "", "GitHub HTML URL")
	updateGitHubAppCmd.Flags().String("custom-user", "", "Custom user for SSH")
	updateGitHubAppCmd.Flags().Int("custom-port", 0, "Custom port for SSH")
	updateGitHubAppCmd.Flags().Int("app-id", 0, "GitHub App ID")
	updateGitHubAppCmd.Flags().Int("installation-id", 0, "GitHub Installation ID")
	updateGitHubAppCmd.Flags().String("client-id", "", "GitHub OAuth Client ID")
	updateGitHubAppCmd.Flags().String("client-secret", "", "GitHub OAuth Client Secret")
	updateGitHubAppCmd.Flags().String("webhook-secret", "", "GitHub Webhook Secret")
	updateGitHubAppCmd.Flags().String("private-key-uuid", "", "UUID of private key")
	updateGitHubAppCmd.Flags().Bool("system-wide", false, "Is this app system-wide")

	// Delete command flags
	deleteGitHubAppCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	rootCmd.AddCommand(githubCmd)
	githubCmd.AddCommand(listGitHubAppsCmd)
	githubCmd.AddCommand(getGitHubAppCmd)
	githubCmd.AddCommand(createGitHubAppCmd)
	githubCmd.AddCommand(updateGitHubAppCmd)
	githubCmd.AddCommand(deleteGitHubAppCmd)
	githubCmd.AddCommand(listRepositoriesCmd)
	githubCmd.AddCommand(listBranchesCmd)
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/parser"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

var applicationsCmd = &cobra.Command{
	Use:     "app",
	Aliases: []string{"apps", "application", "applications"},
	Short:   "Application related commands",
	Long:    `Manage Coolify applications - list, get, create, update, delete, and control application lifecycle.`,
}

var listApplicationsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications",
	Long:  `List all applications in your Coolify instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		appSvc := service.NewApplicationService(client)
		apps, err := appSvc.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list applications: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

		// For JSON/pretty formats, return the full application structure
		if format != output.FormatTable {
			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}
			return formatter.Format(apps)
		}

		// For table format, convert to simplified rows
		var rows []models.ApplicationListItem
		for _, app := range apps {
			rows = append(rows, models.ApplicationListItem{
				UUID:        app.UUID,
				Name:        app.Name,
				Description: app.Description,
				Status:      app.Status,
				GitBranch:   app.GitBranch,
				FQDN:        app.FQDN,
			})
		}

		formatter, err := output.NewFormatter(format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return err
		}

		return formatter.Format(rows)
	},
}

var getApplicationCmd = &cobra.Command{
	Use:   "get <uuid>",
	Short: "Get application details by UUID",
	Long:  `Retrieve detailed information about a specific application.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		appSvc := service.NewApplicationService(client)
		app, err := appSvc.Get(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to get application: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

		formatter, err := output.NewFormatter(format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return err
		}

		return formatter.Format(app)
	},
}

var updateApplicationCmd = &cobra.Command{
	Use:   "update <uuid>",
	Short: "Update application configuration",
	Long:  `Update configuration for a specific application. Only specified fields will be updated.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		// Build update request from flags
		req := models.ApplicationUpdateRequest{}
		hasUpdates := false

		// Basic configuration
		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			req.Name = &name
			hasUpdates = true
		}
		if cmd.Flags().Changed("description") {
			desc, _ := cmd.Flags().GetString("description")
			req.Description = &desc
			hasUpdates = true
		}
		if cmd.Flags().Changed("git-branch") {
			branch, _ := cmd.Flags().GetString("git-branch")
			req.GitBranch = &branch
			hasUpdates = true
		}
		if cmd.Flags().Changed("git-repository") {
			repo, _ := cmd.Flags().GetString("git-repository")
			req.GitRepository = &repo
			hasUpdates = true
		}
		if cmd.Flags().Changed("domains") {
			domains, _ := cmd.Flags().GetString("domains")
			req.Domains = &domains
			hasUpdates = true
		}

		// Build configuration
		if cmd.Flags().Changed("build-command") {
			buildCmd, _ := cmd.Flags().GetString("build-command")
			req.BuildCommand = &buildCmd
			hasUpdates = true
		}
		if cmd.Flags().Changed("start-command") {
			startCmd, _ := cmd.Flags().GetString("start-command")
			req.StartCommand = &startCmd
			hasUpdates = true
		}
		if cmd.Flags().Changed("install-command") {
			installCmd, _ := cmd.Flags().GetString("install-command")
			req.InstallCommand = &installCmd
			hasUpdates = true
		}
		if cmd.Flags().Changed("base-directory") {
			baseDir, _ := cmd.Flags().GetString("base-directory")
			req.BaseDirectory = &baseDir
			hasUpdates = true
		}
		if cmd.Flags().Changed("publish-directory") {
			publishDir, _ := cmd.Flags().GetString("publish-directory")
			req.PublishDirectory = &publishDir
			hasUpdates = true
		}

		// Docker configuration
		if cmd.Flags().Changed("dockerfile") {
			dockerfile, _ := cmd.Flags().GetString("dockerfile")
			req.Dockerfile = &dockerfile
			hasUpdates = true
		}
		if cmd.Flags().Changed("docker-image") {
			image, _ := cmd.Flags().GetString("docker-image")
			req.DockerRegistryImageName = &image
			hasUpdates = true
		}
		if cmd.Flags().Changed("docker-tag") {
			tag, _ := cmd.Flags().GetString("docker-tag")
			req.DockerRegistryImageTag = &tag
			hasUpdates = true
		}

		// Ports
		if cmd.Flags().Changed("ports-exposes") {
			ports, _ := cmd.Flags().GetString("ports-exposes")
			req.PortsExposes = &ports
			hasUpdates = true
		}
		if cmd.Flags().Changed("ports-mappings") {
			ports, _ := cmd.Flags().GetString("ports-mappings")
			req.PortsMappings = &ports
			hasUpdates = true
		}

		// Health check
		if cmd.Flags().Changed("health-check-enabled") {
			enabled, _ := cmd.Flags().GetBool("health-check-enabled")
			req.HealthCheckEnabled = &enabled
			hasUpdates = true
		}
		if cmd.Flags().Changed("health-check-path") {
			path, _ := cmd.Flags().GetString("health-check-path")
			req.HealthCheckPath = &path
			hasUpdates = true
		}

		if !hasUpdates {
			return fmt.Errorf("no fields to update. Use --help to see available flags")
		}

		appSvc := service.NewApplicationService(client)
		app, err := appSvc.Update(ctx, uuid, req)
		if err != nil {
			return fmt.Errorf("failed to update application: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

		formatter, err := output.NewFormatter(format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return err
		}

		return formatter.Format(app)
	},
}

var deleteApplicationCmd = &cobra.Command{
	Use:   "delete <uuid>",
	Short: "Delete an application",
	Long:  `Delete an application. This action cannot be undone.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		force, _ := cmd.Flags().GetBool("force")

		// Prompt for confirmation unless --force is used
		if !force {
			var response string
			fmt.Printf("Are you sure you want to delete application %s? This cannot be undone. (yes/no): ", uuid)
			fmt.Scanln(&response)

			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled.")
				return nil
			}
		}

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		appSvc := service.NewApplicationService(client)
		err = appSvc.Delete(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to delete application: %w", err)
		}

		fmt.Printf("Application %s deleted successfully.\n", uuid)
		return nil
	},
}

func init() {
	// Define update command flags (most common ones)
	updateApplicationCmd.Flags().String("name", "", "Application name")
	updateApplicationCmd.Flags().String("description", "", "Application description")
	updateApplicationCmd.Flags().String("git-branch", "", "Git branch")
	updateApplicationCmd.Flags().String("git-repository", "", "Git repository URL")
	updateApplicationCmd.Flags().String("domains", "", "Domains (comma-separated)")
	updateApplicationCmd.Flags().String("build-command", "", "Build command")
	updateApplicationCmd.Flags().String("start-command", "", "Start command")
	updateApplicationCmd.Flags().String("install-command", "", "Install command")
	updateApplicationCmd.Flags().String("base-directory", "", "Base directory")
	updateApplicationCmd.Flags().String("publish-directory", "", "Publish directory")
	updateApplicationCmd.Flags().String("dockerfile", "", "Dockerfile content")
	updateApplicationCmd.Flags().String("docker-image", "", "Docker image name")
	updateApplicationCmd.Flags().String("docker-tag", "", "Docker image tag")
	updateApplicationCmd.Flags().String("ports-exposes", "", "Exposed ports")
	updateApplicationCmd.Flags().String("ports-mappings", "", "Port mappings")
	updateApplicationCmd.Flags().Bool("health-check-enabled", false, "Enable health check")
	updateApplicationCmd.Flags().String("health-check-path", "", "Health check path")

	// Define delete command flags
	deleteApplicationCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Define start command flags
	startApplicationCmd.Flags().Bool("force", false, "Force rebuild")
	startApplicationCmd.Flags().Bool("instant-deploy", false, "Instant deploy (skip queuing)")

	// Define logs command flags
	logsApplicationCmd.Flags().IntP("lines", "n", 100, "Number of log lines to retrieve")
	logsApplicationCmd.Flags().BoolP("follow", "f", false, "Follow log output (like tail -f)")

	// Define envs create command flags
	createEnvCmd.Flags().String("key", "", "Environment variable key (required)")
	createEnvCmd.Flags().String("value", "", "Environment variable value (required)")
	createEnvCmd.Flags().Bool("build-time", false, "Available at build time")
	createEnvCmd.Flags().Bool("preview", false, "Available in preview deployments")
	createEnvCmd.Flags().Bool("is-literal", false, "Treat value as literal (don't interpolate variables)")
	createEnvCmd.Flags().Bool("is-multiline", false, "Value is multiline")

	// Define envs update command flags
	updateEnvCmd.Flags().String("key", "", "New environment variable key")
	updateEnvCmd.Flags().String("value", "", "New environment variable value")
	updateEnvCmd.Flags().Bool("build-time", false, "Available at build time")
	updateEnvCmd.Flags().Bool("preview", false, "Available in preview deployments")
	updateEnvCmd.Flags().Bool("is-literal", false, "Treat value as literal (don't interpolate variables)")
	updateEnvCmd.Flags().Bool("is-multiline", false, "Value is multiline")

	// Define envs delete command flags
	deleteEnvCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	// Define envs sync command flags
	syncEnvCmd.Flags().StringP("file", "f", "", "Path to .env file (required)")
	syncEnvCmd.Flags().Bool("build-time", false, "Make all variables available at build time")
	syncEnvCmd.Flags().Bool("preview", false, "Make all variables available in preview deployments")
	syncEnvCmd.Flags().Bool("is-literal", false, "Treat all values as literal (don't interpolate variables)")

	rootCmd.AddCommand(applicationsCmd)
	applicationsCmd.AddCommand(listApplicationsCmd)
	applicationsCmd.AddCommand(getApplicationCmd)
	applicationsCmd.AddCommand(updateApplicationCmd)
	applicationsCmd.AddCommand(deleteApplicationCmd)
	applicationsCmd.AddCommand(startApplicationCmd)
	applicationsCmd.AddCommand(stopApplicationCmd)
	applicationsCmd.AddCommand(restartApplicationCmd)
	applicationsCmd.AddCommand(logsApplicationCmd)
	applicationsCmd.AddCommand(envsApplicationCmd)
	envsApplicationCmd.AddCommand(listEnvsCmd)
	envsApplicationCmd.AddCommand(getEnvCmd)
	envsApplicationCmd.AddCommand(createEnvCmd)
	envsApplicationCmd.AddCommand(updateEnvCmd)
	envsApplicationCmd.AddCommand(deleteEnvCmd)
	envsApplicationCmd.AddCommand(syncEnvCmd)
}

var startApplicationCmd = &cobra.Command{
	Use:     "start <uuid>",
	Aliases: []string{"deploy"},
	Short:   "Start an application",
	Long:    `Start an application (initiates a deployment).`,
	Args:    exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		force, _ := cmd.Flags().GetBool("force")
		instantDeploy, _ := cmd.Flags().GetBool("instant-deploy")

		appSvc := service.NewApplicationService(client)
		resp, err := appSvc.Start(ctx, uuid, force, instantDeploy)
		if err != nil {
			return fmt.Errorf("failed to start application: %w", err)
		}

		fmt.Println(resp.Message)
		if resp.DeploymentUUID != nil && *resp.DeploymentUUID != "" {
			fmt.Printf("Deployment UUID: %s\n", *resp.DeploymentUUID)
		}
		return nil
	},
}

var stopApplicationCmd = &cobra.Command{
	Use:   "stop <uuid>",
	Short: "Stop an application",
	Long:  `Stop a running application.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		appSvc := service.NewApplicationService(client)
		resp, err := appSvc.Stop(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to stop application: %w", err)
		}

		fmt.Println(resp.Message)
		return nil
	},
}

var restartApplicationCmd = &cobra.Command{
	Use:   "restart <uuid>",
	Short: "Restart an application",
	Long:  `Restart a running application.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		appSvc := service.NewApplicationService(client)
		resp, err := appSvc.Restart(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to restart application: %w", err)
		}

		fmt.Println(resp.Message)
		return nil
	},
}

var logsApplicationCmd = &cobra.Command{
	Use:   "logs <uuid>",
	Short: "Get application logs",
	Long:  `Retrieve logs for an application. Use --follow to continuously stream new logs.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		lines, _ := cmd.Flags().GetInt("lines")
		follow, _ := cmd.Flags().GetBool("follow")
		appSvc := service.NewApplicationService(client)

		if !follow {
			// One-time fetch
			resp, err := appSvc.Logs(ctx, uuid, lines)
			if err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
			}
			fmt.Print(resp.Logs)
			return nil
		}

		// Follow mode: poll for new logs
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Track the last log content to avoid duplicates
		lastLogs := ""

		// Fetch initial logs
		resp, err := appSvc.Logs(ctx, uuid, lines)
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}
		fmt.Print(resp.Logs)
		lastLogs = resp.Logs

		// Poll for new logs
		for {
			select {
			case <-sigChan:
				fmt.Println("\nStopping log follow...")
				return nil
			case <-ticker.C:
				resp, err := appSvc.Logs(ctx, uuid, lines)
				if err != nil {
					// Don't fail on transient errors in follow mode
					continue
				}
				// Only print if logs have changed
				if resp.Logs != lastLogs {
					// Print only the new content
					if len(resp.Logs) > len(lastLogs) && strings.HasPrefix(resp.Logs, lastLogs) {
						fmt.Print(resp.Logs[len(lastLogs):])
					} else {
						// Logs were truncated or changed, print all
						fmt.Print(resp.Logs)
					}
					lastLogs = resp.Logs
				}
			}
		}
	},
}

var envsApplicationCmd = &cobra.Command{
	Use:     "env",
	Aliases: []string{"envs", "environment"},
	Short:   "Manage application environment variables",
	Long:    `List and manage environment variables for applications. All commands require the application UUID first to establish context.`,
}

var listEnvsCmd = &cobra.Command{
	Use:   "list <app_uuid>",
	Short: "List all environment variables for an application",
	Long:  `List all environment variables for a specific application.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		appSvc := service.NewApplicationService(client)
		envs, err := appSvc.ListEnvs(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to list environment variables: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

		// Mask sensitive values unless --show-sensitive is used
		if !showSensitive {
			for i := range envs {
				envs[i].Value = "********"
				if envs[i].RealValue != nil {
					masked := "********"
					envs[i].RealValue = &masked
				}
			}
		}

		formatter, err := output.NewFormatter(format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return err
		}

		return formatter.Format(envs)
	},
}

var getEnvCmd = &cobra.Command{
	Use:   "get <app_uuid> <env_uuid_or_key>",
	Short: "Get environment variable details",
	Long:  `Get detailed information about a specific environment variable by UUID or key name.`,
	Args:  exactArgs(2, "<uuid1> <uuid2>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		appUUID := args[0]
		envUUID := args[1]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		appSvc := service.NewApplicationService(client)
		env, err := appSvc.GetEnv(ctx, appUUID, envUUID)
		if err != nil {
			return fmt.Errorf("failed to get environment variable: %w", err)
		}

		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

		// Mask sensitive value unless --show-sensitive is used
		if !showSensitive {
			env.Value = "********"
			if env.RealValue != nil {
				masked := "********"
				env.RealValue = &masked
			}
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(env)
	},
}

var createEnvCmd = &cobra.Command{
	Use:   "create <app_uuid>",
	Short: "Create an environment variable for an application",
	Long:  `Create a new environment variable for a specific application. Use --key and --value flags to specify the variable.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		isBuildTime, _ := cmd.Flags().GetBool("build-time")
		isPreview, _ := cmd.Flags().GetBool("preview")
		isLiteral, _ := cmd.Flags().GetBool("is-literal")
		isMultiline, _ := cmd.Flags().GetBool("is-multiline")

		if key == "" {
			return fmt.Errorf("--key is required")
		}
		if value == "" {
			return fmt.Errorf("--value is required")
		}

		req := &models.EnvironmentVariableCreateRequest{
			Key:   key,
			Value: value,
		}

		// Only set flags if they were explicitly provided
		if cmd.Flags().Changed("build-time") {
			req.IsBuildTime = &isBuildTime
		}
		if cmd.Flags().Changed("preview") {
			req.IsPreview = &isPreview
		}
		if cmd.Flags().Changed("is-literal") {
			req.IsLiteral = &isLiteral
		}
		if cmd.Flags().Changed("is-multiline") {
			req.IsMultiline = &isMultiline
		}

		appSvc := service.NewApplicationService(client)
		env, err := appSvc.CreateEnv(ctx, uuid, req)
		if err != nil {
			return fmt.Errorf("failed to create environment variable: %w", err)
		}

		fmt.Printf("Environment variable '%s' created successfully.\n", env.Key)
		fmt.Printf("UUID: %s\n", env.UUID)
		return nil
	},
}

var updateEnvCmd = &cobra.Command{
	Use:   "update <app_uuid> <env_uuid>",
	Short: "Update an environment variable",
	Long:  `Update an existing environment variable. First UUID is the application, second is the specific environment variable to update.`,
	Args:  exactArgs(2, "<uuid1> <uuid2>"),
	RunE: func(cmd *cobra.Command, args []string) error{
		ctx := context.Background()
		appUUID := args[0]
		envUUID := args[1]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		req := &models.EnvironmentVariableUpdateRequest{
			UUID: envUUID,
		}

		// Only set fields that were provided
		if cmd.Flags().Changed("key") {
			key, _ := cmd.Flags().GetString("key")
			req.Key = &key
		}
		if cmd.Flags().Changed("value") {
			value, _ := cmd.Flags().GetString("value")
			req.Value = &value
		}
		if cmd.Flags().Changed("build-time") {
			isBuildTime, _ := cmd.Flags().GetBool("build-time")
			req.IsBuildTime = &isBuildTime
		}
		if cmd.Flags().Changed("preview") {
			isPreview, _ := cmd.Flags().GetBool("preview")
			req.IsPreview = &isPreview
		}
		if cmd.Flags().Changed("is-literal") {
			isLiteral, _ := cmd.Flags().GetBool("is-literal")
			req.IsLiteral = &isLiteral
		}
		if cmd.Flags().Changed("is-multiline") {
			isMultiline, _ := cmd.Flags().GetBool("is-multiline")
			req.IsMultiline = &isMultiline
		}

		// Check if at least one field is being updated
		if req.Key == nil && req.Value == nil && req.IsBuildTime == nil && req.IsPreview == nil && req.IsLiteral == nil && req.IsMultiline == nil {
			return fmt.Errorf("at least one field must be provided to update (--key, --value, --build-time, --preview, --is-literal, or --is-multiline)")
		}

		appSvc := service.NewApplicationService(client)
		env, err := appSvc.UpdateEnv(ctx, appUUID, req)
		if err != nil {
			return fmt.Errorf("failed to update environment variable: %w", err)
		}

		fmt.Printf("Environment variable '%s' updated successfully.\n", env.Key)
		return nil
	},
}

var deleteEnvCmd = &cobra.Command{
	Use:   "delete <app_uuid> <env_uuid>",
	Short: "Delete an environment variable",
	Long:  `Delete an environment variable from an application. First UUID is the application, second is the specific environment variable to delete.`,
	Args:  exactArgs(2, "<uuid1> <uuid2>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		appUUID := args[0]
		envUUID := args[1]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		force, _ := cmd.Flags().GetBool("force")

		// Prompt for confirmation unless --force is used
		if !force {
			var response string
			fmt.Printf("Are you sure you want to delete this environment variable? (yes/no): ")
			fmt.Scanln(&response)

			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled.")
				return nil
			}
		}

		appSvc := service.NewApplicationService(client)
		err = appSvc.DeleteEnv(ctx, appUUID, envUUID)
		if err != nil {
			return fmt.Errorf("failed to delete environment variable: %w", err)
		}

		fmt.Println("Environment variable deleted successfully.")
		return nil
	},
}

var syncEnvCmd = &cobra.Command{
	Use:   "sync <app_uuid>",
	Short: "Sync environment variables from a .env file",
	Long: `Sync environment variables from a .env file. This command intelligently:
- Updates existing environment variables with new values
- Creates new environment variables that don't exist yet
- Uses efficient bulk operations where possible

Example: coolify app env sync abc123 --file .env.production`,
	Args: exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file is required")
		}

		isBuildTime, _ := cmd.Flags().GetBool("build-time")
		isPreview, _ := cmd.Flags().GetBool("preview")
		isLiteral, _ := cmd.Flags().GetBool("is-literal")

		// Parse the .env file
		envVars, err := parser.ParseEnvFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to parse .env file: %w", err)
		}

		if len(envVars) == 0 {
			fmt.Println("No environment variables found in file.")
			return nil
		}

		fmt.Printf("Found %d environment variables in file. Syncing...\n", len(envVars))

		// Fetch existing environment variables
		appSvc := service.NewApplicationService(client)
		existingEnvs, err := appSvc.ListEnvs(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to list existing environment variables: %w", err)
		}

		// Build a map of existing env vars by key
		existingMap := make(map[string]models.EnvironmentVariable)
		for _, env := range existingEnvs {
			existingMap[env.Key] = env
		}

		// Separate into updates and creates
		var toUpdate []models.EnvironmentVariableCreateRequest
		var toCreate []models.EnvironmentVariableCreateRequest

		for _, envVar := range envVars {
			req := models.EnvironmentVariableCreateRequest{
				Key:   envVar.Key,
				Value: envVar.Value,
			}

			// Apply flags if explicitly provided
			if cmd.Flags().Changed("build-time") {
				req.IsBuildTime = &isBuildTime
			}
			if cmd.Flags().Changed("preview") {
				req.IsPreview = &isPreview
			}
			if cmd.Flags().Changed("is-literal") {
				req.IsLiteral = &isLiteral
			}

			// Auto-detect multiline values
			if strings.Contains(envVar.Value, "\n") {
				multiline := true
				req.IsMultiline = &multiline
			}

			if _, exists := existingMap[envVar.Key]; exists {
				toUpdate = append(toUpdate, req)
			} else {
				toCreate = append(toCreate, req)
			}
		}

		updateCount := 0
		createCount := 0
		failCount := 0

		// Perform bulk update if there are vars to update
		if len(toUpdate) > 0 {
			fmt.Printf("Updating %d existing variables...\n", len(toUpdate))
			bulkReq := &service.BulkUpdateEnvsRequest{
				Data: toUpdate,
			}
			_, err := appSvc.BulkUpdateEnvs(ctx, uuid, bulkReq)
			if err != nil {
				fmt.Printf("  ✗ Bulk update failed: %v\n", err)
				failCount += len(toUpdate)
			} else {
				updateCount = len(toUpdate)
				fmt.Printf("  ✓ Successfully updated %d variables\n", updateCount)
			}
		}

		// Create new variables one by one
		if len(toCreate) > 0 {
			fmt.Printf("Creating %d new variables...\n", len(toCreate))
			for _, req := range toCreate {
				_, err := appSvc.CreateEnv(ctx, uuid, &req)
				if err != nil {
					fmt.Printf("  ✗ Failed to create '%s': %v\n", req.Key, err)
					failCount++
				} else {
					fmt.Printf("  ✓ Created '%s'\n", req.Key)
					createCount++
				}
			}
		}

		fmt.Printf("\nSync complete: %d updated, %d created, %d failed\n", updateCount, createCount, failCount)

		if failCount > 0 {
			return fmt.Errorf("some environment variables failed to sync")
		}

		return nil
	},
}

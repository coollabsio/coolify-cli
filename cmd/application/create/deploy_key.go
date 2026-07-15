package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/application/appflags"
	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDeployKeyCommand returns the create deploy-key application command
func NewDeployKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-key",
		Short: "Create an application from a private repository using SSH deploy key",
		Long: `Create a new application from a private git repository using SSH deploy key authentication.

Use 'coolify privatekeys list' to find your private key UUID.

Examples:
  coolify app create deploy-key --server-uuid <uuid> --project-uuid <uuid> --environment-name production \
    --private-key-uuid <uuid> --git-repository "git@github.com:owner/repo.git" --git-branch main \
    --build-pack nixpacks --ports-exposes 3000

  coolify app create deploy-key --server-uuid <uuid> --project-uuid <uuid> --environment-name production \
    --private-key-uuid <uuid> --git-repository "git@gitlab.com:owner/repo.git" --git-branch main \
    --build-pack dockerfile --ports-exposes 8080 --instant-deploy`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			// Get required flags
			serverUUID, _ := cmd.Flags().GetString("server-uuid")
			projectUUID, _ := cmd.Flags().GetString("project-uuid")
			privateKeyUUID, _ := cmd.Flags().GetString("private-key-uuid")
			gitRepository, _ := cmd.Flags().GetString("git-repository")
			gitBranch, _ := cmd.Flags().GetString("git-branch")
			buildPack, _ := cmd.Flags().GetString("build-pack")
			portsExposes, _ := cmd.Flags().GetString("ports-exposes")
			environmentName, _ := cmd.Flags().GetString("environment-name")
			environmentUUID, _ := cmd.Flags().GetString("environment-uuid")

			// Validate required fields
			if serverUUID == "" || projectUUID == "" {
				return fmt.Errorf("--server-uuid and --project-uuid are required")
			}
			if privateKeyUUID == "" {
				return fmt.Errorf("--private-key-uuid is required")
			}
			if gitRepository == "" || gitBranch == "" {
				return fmt.Errorf("--git-repository and --git-branch are required")
			}
			if buildPack == "" || portsExposes == "" {
				return fmt.Errorf("--build-pack and --ports-exposes are required")
			}
			if environmentName == "" && environmentUUID == "" {
				return fmt.Errorf("either --environment-name or --environment-uuid must be provided")
			}

			req := &models.ApplicationCreateDeployKeyRequest{
				ServerUUID:     serverUUID,
				ProjectUUID:    projectUUID,
				PrivateKeyUUID: privateKeyUUID,
				GitRepository:  gitRepository,
				GitBranch:      gitBranch,
				BuildPack:      buildPack,
				PortsExposes:   portsExposes,
			}

			if environmentName != "" {
				req.EnvironmentName = &environmentName
			}
			if environmentUUID != "" {
				req.EnvironmentUUID = &environmentUUID
			}

			// Optional fields
			setOptionalStringFlag(cmd, "name", &req.Name)
			setOptionalStringFlag(cmd, "description", &req.Description)
			setOptionalStringFlag(cmd, "domains", &req.Domains)
			setOptionalStringFlag(cmd, "git-commit-sha", &req.GitCommitSHA)
			setOptionalStringFlag(cmd, "destination-uuid", &req.DestinationUUID)
			setOptionalStringFlag(cmd, "build-command", &req.BuildCommand)
			setOptionalStringFlag(cmd, "start-command", &req.StartCommand)
			setOptionalStringFlag(cmd, "install-command", &req.InstallCommand)
			setOptionalStringFlag(cmd, "base-directory", &req.BaseDirectory)
			setOptionalStringFlag(cmd, "publish-directory", &req.PublishDirectory)
			setOptionalStringFlag(cmd, "ports-mappings", &req.PortsMappings)
			setOptionalStringFlag(cmd, "limits-cpus", &req.LimitsCPUs)
			setOptionalStringFlag(cmd, "limits-memory", &req.LimitsMemory)
			setOptionalBoolFlag(cmd, "instant-deploy", &req.InstantDeploy)
			setOptionalBoolFlag(cmd, "health-check-enabled", &req.HealthCheckEnabled)
			setOptionalStringFlag(cmd, "health-check-path", &req.HealthCheckPath)
			setOptionalStringFlag(cmd, "dockerfile-target-build", &req.DockerfileTargetBuild)
			appflags.ApplySettingsFlags(cmd, &req.ApplicationSettingsRequest)
			appflags.ApplyTagsFlag(cmd, &req.Tags)

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			appSvc := service.NewApplicationService(client)
			app, err := appSvc.CreateDeployKey(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create application: %w", err)
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

	// Required flags
	cmd.Flags().String("server-uuid", "", "Server UUID (required)")
	cmd.Flags().String("project-uuid", "", "Project UUID (required)")
	cmd.Flags().String("environment-name", "", "Environment name")
	cmd.Flags().String("environment-uuid", "", "Environment UUID")
	cmd.Flags().String("private-key-uuid", "", "Private key UUID (required)")
	cmd.Flags().String("git-repository", "", "Git repository SSH URL, e.g., 'git@github.com:owner/repo.git' (required)")
	cmd.Flags().String("git-branch", "", "Git branch (required)")
	cmd.Flags().String("build-pack", "", "Build pack: nixpacks, static, dockerfile, dockercompose (required)")
	cmd.Flags().String("ports-exposes", "", "Exposed ports, e.g., '3000' or '3000,8080' (required)")

	// Optional flags
	cmd.Flags().String("name", "", "Application name")
	cmd.Flags().String("description", "", "Application description")
	cmd.Flags().String("domains", "", "Domain(s) for the application")
	cmd.Flags().Bool("instant-deploy", false, "Deploy immediately after creation")
	cmd.Flags().String("git-commit-sha", "", "Specific commit SHA to deploy")
	cmd.Flags().String("destination-uuid", "", "Destination UUID if server has multiple destinations")
	cmd.Flags().String("build-command", "", "Custom build command")
	cmd.Flags().String("start-command", "", "Custom start command")
	cmd.Flags().String("install-command", "", "Custom install command")
	cmd.Flags().String("base-directory", "", "Base directory for the application")
	cmd.Flags().String("publish-directory", "", "Publish directory for static builds")
	cmd.Flags().String("ports-mappings", "", "Port mappings (host:container)")
	cmd.Flags().String("limits-cpus", "", "CPU limit")
	cmd.Flags().String("limits-memory", "", "Memory limit")
	cmd.Flags().Bool("health-check-enabled", false, "Enable health checks")
	cmd.Flags().String("health-check-path", "", "Health check path")
	cmd.Flags().String("dockerfile-target-build", "", "Dockerfile target build stage")
	appflags.BindSettingsFlags(cmd)
	appflags.BindTagsFlag(cmd)

	return cmd
}

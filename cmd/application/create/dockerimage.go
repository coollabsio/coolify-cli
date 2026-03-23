package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDockerImageCommand returns the create dockerimage application command
func NewDockerImageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dockerimage",
		Short: "Create an application from a pre-built Docker image",
		Long: `Create a new application from a pre-built Docker image from a registry.

Examples:
  coolify app create dockerimage --server-uuid <uuid> --project-uuid <uuid> --environment-name production \
    --docker-registry-image-name "nginx:latest" --ports-exposes 80

  coolify app create dockerimage --server-uuid <uuid> --project-uuid <uuid> --environment-name production \
    --docker-registry-image-name "ghcr.io/myorg/myapp" --docker-registry-image-tag "v1.0.0" \
    --ports-exposes 3000 --domains "myapp.example.com" --instant-deploy`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			// Get required flags
			serverUUID, _ := cmd.Flags().GetString("server-uuid")
			projectUUID, _ := cmd.Flags().GetString("project-uuid")
			dockerRegistryImageName, _ := cmd.Flags().GetString("docker-registry-image-name")
			portsExposes, _ := cmd.Flags().GetString("ports-exposes")
			environmentName, _ := cmd.Flags().GetString("environment-name")
			environmentUUID, _ := cmd.Flags().GetString("environment-uuid")

			// Validate required fields
			if serverUUID == "" || projectUUID == "" {
				return fmt.Errorf("--server-uuid and --project-uuid are required")
			}
			if dockerRegistryImageName == "" {
				return fmt.Errorf("--docker-registry-image-name is required")
			}
			if portsExposes == "" {
				return fmt.Errorf("--ports-exposes is required")
			}
			if environmentName == "" && environmentUUID == "" {
				return fmt.Errorf("either --environment-name or --environment-uuid must be provided")
			}

			req := &models.ApplicationCreateDockerImageRequest{
				ServerUUID:              serverUUID,
				ProjectUUID:             projectUUID,
				DockerRegistryImageName: dockerRegistryImageName,
				PortsExposes:            portsExposes,
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
			setOptionalStringFlag(cmd, "destination-uuid", &req.DestinationUUID)
			setOptionalStringFlag(cmd, "docker-registry-image-tag", &req.DockerRegistryImageTag)
			setOptionalStringFlag(cmd, "ports-mappings", &req.PortsMappings)
			setOptionalStringFlag(cmd, "limits-cpus", &req.LimitsCPUs)
			setOptionalStringFlag(cmd, "limits-memory", &req.LimitsMemory)
			setOptionalBoolFlag(cmd, "instant-deploy", &req.InstantDeploy)
			setOptionalBoolFlag(cmd, "health-check-enabled", &req.HealthCheckEnabled)
			setOptionalStringFlag(cmd, "health-check-path", &req.HealthCheckPath)
			setOptionalStringFlag(cmd, "dockerfile-target-build", &req.DockerfileTargetBuild)

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			appSvc := service.NewApplicationService(client)
			app, err := appSvc.CreateDockerImage(ctx, req)
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
	cmd.Flags().String("docker-registry-image-name", "", "Docker image name from registry (required)")
	cmd.Flags().String("ports-exposes", "", "Exposed ports, e.g., '80' or '80,443' (required)")

	// Optional flags
	cmd.Flags().String("name", "", "Application name")
	cmd.Flags().String("description", "", "Application description")
	cmd.Flags().String("domains", "", "Domain(s) for the application")
	cmd.Flags().Bool("instant-deploy", false, "Deploy immediately after creation")
	cmd.Flags().String("destination-uuid", "", "Destination UUID if server has multiple destinations")
	cmd.Flags().String("docker-registry-image-tag", "", "Docker image tag (defaults to 'latest')")
	cmd.Flags().String("ports-mappings", "", "Port mappings (host:container)")
	cmd.Flags().String("limits-cpus", "", "CPU limit")
	cmd.Flags().String("limits-memory", "", "Memory limit")
	cmd.Flags().Bool("health-check-enabled", false, "Enable health checks")
	cmd.Flags().String("health-check-path", "", "Health check path")
	cmd.Flags().String("dockerfile-target-build", "", "Dockerfile target build stage")

	return cmd
}

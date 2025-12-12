package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDockerfileCommand returns the create dockerfile application command
func NewDockerfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dockerfile",
		Short: "Create an application from a custom Dockerfile",
		Long: `Create a new application from a custom Dockerfile content.

Examples:
  coolify app create dockerfile --server-uuid <uuid> --project-uuid <uuid> --environment-name production \
    --dockerfile "FROM node:18\nWORKDIR /app\nCOPY . .\nRUN npm install\nCMD [\"npm\", \"start\"]"

  coolify app create dockerfile --server-uuid <uuid> --project-uuid <uuid> --environment-name production \
    --dockerfile "$(cat Dockerfile)" --ports-exposes 3000 --instant-deploy`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			// Get required flags
			serverUUID, _ := cmd.Flags().GetString("server-uuid")
			projectUUID, _ := cmd.Flags().GetString("project-uuid")
			dockerfile, _ := cmd.Flags().GetString("dockerfile")
			environmentName, _ := cmd.Flags().GetString("environment-name")
			environmentUUID, _ := cmd.Flags().GetString("environment-uuid")

			// Validate required fields
			if serverUUID == "" || projectUUID == "" {
				return fmt.Errorf("--server-uuid and --project-uuid are required")
			}
			if dockerfile == "" {
				return fmt.Errorf("--dockerfile is required")
			}
			if environmentName == "" && environmentUUID == "" {
				return fmt.Errorf("either --environment-name or --environment-uuid must be provided")
			}

			req := &models.ApplicationCreateDockerfileRequest{
				ServerUUID:  serverUUID,
				ProjectUUID: projectUUID,
				Dockerfile:  dockerfile,
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
			setOptionalStringFlag(cmd, "ports-exposes", &req.PortsExposes)
			setOptionalStringFlag(cmd, "ports-mappings", &req.PortsMappings)
			setOptionalStringFlag(cmd, "limits-cpus", &req.LimitsCPUs)
			setOptionalStringFlag(cmd, "limits-memory", &req.LimitsMemory)
			setOptionalBoolFlag(cmd, "instant-deploy", &req.InstantDeploy)
			setOptionalBoolFlag(cmd, "health-check-enabled", &req.HealthCheckEnabled)
			setOptionalStringFlag(cmd, "health-check-path", &req.HealthCheckPath)

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			appSvc := service.NewApplicationService(client)
			app, err := appSvc.CreateDockerfile(ctx, req)
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
	cmd.Flags().String("dockerfile", "", "Dockerfile content (required)")

	// Optional flags
	cmd.Flags().String("name", "", "Application name")
	cmd.Flags().String("description", "", "Application description")
	cmd.Flags().String("domains", "", "Domain(s) for the application")
	cmd.Flags().Bool("instant-deploy", false, "Deploy immediately after creation")
	cmd.Flags().String("destination-uuid", "", "Destination UUID if server has multiple destinations")
	cmd.Flags().String("ports-exposes", "", "Exposed ports, e.g., '3000' or '3000,8080'")
	cmd.Flags().String("ports-mappings", "", "Port mappings (host:container)")
	cmd.Flags().String("limits-cpus", "", "CPU limit")
	cmd.Flags().String("limits-memory", "", "Memory limit")
	cmd.Flags().Bool("health-check-enabled", false, "Enable health checks")
	cmd.Flags().String("health-check-path", "", "Health check path")

	return cmd
}

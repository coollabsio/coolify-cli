package application

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <uuid>",
		Short: "Update application configuration",
		Long:  `Update configuration for a specific application. Only specified fields will be updated.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			req := models.ApplicationUpdateRequest{}
			hasUpdates := false

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

	cmd.Flags().String("name", "", "Application name")
	cmd.Flags().String("description", "", "Application description")
	cmd.Flags().String("git-branch", "", "Git branch")
	cmd.Flags().String("git-repository", "", "Git repository URL")
	cmd.Flags().String("domains", "", "Domains (comma-separated)")
	cmd.Flags().String("build-command", "", "Build command")
	cmd.Flags().String("start-command", "", "Start command")
	cmd.Flags().String("install-command", "", "Install command")
	cmd.Flags().String("base-directory", "", "Base directory")
	cmd.Flags().String("publish-directory", "", "Publish directory")
	cmd.Flags().String("dockerfile", "", "Dockerfile content")
	cmd.Flags().String("docker-image", "", "Docker image name")
	cmd.Flags().String("docker-tag", "", "Docker image tag")
	cmd.Flags().String("ports-exposes", "", "Exposed ports")
	cmd.Flags().String("ports-mappings", "", "Port mappings")
	cmd.Flags().Bool("health-check-enabled", false, "Enable health check")
	cmd.Flags().String("health-check-path", "", "Health check path")

	return cmd
}

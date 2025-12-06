package service

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// validServiceTypes contains all supported one-click service types
var validServiceTypes = []string{
	"activepieces",
	"appsmith",
	"appwrite",
	"authentik",
	"babybuddy",
	"budge",
	"changedetection",
	"chatwoot",
	"classicpress-with-mariadb",
	"classicpress-with-mysql",
	"classicpress-without-database",
	"cloudflared",
	"code-server",
	"dashboard",
	"directus",
	"directus-with-postgresql",
	"docker-registry",
	"docuseal",
	"docuseal-with-postgres",
	"dokuwiki",
	"duplicati",
	"emby",
	"embystat",
	"fider",
	"filebrowser",
	"firefly",
	"formbricks",
	"ghost",
	"gitea",
	"gitea-with-mariadb",
	"gitea-with-mysql",
	"gitea-with-postgresql",
	"glance",
	"glances",
	"glitchtip",
	"grafana",
	"grafana-with-postgresql",
	"grocy",
	"heimdall",
	"homepage",
	"jellyfin",
	"kuzzle",
	"listmonk",
	"logto",
	"mediawiki",
	"meilisearch",
	"metabase",
	"metube",
	"minio",
	"moodle",
	"n8n",
	"n8n-with-postgresql",
	"next-image-transformation",
	"nextcloud",
	"nocodb",
	"odoo",
	"openblocks",
	"pairdrop",
	"penpot",
	"phpmyadmin",
	"pocketbase",
	"posthog",
	"reactive-resume",
	"rocketchat",
	"shlink",
	"slash",
	"snapdrop",
	"statusnook",
	"stirling-pdf",
	"supabase",
	"syncthing",
	"tolgee",
	"trigger",
	"trigger-with-external-database",
	"twenty",
	"umami",
	"unleash-with-postgresql",
	"unleash-without-database",
	"uptime-kuma",
	"vaultwarden",
	"vikunja",
	"weblate",
	"whoogle",
	"wordpress-with-mariadb",
	"wordpress-with-mysql",
	"wordpress-without-database",
}

func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <type>",
		Short: "Create a new one-click service",
		Long: `Create a new one-click service of the specified type.

Use 'coolify service create --list-types' to see all available service types.

Examples:
  coolify service create wordpress-with-mysql --server-uuid=<uuid> --project-uuid=<uuid> --environment-name=production
  coolify service create ghost --server-uuid=<uuid> --project-uuid=<uuid> --environment-name=production --name="My Blog"
  coolify service create n8n --server-uuid=<uuid> --project-uuid=<uuid> --environment-name=production --instant-deploy

Popular service types:
  - wordpress-with-mysql, wordpress-with-mariadb, wordpress-without-database
  - ghost, plausible, umami, uptime-kuma
  - n8n, n8n-with-postgresql
  - nextcloud, gitea, minio
  - grafana, metabase, nocodb
  - supabase, pocketbase, appwrite`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Handle --list-types flag
			listTypes, _ := cmd.Flags().GetBool("list-types")
			if listTypes {
				fmt.Println("Available one-click service types:")
				fmt.Println()
				for _, t := range validServiceTypes {
					fmt.Printf("  %s\n", t)
				}
				return nil
			}

			// Require type argument if not listing
			if len(args) == 0 {
				return fmt.Errorf("service type is required. Use --list-types to see available types")
			}

			serviceType := args[0]

			// Validate service type
			isValid := false
			for _, t := range validServiceTypes {
				if t == serviceType {
					isValid = true
					break
				}
			}
			if !isValid {
				return fmt.Errorf("invalid service type '%s'. Use --list-types to see available types", serviceType)
			}

			serverUUID, _ := cmd.Flags().GetString("server-uuid")
			projectUUID, _ := cmd.Flags().GetString("project-uuid")
			environmentName, _ := cmd.Flags().GetString("environment-name")
			environmentUUID, _ := cmd.Flags().GetString("environment-uuid")

			if serverUUID == "" || projectUUID == "" {
				return fmt.Errorf("--server-uuid and --project-uuid are required")
			}

			if environmentName == "" && environmentUUID == "" {
				return fmt.Errorf("either --environment-name or --environment-uuid must be provided")
			}

			req := &models.ServiceCreateRequest{
				Type:        serviceType,
				ServerUUID:  serverUUID,
				ProjectUUID: projectUUID,
			}

			if environmentName != "" {
				req.EnvironmentName = environmentName
			}
			if environmentUUID != "" {
				req.EnvironmentUUID = &environmentUUID
			}

			// Handle optional flags
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = &name
			}
			if cmd.Flags().Changed("description") {
				desc, _ := cmd.Flags().GetString("description")
				req.Description = &desc
			}
			if cmd.Flags().Changed("destination-uuid") {
				dest, _ := cmd.Flags().GetString("destination-uuid")
				req.Destination = &dest
			}
			if cmd.Flags().Changed("instant-deploy") {
				instant, _ := cmd.Flags().GetBool("instant-deploy")
				req.InstantDeploy = &instant
			}
			if cmd.Flags().Changed("docker-compose") {
				compose, _ := cmd.Flags().GetString("docker-compose")
				req.DockerCompose = &compose
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			svc := service.NewService(client)
			result, err := svc.Create(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(result)
		},
	}

	// List types flag
	cmd.Flags().Bool("list-types", false, "List all available service types")

	// Required flags
	cmd.Flags().String("server-uuid", "", "Server UUID (required)")
	cmd.Flags().String("project-uuid", "", "Project UUID (required)")
	cmd.Flags().String("environment-name", "", "Environment name")
	cmd.Flags().String("environment-uuid", "", "Environment UUID")

	// Optional flags
	cmd.Flags().String("name", "", "Service name")
	cmd.Flags().String("description", "", "Service description")
	cmd.Flags().String("destination-uuid", "", "Destination UUID if server has multiple destinations")
	cmd.Flags().Bool("instant-deploy", false, "Deploy immediately after creation")
	cmd.Flags().String("docker-compose", "", "Custom Docker Compose content (for advanced customization)")

	// Add completion for service type positional argument
	cmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return validServiceTypes, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return cmd
}

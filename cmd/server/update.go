package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewUpdateCommand creates the server update command.
func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <uuid>",
		Short: "Update a server",
		Long: `Update server configuration by UUID. Only specified fields are updated.

Example:
  coolify server update <uuid> --name "prod-1" --ip 10.0.0.5 --port 22`,
		Args: cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			req := models.ServerUpdateRequest{}
			hasUpdates := false

			if cmd.Flags().Changed("name") {
				v, _ := cmd.Flags().GetString("name")
				req.Name = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("description") {
				v, _ := cmd.Flags().GetString("description")
				req.Description = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("ip") {
				v, _ := cmd.Flags().GetString("ip")
				req.IP = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("port") {
				v, _ := cmd.Flags().GetInt("port")
				req.Port = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("user") {
				v, _ := cmd.Flags().GetString("user")
				req.User = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("private-key-uuid") {
				v, _ := cmd.Flags().GetString("private-key-uuid")
				req.PrivateKeyUUID = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("instant-validate") {
				v, _ := cmd.Flags().GetBool("instant-validate")
				req.InstantValidate = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("is-build-server") {
				v, _ := cmd.Flags().GetBool("is-build-server")
				req.IsBuildServer = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("proxy-type") {
				v, _ := cmd.Flags().GetString("proxy-type")
				req.ProxyType = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("connection-timeout") {
				v, _ := cmd.Flags().GetInt("connection-timeout")
				req.ConnectionTimeout = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("is-terminal-enabled") {
				v, _ := cmd.Flags().GetBool("is-terminal-enabled")
				req.IsTerminalEnabled = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("concurrent-builds") {
				v, _ := cmd.Flags().GetInt("concurrent-builds")
				req.ConcurrentBuilds = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("dynamic-timeout") {
				v, _ := cmd.Flags().GetInt("dynamic-timeout")
				req.DynamicTimeout = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("deployment-queue-limit") {
				v, _ := cmd.Flags().GetInt("deployment-queue-limit")
				req.DeploymentQueueLimit = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("disk-usage-notification-threshold") {
				v, _ := cmd.Flags().GetInt("disk-usage-notification-threshold")
				req.ServerDiskUsageNotificationThreshold = &v
				hasUpdates = true
			}
			if cmd.Flags().Changed("disk-usage-check-frequency") {
				v, _ := cmd.Flags().GetString("disk-usage-check-frequency")
				req.ServerDiskUsageCheckFrequency = &v
				hasUpdates = true
			}

			if !hasUpdates {
				return fmt.Errorf("no fields to update. Use --help to see available flags")
			}

			serverSvc := service.NewServerService(client)
			resp, err := serverSvc.Update(ctx, uuid, req)
			if err != nil {
				return fmt.Errorf("failed to update server: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}
			return formatter.Format(resp)
		},
	}

	cmd.Flags().String("name", "", "Server name")
	cmd.Flags().String("description", "", "Server description")
	cmd.Flags().String("ip", "", "Server IP address")
	cmd.Flags().Int("port", 22, "SSH port")
	cmd.Flags().String("user", "", "SSH user")
	cmd.Flags().String("private-key-uuid", "", "Private key UUID")
	cmd.Flags().Bool("instant-validate", false, "Validate server after update")
	cmd.Flags().Bool("is-build-server", false, "Mark as build server")
	cmd.Flags().String("proxy-type", "", "Proxy type (e.g. TRAEFIK, CADDY, NGINX)")
	cmd.Flags().Int("connection-timeout", 0, "SSH connection timeout in seconds")
	cmd.Flags().Bool("is-terminal-enabled", false, "Enable terminal access")
	cmd.Flags().Int("concurrent-builds", 0, "Max concurrent builds")
	cmd.Flags().Int("dynamic-timeout", 0, "Dynamic timeout")
	cmd.Flags().Int("deployment-queue-limit", 0, "Deployment queue limit")
	cmd.Flags().Int("disk-usage-notification-threshold", 0, "Disk usage notification threshold percent")
	cmd.Flags().String("disk-usage-check-frequency", "", "Disk usage check frequency (cron)")

	return cmd
}

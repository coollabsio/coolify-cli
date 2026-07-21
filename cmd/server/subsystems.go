package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func newSubsystemSvc(cmd *cobra.Command) (*service.ServerSubsystemService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewServerSubsystemService(client), nil
}

func printJSON(cmd *cobra.Command, value any) error {
	formatName, _ := cmd.Flags().GetString("format")
	if formatName == "table" {
		formatName = "pretty"
	}
	formatter, err := output.NewFormatter(formatName, output.Options{})
	if err != nil {
		return err
	}
	return formatter.Format(value)
}

// NewDockerCleanupCommand manages server docker cleanup settings.
func NewDockerCleanupCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "docker-cleanup", Short: "Manage server Docker cleanup settings"}
	cmd.AddCommand(
		&cobra.Command{Use: "get <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Get docker cleanup settings", RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newSubsystemSvc(cmd)
			if err != nil {
				return err
			}
			out, err := s.GetDockerCleanup(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return printJSON(cmd, out)
		}},
	)
	update := &cobra.Command{Use: "update <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Update docker cleanup settings", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		req := models.DockerCleanupUpdateRequest{}
		if cmd.Flags().Changed("frequency") {
			v, _ := cmd.Flags().GetString("frequency")
			req.DockerCleanupFrequency = &v
		}
		if cmd.Flags().Changed("threshold") {
			v, _ := cmd.Flags().GetInt("threshold")
			req.DockerCleanupThreshold = &v
		}
		if cmd.Flags().Changed("force") {
			v, _ := cmd.Flags().GetBool("force")
			req.ForceDockerCleanup = &v
		}
		if cmd.Flags().Changed("delete-unused-volumes") {
			v, _ := cmd.Flags().GetBool("delete-unused-volumes")
			req.DeleteUnusedVolumes = &v
		}
		if cmd.Flags().Changed("delete-unused-networks") {
			v, _ := cmd.Flags().GetBool("delete-unused-networks")
			req.DeleteUnusedNetworks = &v
		}
		if cmd.Flags().Changed("disable-image-retention") {
			v, _ := cmd.Flags().GetBool("disable-image-retention")
			req.DisableApplicationImageRetention = &v
		}
		out, err := s.UpdateDockerCleanup(cmd.Context(), args[0], req)
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}}
	update.Flags().String("frequency", "", "Cron frequency")
	update.Flags().Int("threshold", 0, "Disk usage threshold percent")
	update.Flags().Bool("force", false, "Force cleanup")
	update.Flags().Bool("delete-unused-volumes", false, "Delete unused volumes")
	update.Flags().Bool("delete-unused-networks", false, "Delete unused networks")
	update.Flags().Bool("disable-image-retention", false, "Disable application image retention")
	cmd.AddCommand(update)
	cmd.AddCommand(&cobra.Command{Use: "run <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Run docker cleanup now", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.RunDockerCleanup(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}})
	cmd.AddCommand(&cobra.Command{Use: "executions <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "List docker cleanup executions", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.ListDockerCleanupExecutions(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}})
	return cmd
}

// NewLogDrainsCommand manages server log drains.
func NewLogDrainsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "log-drains", Short: "Manage server log drain settings"}
	cmd.AddCommand(&cobra.Command{Use: "get <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Get log drain settings", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.GetLogDrains(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}})
	update := &cobra.Command{Use: "update <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Update log drain settings via JSON", RunE: func(cmd *cobra.Command, args []string) error {
		jsonBody, _ := cmd.Flags().GetString("json")
		if jsonBody == "" {
			return fmt.Errorf("--json is required")
		}
		var body map[string]any
		if err := json.Unmarshal([]byte(jsonBody), &body); err != nil {
			return fmt.Errorf("invalid --json: %w", err)
		}
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.UpdateLogDrains(cmd.Context(), args[0], body)
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}}
	update.Flags().String("json", "", "JSON fields to patch")
	cmd.AddCommand(update)
	return cmd
}

// NewSentinelCommand manages Sentinel settings.
func NewSentinelCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "sentinel", Short: "Manage server Sentinel/metrics settings"}
	cmd.AddCommand(&cobra.Command{Use: "get <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Get Sentinel settings", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.GetSentinel(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}})
	update := &cobra.Command{Use: "update <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Update Sentinel settings via JSON", RunE: func(cmd *cobra.Command, args []string) error {
		jsonBody, _ := cmd.Flags().GetString("json")
		if jsonBody == "" {
			return fmt.Errorf("--json is required")
		}
		var body map[string]any
		if err := json.Unmarshal([]byte(jsonBody), &body); err != nil {
			return fmt.Errorf("invalid --json: %w", err)
		}
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.UpdateSentinel(cmd.Context(), args[0], body)
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}}
	update.Flags().String("json", "", "JSON fields to patch")
	cmd.AddCommand(update)
	return cmd
}

// NewCloudflareTunnelCommand manages Cloudflare tunnel flag/actions.
func NewCloudflareTunnelCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "cloudflare-tunnel", Aliases: []string{"cf-tunnel"}, Short: "Manage Cloudflare Tunnel settings"}
	cmd.AddCommand(&cobra.Command{Use: "get <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Get Cloudflare tunnel state", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.GetCloudflareTunnel(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}})
	cmd.AddCommand(&cobra.Command{Use: "enable <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Enable Cloudflare tunnel flag", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.EnableCloudflareTunnel(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}})
	cmd.AddCommand(&cobra.Command{Use: "disable <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Disable Cloudflare tunnel flag", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.DisableCloudflareTunnel(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}})
	return cmd
}

// NewProxyCommand manages reverse proxy settings.
func NewProxyCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "proxy", Short: "Manage server reverse proxy"}
	cmd.AddCommand(&cobra.Command{Use: "get <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Get proxy settings", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.GetProxy(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}})
	update := &cobra.Command{Use: "update <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Update proxy settings", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		req := models.ServerProxyUpdateRequest{}
		if cmd.Flags().Changed("redirect-enabled") {
			v, _ := cmd.Flags().GetBool("redirect-enabled")
			req.RedirectEnabled = &v
		}
		if cmd.Flags().Changed("redirect-url") {
			v, _ := cmd.Flags().GetString("redirect-url")
			req.RedirectURL = &v
		}
		if cmd.Flags().Changed("generate-exact-labels") {
			v, _ := cmd.Flags().GetBool("generate-exact-labels")
			req.GenerateExactLabels = &v
		}
		if cmd.Flags().Changed("proxy-type") {
			v, _ := cmd.Flags().GetString("proxy-type")
			req.ProxyType = &v
		}
		out, err := s.UpdateProxy(cmd.Context(), args[0], req)
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}}
	update.Flags().Bool("redirect-enabled", true, "Enable proxy redirect")
	update.Flags().String("redirect-url", "", "Redirect URL")
	update.Flags().Bool("generate-exact-labels", false, "Generate exact labels")
	update.Flags().String("proxy-type", "", "Proxy type (e.g. TRAEFIK)")
	cmd.AddCommand(update)
	saveCfg := &cobra.Command{Use: "set-config <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Save raw proxy configuration", RunE: func(cmd *cobra.Command, args []string) error {
		configStr, _ := cmd.Flags().GetString("configuration")
		configFile, _ := cmd.Flags().GetString("configuration-file")
		if configFile != "" {
			b, err := os.ReadFile(configFile)
			if err != nil {
				return err
			}
			configStr = string(b)
		}
		if configStr == "" {
			return fmt.Errorf("--configuration or --configuration-file is required")
		}
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.SaveProxyConfiguration(cmd.Context(), args[0], configStr)
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}}
	saveCfg.Flags().String("configuration", "", "Raw proxy configuration")
	saveCfg.Flags().String("configuration-file", "", "Read configuration from file")
	cmd.AddCommand(saveCfg)
	cmd.AddCommand(&cobra.Command{Use: "restart <server_uuid>", Args: cli.ExactArgs(1, "<server_uuid>"), Short: "Restart proxy", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newSubsystemSvc(cmd)
		if err != nil {
			return err
		}
		out, err := s.RestartProxy(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		return printJSON(cmd, out)
	}})
	return cmd
}

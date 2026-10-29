package integrationtoken

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewListCommand lists the integration tokens of the team (GET /security/integration-tokens)
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the integration tokens of the team",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			tokens, err := svc.ListIntegrationTokens(cmd.Context())
			if err != nil {
				return err
			}
			return format(cmd, tokens)
		},
	}
}

// NewGetCommand gets an integration token (GET /security/integration-tokens/{uuid})
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <token_uuid>",
		Short: "Get an integration token",
		Args:  cli.ExactArgs(1, "<token_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			token, err := svc.GetIntegrationToken(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return format(cmd, token)
		},
	}
}

// NewUpdateCommand updates an integration token (PATCH /security/integration-tokens/{uuid})
func NewUpdateCommand() *cobra.Command {
	var name, metadata string
	var capabilities []string
	var rotate, tokenStdin, automaticDNS, proxiedDNS bool
	cmd := &cobra.Command{
		Use:   "update <token_uuid>",
		Short: "Rename an integration token, rotate its secret, or change its settings",
		Long: `Change an integration token. Options you do not set keep their value. Coolify checks the token
with the provider again when the secret, the capabilities, or the settings change.

Use --rotate to enter a new secret (with --provider-token-stdin, the ` + TokenEnvVar + `
environment variable, or a hidden prompt). A new secret is required when base_url, client_id
(Infisical), or namespace (Vault) changes.`,
		Example: `  coolify integration-token update <token_uuid> --automatic-dns=false
  echo "$NEW_TOKEN" | coolify integration-token update <token_uuid> --rotate --provider-token-stdin`,
		Args: cli.ExactArgs(1, "<token_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			meta, err := service.ParseJSONObject("metadata", metadata)
			if err != nil {
				return err
			}
			req := models.IntegrationTokenUpdateRequest{Metadata: meta, Capabilities: capabilities}
			if cmd.Flags().Changed("name") {
				trimmed := strings.TrimSpace(name)
				req.Name = &trimmed
			}
			if cmd.Flags().Changed("automatic-dns") {
				req.AutomaticDNS = &automaticDNS
			}
			if cmd.Flags().Changed("proxied-dns") {
				req.ProxiedDNS = &proxiedDNS
			}
			if tokenStdin && !rotate {
				return fmt.Errorf("--provider-token-stdin needs --rotate")
			}
			if rotate {
				if req.Token, err = readProviderToken(cmd, tokenStdin); err != nil {
					return err
				}
			}
			if req.Name == nil && req.Token == "" && req.Capabilities == nil && req.Metadata == nil && req.AutomaticDNS == nil && req.ProxiedDNS == nil {
				return fmt.Errorf("nothing to update: set --name, --rotate, --capabilities, --metadata, --automatic-dns, or --proxied-dns")
			}
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			token, err := svc.UpdateIntegrationToken(cmd.Context(), args[0], req)
			if err != nil {
				return err
			}
			return format(cmd, token)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New token name")
	cmd.Flags().StringVar(&metadata, "metadata", "", "Secret managers: provider settings to change as a JSON object")
	cmd.Flags().StringSliceVar(&capabilities, "capabilities", nil, "Capabilities: dns, tunnel (cloudflare) or secrets")
	cmd.Flags().BoolVar(&rotate, "rotate", false, "Enter a new token secret")
	cmd.Flags().BoolVar(&tokenStdin, "provider-token-stdin", false, "Read the new secret from stdin (with --rotate)")
	cmd.Flags().BoolVar(&automaticDNS, "automatic-dns", true, "Cloudflare: create DNS records when a domain is added")
	cmd.Flags().BoolVar(&proxiedDNS, "proxied-dns", false, "Cloudflare: create new records proxied")
	return cmd
}

// NewDeleteCommand deletes an integration token (DELETE /security/integration-tokens/{uuid})
func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <token_uuid>",
		Short: "Delete an integration token that no resource uses and that manages no DNS records",
		Args:  cli.ExactArgs(1, "<token_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if force, _ := cmd.Flags().GetBool("force"); !force {
				ok, err := cli.Confirm(cmd, "Delete integration token "+args[0]+"?")
				if err != nil || !ok {
					return err
				}
			}
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			if err := svc.DeleteIntegrationToken(cmd.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Integration token deleted.")
			return nil
		},
	}
	cmd.Flags().Bool("force", false, "Skip the confirmation")
	return cmd
}

// NewSyncZonesCommand reads the zones of a Cloudflare token again (POST /security/integration-tokens/{uuid}/sync-zones)
func NewSyncZonesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync-zones <token_uuid>",
		Short: "Read the DNS zones of a Cloudflare token again",
		Args:  cli.ExactArgs(1, "<token_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			resp, err := svc.SyncZones(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return format(cmd, resp)
		},
	}
}

// NewZonesCommand lists the DNS zones of the team's Cloudflare tokens (GET /dns/zones)
func NewZonesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "zones",
		Short: "List the DNS zones that the Cloudflare tokens of the team can access",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			zones, err := svc.ListDNSZones(cmd.Context())
			if err != nil {
				return err
			}
			return format(cmd, zones)
		},
	}
}

func newService(cmd *cobra.Command) (*service.SecretManagerService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewSecretManagerService(client), nil
}

func format(cmd *cobra.Command, value any) error {
	formatName, _ := cmd.Flags().GetString("format")
	formatter, err := output.NewFormatter(formatName, output.Options{Writer: cmd.OutOrStdout()})
	if err != nil {
		return err
	}
	return formatter.Format(value)
}

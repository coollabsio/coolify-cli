package integrationtoken

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// TokenEnvVar is the environment variable the provider token can be read from.
const TokenEnvVar = "COOLIFY_INTEGRATION_TOKEN"

// maxTokenBytes caps how much is read from stdin for the provider token.
const maxTokenBytes = 64 * 1024

// terminalFD returns the file descriptor of in when it is an interactive terminal.
// It is a variable so tests can simulate a TTY.
var terminalFD = func(in io.Reader) (int, bool) {
	file, ok := in.(*os.File)
	if !ok {
		return 0, false
	}
	fd := int(file.Fd())
	return fd, term.IsTerminal(fd)
}

// readHiddenToken reads a secret from a terminal without echo.
// It is a variable so tests can simulate a TTY.
var readHiddenToken = term.ReadPassword

// NewIntegrationTokenCommand creates the integration-token parent command
func NewIntegrationTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "integration-token",
		Aliases: []string{"integration-tokens"},
		Short:   "Manage Cloudflare and secret manager integration tokens",
		Long: `Manage integration tokens for Cloudflare (DNS, Tunnel) and external secret managers (Doppler, Infisical, Vault).
The token secret is never shown.

Resources use a secret manager token as their secret source, see:
  coolify app secret-manager set <app_uuid> --integration-token <token_uuid>
Cloudflare tokens manage the DNS records of domains, see:
  coolify app dns create <app_uuid> --hostname app.example.com`,
	}
	cmd.AddCommand(NewListCommand(), NewGetCommand(), NewCreateCommand(), NewUpdateCommand(), NewDeleteCommand(), NewSyncZonesCommand(), NewZonesCommand())
	return cmd
}

// NewCreateCommand creates an integration token (POST /security/integration-tokens)
func NewCreateCommand() *cobra.Command {
	var provider, name, metadata string
	var capabilities []string
	var tokenStdin, automaticDNS, proxiedDNS bool
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create and validate an integration token",
		Long: `Create a Cloudflare, Doppler, Infisical or Vault integration token. Coolify validates the
token with the provider before saving it. A Cloudflare token also gets its DNS zones.

The provider token is never accepted as a flag value and never printed. Pass it
with --provider-token-stdin, the ` + TokenEnvVar + ` environment variable, or
enter it at the hidden prompt when running in a terminal.

Cloudflare options:
  --capabilities dns,tunnel   (default dns; tunnel needs dns)
  --automatic-dns=false       do not create DNS records when a domain is added (default true)
  --proxied-dns               create new records proxied (orange cloud)

Provider specific --metadata (JSON object):
  doppler    none (token must start with dp.st. or dp.sa.)
  infisical  {"base_url":"https://app.infisical.com","client_id":"<machine identity client id>"}
  vault      {"base_url":"https://vault.example.com","namespace":"<optional>"}`,
		Example: `  echo "$CF_TOKEN" | coolify integration-token create --provider cloudflare --name cloudflare --provider-token-stdin
  echo "$DOPPLER_TOKEN" | coolify integration-token create --provider doppler --name prod --provider-token-stdin
  COOLIFY_INTEGRATION_TOKEN=hvs.xxx coolify integration-token create --provider vault --name vault \
    --metadata '{"base_url":"https://vault.example.com"}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			meta, err := service.ParseJSONObject("metadata", metadata)
			if err != nil {
				return err
			}

			req := models.IntegrationTokenCreateRequest{
				Provider:     strings.ToLower(strings.TrimSpace(provider)),
				Name:         strings.TrimSpace(name),
				Metadata:     meta,
				Capabilities: capabilities,
			}
			if cmd.Flags().Changed("automatic-dns") {
				req.AutomaticDNS = &automaticDNS
			}
			if cmd.Flags().Changed("proxied-dns") {
				req.ProxiedDNS = &proxiedDNS
			}
			// Validate everything except the token before prompting for it.
			if err := service.ValidateIntegrationTokenOptions(req); err != nil {
				return err
			}

			token, err := readProviderToken(cmd, tokenStdin)
			if err != nil {
				return err
			}
			req.Token = token

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			resp, err := service.NewSecretManagerService(client).CreateIntegrationToken(cmd.Context(), req)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{Writer: cmd.OutOrStdout()})
			if err != nil {
				return err
			}
			return formatter.Format(resp)
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "Provider: "+strings.Join(models.IntegrationTokenProviders, ", ")+" (required)")
	cmd.Flags().StringVar(&name, "name", "", "Friendly token name (required)")
	cmd.Flags().StringVar(&metadata, "metadata", "", "Provider metadata as a JSON object (required for infisical and vault)")
	cmd.Flags().BoolVar(&tokenStdin, "provider-token-stdin", false, "Read the provider token from stdin")
	cmd.Flags().StringSliceVar(&capabilities, "capabilities", nil, "Capabilities: dns, tunnel (cloudflare) or secrets")
	cmd.Flags().BoolVar(&automaticDNS, "automatic-dns", true, "Cloudflare: create DNS records when a domain is added")
	cmd.Flags().BoolVar(&proxiedDNS, "proxied-dns", false, "Cloudflare: create new records proxied")
	return cmd
}

// readProviderToken reads the provider token from stdin, the environment, or a hidden prompt.
func readProviderToken(cmd *cobra.Command, fromStdin bool) (string, error) {
	in := cmd.InOrStdin()

	var token string
	switch {
	case fromStdin:
		data, err := io.ReadAll(io.LimitReader(in, maxTokenBytes+1))
		if err != nil {
			return "", fmt.Errorf("failed to read provider token from stdin: %w", err)
		}
		if len(data) > maxTokenBytes {
			return "", fmt.Errorf("provider token must not be longer than %d bytes", maxTokenBytes)
		}
		token = string(data)
	case os.Getenv(TokenEnvVar) != "":
		token = os.Getenv(TokenEnvVar)
	default:
		fd, ok := terminalFD(in)
		if !ok {
			return "", errors.New("no provider token given: pipe it with --provider-token-stdin or set " + TokenEnvVar + " (the token cannot be passed as a flag)")
		}
		fmt.Fprint(cmd.ErrOrStderr(), "Provider token: ")
		data, err := readHiddenToken(fd)
		fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			return "", fmt.Errorf("failed to read provider token: %w", err)
		}
		token = string(data)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("provider token must not be empty")
	}
	return token, nil
}

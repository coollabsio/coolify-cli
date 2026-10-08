package secretmanager

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret-manager",
		Aliases: []string{"secret-managers"},
		Short:   "Manage secret manager integrations",
	}
	cmd.AddCommand(newCreateTokenCommand())
	return cmd
}

func newCreateTokenCommand() *cobra.Command {
	var provider, name, providerToken, baseURL, clientID, namespace string
	cmd := &cobra.Command{
		Use:   "create-token",
		Short: "Create a Doppler, Infisical, or Vault integration token",
		RunE: func(cmd *cobra.Command, _ []string) error {
			provider = strings.ToLower(provider)
			if provider != "doppler" && provider != "infisical" && provider != "vault" {
				return fmt.Errorf("--provider must be doppler, infisical, or vault")
			}
			if name == "" || providerToken == "" {
				return fmt.Errorf("--name and --provider-token are required")
			}
			if provider == "infisical" && (baseURL == "" || clientID == "") {
				return fmt.Errorf("--base-url and --client-id are required for Infisical")
			}
			if provider == "vault" && baseURL == "" {
				return fmt.Errorf("--base-url is required for Vault")
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			response, err := service.NewSecretManagerService(client).CreateToken(cmd.Context(), models.SecretManagerTokenCreateRequest{
				Provider: provider,
				Name:     name,
				Token:    providerToken,
				Metadata: models.SecretManagerMetadata{BaseURL: baseURL, ClientID: clientID, Namespace: namespace},
			})
			if err != nil {
				return fmt.Errorf("failed to create secret manager token: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(response)
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "", "Secret manager provider: doppler, infisical, or vault")
	cmd.Flags().StringVar(&name, "name", "", "Friendly token name")
	cmd.Flags().StringVar(&providerToken, "provider-token", "", "Provider token or client secret (sensitive; never included in output)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "Infisical or Vault base URL")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Infisical machine identity client ID")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Vault namespace")
	return cmd
}

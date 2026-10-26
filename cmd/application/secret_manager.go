package application

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// secretManagerRow is the table view of an application's secret manager configuration.
type secretManagerRow struct {
	IntegrationTokenUUID string `json:"integration_token_uuid"`
	Provider             string `json:"provider"`
	Settings             string `json:"settings"`
}

// NewSecretManagerCommand creates the app secret-manager parent command
func NewSecretManagerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret-manager",
		Aliases: []string{"secret-managers"},
		Short:   "Configure the external secret manager of an application",
		Long: `Configure the external secret manager (Doppler, Infisical, Vault) an application reads secrets from.
Create the integration token first with: coolify integration-token create`,
	}
	cmd.AddCommand(NewSecretManagerSetCommand())
	return cmd
}

// NewSecretManagerSetCommand sets the secret manager of an application (PATCH /applications/{uuid}/secret-manager)
func NewSecretManagerSetCommand() *cobra.Command {
	var tokenUUID, settings string
	cmd := &cobra.Command{
		Use:   "set <app_uuid>",
		Short: "Set the secret manager source of an application",
		Long: `Set the secret manager integration token and provider settings an application uses.

Provider specific --settings (JSON object):
  doppler    {"project":"...","config":"..."} (required for service account tokens)
  infisical  {"project_id":"...","environment":"...","secret_path":"/optional"}
  vault      {"mount":"...","path":"..."}`,
		Example: `  coolify app secret-manager set <app_uuid> --integration-token <token_uuid>
  coolify app secret-manager set <app_uuid> --integration-token <token_uuid> --settings '{"mount":"secret","path":"apps/web"}'`,
		Args: cli.ExactArgs(1, "<app_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			tokenUUID = strings.TrimSpace(tokenUUID)
			if tokenUUID == "" {
				return fmt.Errorf("--integration-token is required")
			}
			parsedSettings, err := service.ParseJSONObject("settings", settings)
			if err != nil {
				return err
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			resp, err := service.NewSecretManagerService(client).ConfigureApplication(cmd.Context(), args[0], models.ApplicationSecretManagerRequest{
				IntegrationTokenUUID: tokenUUID,
				Settings:             parsedSettings,
			})
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{Writer: cmd.OutOrStdout()})
			if err != nil {
				return err
			}
			if format != output.FormatTable {
				return formatter.Format(resp)
			}
			return formatter.Format(toSecretManagerRow(resp))
		},
	}
	cmd.Flags().StringVar(&tokenUUID, "integration-token", "", "UUID of the secret manager integration token (required)")
	cmd.Flags().StringVar(&settings, "settings", "", "Provider settings as a JSON object")
	return cmd
}

func toSecretManagerRow(resp *models.ApplicationSecretManager) secretManagerRow {
	row := secretManagerRow{IntegrationTokenUUID: resp.IntegrationTokenUUID, Provider: resp.Provider}
	if len(resp.Settings) > 0 {
		if encoded, err := json.Marshal(resp.Settings); err == nil {
			row.Settings = string(encoded)
		}
	}
	return row
}

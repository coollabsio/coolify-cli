package application

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewSecretManagerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret-manager",
		Short: "Configure an application's secret manager",
	}
	cmd.AddCommand(newSetSecretManagerCommand())
	return cmd
}

func newSetSecretManagerCommand() *cobra.Command {
	var tokenUUID string
	settings := map[string]string{}
	cmd := &cobra.Command{
		Use:   "set <uuid>",
		Short: "Configure the secret manager used by an application",
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenUUID == "" {
				return fmt.Errorf("--integration-token-uuid is required")
			}
			for _, key := range []string{"project", "config", "project-id", "environment", "secret-path", "mount", "path"} {
				value, _ := cmd.Flags().GetString(key)
				if value != "" {
					settings[flagToSetting(key)] = value
				}
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			response, err := service.NewSecretManagerService(client).ConfigureApplication(cmd.Context(), args[0], models.ApplicationSecretManagerRequest{
				IntegrationTokenUUID: tokenUUID,
				Settings:             settings,
			})
			if err != nil {
				return fmt.Errorf("failed to configure application secret manager: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(response)
		},
	}
	cmd.Flags().StringVar(&tokenUUID, "integration-token-uuid", "", "Secret manager integration token UUID")
	cmd.Flags().String("project", "", "Doppler project (service account tokens)")
	cmd.Flags().String("config", "", "Doppler config (service account tokens)")
	cmd.Flags().String("project-id", "", "Infisical project ID")
	cmd.Flags().String("environment", "", "Infisical environment slug")
	cmd.Flags().String("secret-path", "", "Infisical secret path")
	cmd.Flags().String("mount", "", "Vault secrets engine mount")
	cmd.Flags().String("path", "", "Vault secret path")
	return cmd
}

func flagToSetting(flag string) string {
	if flag == "project-id" {
		return "project_id"
	}
	if flag == "secret-path" {
		return "secret_path"
	}
	return flag
}

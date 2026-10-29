// Package secretmanager has the secret-manager commands of applications, services, and databases.
package secretmanager

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

// Resource describes the resource type a secret-manager command tree works on.
type Resource struct {
	// APIType is the first path segment of the API route: applications, services, or databases.
	APIType string
	// Command is the CLI command of the resource (app, service, database), used in examples.
	Command string
	// Arg is the usage name of the UUID argument, for example "<app_uuid>".
	Arg string
}

// sourceRow is the table view of a secret manager source.
type sourceRow struct {
	Linked               bool   `json:"linked"`
	IntegrationTokenUUID string `json:"integration_token_uuid"`
	Provider             string `json:"provider"`
	Settings             string `json:"settings"`
	AvailableTokens      string `json:"available_tokens"`
}

// configRow is the table view of a secret manager configuration.
type configRow struct {
	IntegrationTokenUUID string `json:"integration_token_uuid"`
	Provider             string `json:"provider"`
	Settings             string `json:"settings"`
}

// NewCommand creates the secret-manager parent command for the resource type.
func NewCommand(r Resource) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secret-manager",
		Aliases: []string{"secret-managers"},
		Short:   "Configure the external secret manager of a resource",
		Long: `Configure the external secret manager (Doppler, Infisical, Vault) a resource reads secrets from,
and add {{vault.KEY}} reference variables. Secret values are never shown.
Create the integration token first with: coolify integration-token create`,
	}
	cmd.AddCommand(newGetCommand(r), newSetCommand(r), newRemoveCommand(r), newKeysCommand(r), newAddReferenceCommand(r), newImportCommand(r))
	return cmd
}

func newGetCommand(r Resource) *cobra.Command {
	return &cobra.Command{
		Use:   "get " + r.Arg,
		Short: "Show the secret manager source and the tokens that can be linked",
		Args:  cli.ExactArgs(1, r.Arg),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			resp, err := svc.GetSource(cmd.Context(), r.APIType, args[0])
			if err != nil {
				return err
			}
			row := sourceRow{Linked: resp.Linked, Settings: encodeSettings(resp.Settings)}
			if resp.IntegrationTokenUUID != nil {
				row.IntegrationTokenUUID = *resp.IntegrationTokenUUID
			}
			if resp.Provider != nil {
				row.Provider = *resp.Provider
			}
			tokens := make([]string, 0, len(resp.AvailableTokens))
			for _, token := range resp.AvailableTokens {
				tokens = append(tokens, fmt.Sprintf("%s (%s, %s)", token.UUID, token.Name, token.Provider))
			}
			row.AvailableTokens = strings.Join(tokens, ", ")
			return format(cmd, resp, row)
		},
	}
}

func newSetCommand(r Resource) *cobra.Command {
	var tokenUUID, settings string
	cmd := &cobra.Command{
		Use:   "set " + r.Arg,
		Short: "Set the secret manager source",
		Long: `Set the secret manager integration token and provider settings a resource uses.
References resolve at the next deployment.

Provider specific --settings (JSON object):
  doppler    {"project":"...","config":"..."} (required for service account tokens)
  infisical  {"project_id":"...","environment":"...","secret_path":"/optional"}
  vault      {"mount":"...","path":"..."}`,
		Example: fmt.Sprintf(`  coolify %[1]s secret-manager set %[2]s --integration-token <token_uuid>
  coolify %[1]s secret-manager set %[2]s --integration-token <token_uuid> --settings '{"mount":"secret","path":"apps/web"}'`, r.Command, r.Arg),
		Args: cli.ExactArgs(1, r.Arg),
		RunE: func(cmd *cobra.Command, args []string) error {
			tokenUUID = strings.TrimSpace(tokenUUID)
			if tokenUUID == "" {
				return fmt.Errorf("--integration-token is required")
			}
			parsedSettings, err := service.ParseJSONObject("settings", settings)
			if err != nil {
				return err
			}
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			resp, err := svc.Configure(cmd.Context(), r.APIType, args[0], models.ApplicationSecretManagerRequest{
				IntegrationTokenUUID: tokenUUID,
				Settings:             parsedSettings,
			})
			if err != nil {
				return err
			}
			return format(cmd, resp, configRow{IntegrationTokenUUID: resp.IntegrationTokenUUID, Provider: resp.Provider, Settings: encodeSettings(resp.Settings)})
		},
	}
	cmd.Flags().StringVar(&tokenUUID, "integration-token", "", "UUID of the secret manager integration token (required)")
	cmd.Flags().StringVar(&settings, "settings", "", "Provider settings as a JSON object")
	return cmd
}

func newRemoveCommand(r Resource) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove " + r.Arg,
		Short: "Remove the secret manager source",
		Long:  "Remove the secret manager source. Existing {{vault.*}} variables stay and fail the next deployment until you remove them.",
		Args:  cli.ExactArgs(1, r.Arg),
		RunE: func(cmd *cobra.Command, args []string) error {
			if force, _ := cmd.Flags().GetBool("force"); !force {
				ok, err := cli.Confirm(cmd, "Remove the secret manager source of "+args[0]+"? Existing {{vault.*}} variables fail the next deployment.")
				if err != nil || !ok {
					return err
				}
			}
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			if err := svc.RemoveSource(cmd.Context(), r.APIType, args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Secret manager source removed.")
			return nil
		},
	}
	cmd.Flags().Bool("force", false, "Skip the confirmation")
	return cmd
}

func newKeysCommand(r Resource) *cobra.Command {
	return &cobra.Command{
		Use:   "keys " + r.Arg,
		Short: "List the key names of the remote secrets (values are never shown)",
		Args:  cli.ExactArgs(1, r.Arg),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			resp, err := svc.Keys(cmd.Context(), r.APIType, args[0])
			if err != nil {
				return err
			}
			rows := make([]struct {
				Key string `json:"key"`
			}, len(resp.Keys))
			for i, key := range resp.Keys {
				rows[i].Key = key
			}
			return format(cmd, resp, rows)
		},
	}
}

func newAddReferenceCommand(r Resource) *cobra.Command {
	var key string
	cmd := &cobra.Command{
		Use:     "add-reference " + r.Arg,
		Short:   "Add a {{vault.KEY}} variable for one remote key",
		Example: fmt.Sprintf("  coolify %s secret-manager add-reference %s --key DATABASE_URL", r.Command, r.Arg),
		Args:    cli.ExactArgs(1, r.Arg),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			resp, err := svc.AddReference(cmd.Context(), r.APIType, args[0], strings.TrimSpace(key))
			if err != nil {
				return err
			}
			return format(cmd, resp, *resp)
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "Remote key name (required)")
	return cmd
}

func newImportCommand(r Resource) *cobra.Command {
	return &cobra.Command{
		Use:   "import " + r.Arg,
		Short: "Add a {{vault.KEY}} variable for every remote key that has no variable yet",
		Args:  cli.ExactArgs(1, r.Arg),
		RunE: func(cmd *cobra.Command, args []string) error {
			svc, err := newService(cmd)
			if err != nil {
				return err
			}
			resp, err := svc.Import(cmd.Context(), r.APIType, args[0])
			if err != nil {
				return err
			}
			return format(cmd, resp, struct {
				Imported string `json:"imported"`
				Skipped  string `json:"skipped"`
			}{strings.Join(resp.Imported, ", "), strings.Join(resp.Skipped, ", ")})
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

// format prints the API response for json and pretty, and the table view for table.
func format(cmd *cobra.Command, resp, table any) error {
	formatName, _ := cmd.Flags().GetString("format")
	formatter, err := output.NewFormatter(formatName, output.Options{Writer: cmd.OutOrStdout()})
	if err != nil {
		return err
	}
	if formatName != output.FormatTable {
		return formatter.Format(resp)
	}
	return formatter.Format(table)
}

func encodeSettings(settings map[string]any) string {
	if len(settings) == 0 {
		return ""
	}
	encoded, err := json.Marshal(settings)
	if err != nil {
		return ""
	}
	return string(encoded)
}

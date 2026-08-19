package settings

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewSettingsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "settings", Short: "Manage instance settings"}
	cmd.AddCommand(newEmailCommand())
	return cmd
}

func newEmailCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "email",
		Short: "Manage instance-wide SMTP and Resend settings",
		Long:  "Requires a root-team API token belonging to a root-team admin or owner.",
	}
	cmd.AddCommand(newEmailGetCommand(), newEmailUpdateCommand())
	return cmd
}

func emailService(cmd *cobra.Command) (*service.InstanceEmailSettingsService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewInstanceEmailSettingsService(client), nil
}

func newEmailGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get instance email settings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			svc, err := emailService(cmd)
			if err != nil {
				return err
			}
			settings, err := svc.Get(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get instance email settings: %w", err)
			}
			return formatSettings(cmd, settings)
		},
	}
}

func newEmailUpdateCommand() *cobra.Command {
	var jsonBody string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update instance email settings",
		Long:  `Pass settings as JSON via --json. Requires write:sensitive. Example: coolify settings email update --json '{"smtp_ehlo_domain":"coolify.example.com"}'`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if jsonBody == "" {
				return fmt.Errorf("--json is required")
			}
			var body map[string]any
			if err := json.Unmarshal([]byte(jsonBody), &body); err != nil {
				return fmt.Errorf("invalid --json: %w", err)
			}
			svc, err := emailService(cmd)
			if err != nil {
				return err
			}
			settings, err := svc.Update(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("failed to update instance email settings: %w", err)
			}
			return formatSettings(cmd, settings)
		},
	}
	cmd.Flags().StringVar(&jsonBody, "json", "", "JSON object of fields to update")
	return cmd
}

func formatSettings(cmd *cobra.Command, settings map[string]any) error {
	formatName, _ := cmd.Flags().GetString("format")
	if formatName == "table" {
		formatName = "pretty"
	}
	formatter, err := output.NewFormatter(formatName, output.Options{})
	if err != nil {
		return err
	}
	return formatter.Format(settings)
}

package cloudtoken

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewCloudTokenCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "cloud-token", Aliases: []string{"cloud-tokens"}, Short: "Manage cloud provider API tokens"}
	cmd.AddCommand(newList(), newGet(), newCreate(), newUpdate(), newDelete(), newValidate())
	return cmd
}

func svc(cmd *cobra.Command) (*service.CloudTokenService, error) {
	client, err := cli.GetAPIClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get API client: %w", err)
	}
	return service.NewCloudTokenService(client), nil
}

func format(cmd *cobra.Command, value any) error {
	formatName, _ := cmd.Flags().GetString("format")
	options := outputOptions(cmd)
	formatter, err := output.NewFormatter(formatName, options)
	if err != nil {
		return err
	}
	return formatter.Format(prepareOutput(value, options.ShowSensitive))
}

func outputOptions(cmd *cobra.Command) output.Options {
	showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

	return output.Options{ShowSensitive: showSensitive}
}

func prepareOutput(value any, showSensitive bool) any {
	if showSensitive {
		return value
	}

	switch typed := value.(type) {
	case *models.CloudToken:
		redacted := *typed
		if typed.Token != nil {
			overlay := output.SensitiveOverlay
			redacted.Token = &overlay
		}

		return &redacted
	case []models.CloudToken:
		redacted := append([]models.CloudToken(nil), typed...)
		for index := range redacted {
			if redacted[index].Token != nil {
				overlay := output.SensitiveOverlay
				redacted[index].Token = &overlay
			}
		}

		return redacted
	default:
		return value
	}
}

func newList() *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List cloud provider tokens", RunE: func(cmd *cobra.Command, _ []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		tokens, err := s.List(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to list cloud tokens: %w", err)
		}
		return format(cmd, tokens)
	}}
}
func newGet() *cobra.Command {
	return &cobra.Command{Use: "get <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Show a cloud provider token", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		token, err := s.Get(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to get cloud token: %w", err)
		}
		return format(cmd, token)
	}}
}

func newCreate() *cobra.Command {
	var provider, token, name string
	cmd := &cobra.Command{Use: "create", Short: "Create a cloud provider token", RunE: func(cmd *cobra.Command, _ []string) error {
		cloudProvider := models.CloudProvider(provider)
		if !cloudProvider.Valid() {
			return fmt.Errorf("--provider must be hetzner, digitalocean, or vultr")
		}
		if token == "" || name == "" {
			return fmt.Errorf("--name and --provider-token are required")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		response, err := s.Create(cmd.Context(), models.CloudTokenCreateRequest{Provider: cloudProvider, Token: token, Name: name})
		if err != nil {
			return fmt.Errorf("failed to create cloud token: %w", err)
		}
		return format(cmd, response)
	}}
	cmd.Flags().StringVar(&provider, "provider", "", "Cloud provider: hetzner, digitalocean, or vultr")
	cmd.Flags().StringVar(&name, "name", "", "Friendly token name")
	cmd.Flags().StringVar(&token, "provider-token", "", "Provider API token (sensitive; never included in output)")
	return cmd
}

func newUpdate() *cobra.Command {
	var name string
	cmd := &cobra.Command{Use: "update <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Rename a cloud provider token", RunE: func(cmd *cobra.Command, args []string) error {
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		response, err := s.Update(cmd.Context(), args[0], models.CloudTokenUpdateRequest{Name: name})
		if err != nil {
			return fmt.Errorf("failed to update cloud token: %w", err)
		}
		return format(cmd, response)
	}}
	cmd.Flags().StringVar(&name, "name", "", "New friendly token name")
	return cmd
}

func newDelete() *cobra.Command {
	return &cobra.Command{Use: "delete <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Delete a cloud provider token", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		if err := s.Delete(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("failed to delete cloud token: %w", err)
		}
		fmt.Println("Cloud provider token deleted successfully.")
		return nil
	}}
}
func newValidate() *cobra.Command {
	return &cobra.Command{Use: "validate <uuid>", Args: cli.ExactArgs(1, "<uuid>"), Short: "Validate a cloud provider token", RunE: func(cmd *cobra.Command, args []string) error {
		s, err := svc(cmd)
		if err != nil {
			return err
		}
		response, err := s.Validate(cmd.Context(), args[0])
		if err != nil {
			return fmt.Errorf("failed to validate cloud token: %w", err)
		}
		return format(cmd, response)
	}}
}

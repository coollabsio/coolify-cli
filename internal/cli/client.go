package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/config"
)

// GetAPIClient creates an API client from command flags or config
func GetAPIClient(cmd *cobra.Command) (*api.Client, error) {
	token, _ := cmd.Flags().GetString("token")
	debug, _ := cmd.Flags().GetBool("debug")

	instance, err := currentInstance(cmd)
	if err != nil {
		return nil, err
	}

	// Use token from flag if provided, otherwise use instance token
	if token == "" {
		token = instance.Token
	}

	return api.NewClient(instance.FQDN, token, api.WithDebug(debug)), nil
}

// CurrentContextName returns the name of the context the command talks to:
// the --context flag, or the default context.
func CurrentContextName(cmd *cobra.Command) (string, error) {
	instance, err := currentInstance(cmd)
	if err != nil {
		return "", err
	}
	return instance.Name, nil
}

func currentInstance(cmd *cobra.Command) (*config.Instance, error) {
	contextName, _ := cmd.Flags().GetString("context")

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if contextName != "" {
		instance, err := cfg.GetInstance(contextName)
		if err != nil {
			return nil, fmt.Errorf("context '%s' not found: %w", contextName, err)
		}
		return instance, nil
	}
	instance, err := cfg.GetDefault()
	if err != nil {
		return nil, fmt.Errorf("no default instance configured: %w", err)
	}
	return instance, nil
}

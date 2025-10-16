package cli

import (
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/config"
	"github.com/spf13/cobra"
)

// GetAPIClient creates an API client from command flags or config
func GetAPIClient(cmd *cobra.Command) (*api.Client, error) {
	// Get flags
	token, _ := cmd.Flags().GetString("token")
	contextName, _ := cmd.Flags().GetString("context")
	debug, _ := cmd.Flags().GetBool("debug")

	// Load config to get instance details
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var instance *config.Instance
	// Use context if specified, otherwise use default
	if contextName != "" {
		instance, err = cfg.GetInstance(contextName)
		if err != nil {
			return nil, fmt.Errorf("context '%s' not found: %w", contextName, err)
		}
	} else {
		instance, err = cfg.GetDefault()
		if err != nil {
			return nil, fmt.Errorf("no default instance configured: %w", err)
		}
	}

	// Get FQDN from instance
	fqdn := instance.FQDN

	// Use token from flag if provided, otherwise use instance token
	if token == "" {
		token = instance.Token
	}

	// Create client
	client := api.NewClient(fqdn, token, api.WithDebug(debug))

	return client, nil
}

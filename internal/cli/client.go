package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/config"
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

	// Precedence: --context flag > repo-local config file > global default
	localPath := ""
	if contextName == "" {
		contextName, localPath, err = config.LocalContext()
		if err != nil {
			return nil, err
		}
	}

	var instance *config.Instance
	if contextName != "" {
		instance, err = cfg.GetInstance(contextName)
		if err != nil {
			if localPath != "" {
				return nil, fmt.Errorf("context '%s' from %s not found in %s", contextName, localPath, config.Path())
			}
			return nil, fmt.Errorf("context '%s' not found: %w", contextName, err)
		}
	} else {
		instance, err = cfg.GetDefault()
		if err != nil {
			return nil, fmt.Errorf("no default instance configured: %w", err)
		}
	}

	if debug && localPath != "" {
		fmt.Fprintf(os.Stderr, "Using context '%s' from %s\n", contextName, localPath)
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

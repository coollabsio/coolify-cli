package context

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewVerifyCommand creates the verify command for contexts
func NewVerifyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify current context connection and authentication",
		Long: `Verify that the current context is properly configured by testing the connection
to the Coolify instance and validating the API token.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get API client - this will use the current default context
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Try to get version - this verifies both connection and authentication
			version, err := client.GetVersion(ctx)
			if err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}

			// If we got here, connection and authentication are working
			fmt.Printf("✓ Connection successful\n")
			fmt.Printf("✓ Authentication valid\n")
			fmt.Printf("✓ Coolify version: %s\n", version)

			return nil
		},
	}
}

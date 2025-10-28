package context

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
)

// NewVersionCommand creates the version command for contexts
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Get current context's Coolify version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			// Get API client
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Get version using API client
			version, err := client.GetVersion(ctx)
			if err != nil {
				return fmt.Errorf("failed to get version: %w", err)
			}

			fmt.Println(version)
			return nil
		},
	}
}

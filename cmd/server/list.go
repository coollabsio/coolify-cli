package server

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewListCommand creates the list command
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get API client
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Use service layer
			serverSvc := service.NewServerService(client)
			servers, err := serverSvc.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list servers: %w", err)
			}

			// Use output formatter
			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			if err := formatter.Format(servers); err != nil {
				return err
			}

			if !showSensitive && format == output.FormatTable {
				fmt.Println("\nNote: Use -s to show sensitive information.")
			}

			return nil
		},
	}
}

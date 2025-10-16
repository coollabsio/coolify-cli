package server

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the get command
func NewGetDomainsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "domains <uuid>",
		Aliases: []string{"domain"},
		Args:    cli.ExactArgs(1, "<uuid>"),
		Short:   "Get server domains by uuid",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get API client
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Use service layer
			serverSvc := service.NewServerService(client)
			uuid := args[0]

			// Get format flags
			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			domains, err := serverSvc.GetDomains(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to get server domains: %w", err)
			}

			// Use output formatter
			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			if err := formatter.Format(domains); err != nil {
				return err
			}

			if !showSensitive && format == output.FormatTable {
				fmt.Println("\nNote: Use -s to show sensitive information.")
			}

			return nil
		},
	}

	return cmd
}

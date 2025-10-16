package resources

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewListCommand returns the list projects command
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			resourceSvc := service.NewResourceService(client)
			resources, err := resourceSvc.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			return formatter.Format(resources)
		},
	}
}

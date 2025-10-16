package application

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all applications",
		Long:  `List all applications in your Coolify instance.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			appSvc := service.NewApplicationService(client)
			apps, err := appSvc.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list applications: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			// For JSON/pretty formats, return the full application structure
			if format != output.FormatTable {
				formatter, err := output.NewFormatter(format, output.Options{
					ShowSensitive: showSensitive,
				})
				if err != nil {
					return err
				}
				return formatter.Format(apps)
			}

			// For table format, convert to simplified rows
			var rows []models.ApplicationListItem
			for _, app := range apps {
				rows = append(rows, models.ApplicationListItem{
					UUID:        app.UUID,
					Name:        app.Name,
					Description: app.Description,
					Status:      app.Status,
					GitBranch:   app.GitBranch,
					FQDN:        app.FQDN,
				})
			}

			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			return formatter.Format(rows)
		},
	}
}

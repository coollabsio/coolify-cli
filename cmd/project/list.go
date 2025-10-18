package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// ListRow represents a project for list display (without environments)
type ListRow struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewListCommand returns the list projects command
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			projectSvc := service.NewProjectService(client)
			projects, err := projectSvc.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list projects: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			// For JSON/pretty formats, return the full project structure
			if format != output.FormatTable {
				formatter, err := output.NewFormatter(format, output.Options{
					ShowSensitive: showSensitive,
				})
				if err != nil {
					return err
				}
				return formatter.Format(projects)
			}

			// For table format, convert to simplified rows without environments
			var rows []ListRow
			for _, p := range projects {
				desc := ""
				if p.Description != nil {
					desc = *p.Description
				}
				rows = append(rows, ListRow{
					UUID:        p.UUID,
					Name:        p.Name,
					Description: desc,
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

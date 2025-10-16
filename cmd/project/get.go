package project

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// EnvironmentRow represents an environment for display
type EnvironmentRow struct {
	UUID            string `json:"environment_uuid"`
	EnvironmentName string `json:"environment_name"`
}

// NewGetCommand returns the get project command
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <uuid>",
		Short: "Get a project by uuid",
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			projectSvc := service.NewProjectService(client)
			project, err := projectSvc.Get(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to get project: %w", err)
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
				return formatter.Format(project)
			}

			// For table format, expand environments into separate rows
			var rows []EnvironmentRow

			// If the project has environments, expand them
			if project.Environments != nil && len(project.Environments) > 0 {
				for _, env := range project.Environments {
					rows = append(rows, EnvironmentRow{
						UUID:            env.UUID,
						EnvironmentName: env.Name,
					})
				}
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

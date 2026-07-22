package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewUpdateCommand returns the update project command
func NewUpdateCommand() *cobra.Command {
	var name, description string
	cmd := &cobra.Command{
		Use:   "update <uuid>",
		Short: "Update a project",
		Long: `Update a project's name and/or description.

Examples:
  coolify project update <uuid> --name "New Name"
  coolify project update <uuid> --description "New description"
  coolify project update <uuid> --name "New Name" --description "New description"`,
		Args: cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			req := models.ProjectUpdateRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = &name
			}
			if cmd.Flags().Changed("description") {
				req.Description = &description
			}
			if req.Name == nil && req.Description == nil {
				return fmt.Errorf("provide --name and/or --description")
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			project, err := service.NewProjectService(client).Update(ctx, uuid, req)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}
			return formatter.Format(project)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Project name")
	cmd.Flags().StringVar(&description, "description", "", "Project description")

	return cmd
}

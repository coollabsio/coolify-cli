package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewCreateCommand returns the create project command
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long: `Create a new project in Coolify.

Examples:
  coolify project create --name "My Project"
  coolify project create --name "My Project" --description "A description"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			req := &models.ProjectCreateRequest{
				Name: name,
			}

			if cmd.Flags().Changed("description") {
				desc, _ := cmd.Flags().GetString("description")
				req.Description = &desc
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			projectSvc := service.NewProjectService(client)
			project, err := projectSvc.Create(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create project: %w", err)
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

	cmd.Flags().String("name", "", "Project name (required)")
	cmd.Flags().String("description", "", "Project description")

	return cmd
}

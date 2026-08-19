package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewProjectCommand creates the project parent command with all subcommands
func NewProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"projects"},
		Short:   "Project related commands",
		Long:    `Manage Coolify projects and their environments.`,
	}

	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewUpdateCommand())
	cmd.AddCommand(NewDeleteCommand())
	cmd.AddCommand(NewEnvironmentsCommand())
	// Keep legacy top-level alias for environment update
	cmd.AddCommand(NewUpdateEnvironmentCommand())

	return cmd
}

// NewUpdateEnvironmentCommand patches a project environment name/description.
//
// Deprecated: use coolify project environments update instead.
func NewUpdateEnvironmentCommand() *cobra.Command {
	var name, description string
	cmd := &cobra.Command{
		Use:   "update-environment <project_uuid> <environment>",
		Args:  cli.ExactArgs(2, "<project_uuid> <environment>"),
		Short: "Update a project environment name or description",
		Long:  `Update a project environment. Prefer: coolify project environments update`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return err
			}
			req := models.EnvironmentUpdateRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = &name
			}
			if cmd.Flags().Changed("description") {
				req.Description = &description
			}
			if req.Name == nil && req.Description == nil {
				return fmt.Errorf("provide --name and/or --description")
			}
			out, err := service.NewProjectService(client).UpdateEnvironment(cmd.Context(), args[0], args[1], req)
			if err != nil {
				return err
			}
			formatName, _ := cmd.Flags().GetString("format")
			if formatName == "table" {
				formatName = "pretty"
			}
			formatter, err := output.NewFormatter(formatName, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(out)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Environment name")
	cmd.Flags().StringVar(&description, "description", "", "Environment description")
	return cmd
}

package project

import "github.com/spf13/cobra"

// NewProjectCommand creates the project parent command with all subcommands
func NewProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"projects"},
		Short:   "Project related commands",
		Long:    `Manage Coolify projects - list all projects or get details about a specific project.`,
	}

	// Add all project subcommands
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewGetCommand())

	return cmd
}

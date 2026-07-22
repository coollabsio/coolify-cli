package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewEnvironmentsCommand returns the nested environments parent command
func NewEnvironmentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environments",
		Aliases: []string{"environment", "env"},
		Short:   "Manage project environments",
		Long:    `List, create, update, and delete environments within a project.`,
	}

	cmd.AddCommand(NewEnvironmentsListCommand())
	cmd.AddCommand(NewEnvironmentsCreateCommand())
	cmd.AddCommand(NewEnvironmentsUpdateCommand())
	cmd.AddCommand(NewEnvironmentsDeleteCommand())

	return cmd
}

// NewEnvironmentsListCommand lists environments in a project
func NewEnvironmentsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list <project_uuid>",
		Short: "List environments in a project",
		Args:  cli.ExactArgs(1, "<project_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			environments, err := service.NewProjectService(client).ListEnvironments(cmd.Context(), args[0])
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
			return formatter.Format(environments)
		},
	}
}

// NewEnvironmentsCreateCommand creates an environment in a project
func NewEnvironmentsCreateCommand() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "create <project_uuid>",
		Short: "Create an environment in a project",
		Long: `Create a new environment in a project.

Examples:
  coolify project environments create <project_uuid> --name staging`,
		Args: cli.ExactArgs(1, "<project_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			resp, err := service.NewProjectService(client).CreateEnvironment(cmd.Context(), args[0], models.EnvironmentCreateRequest{
				Name: name,
			})
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
			return formatter.Format(resp)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Environment name (required)")
	return cmd
}

// NewEnvironmentsUpdateCommand updates an environment name/description
func NewEnvironmentsUpdateCommand() *cobra.Command {
	var name, description string
	cmd := &cobra.Command{
		Use:   "update <project_uuid> <environment>",
		Short: "Update a project environment",
		Long: `Update an environment's name and/or description.

Examples:
  coolify project environments update <project_uuid> staging --name production
  coolify project environments update <project_uuid> <env_uuid> --description "Prod env"`,
		Args: cli.ExactArgs(2, "<project_uuid> <environment>"),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
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

// NewEnvironmentsDeleteCommand deletes an environment
func NewEnvironmentsDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <project_uuid> <environment>",
		Short: "Delete a project environment",
		Long: `Delete an environment by name or UUID. The environment must be empty.

Examples:
  coolify project environments delete <project_uuid> staging
  coolify project environments delete <project_uuid> <env_uuid> --force`,
		Args: cli.ExactArgs(2, "<project_uuid> <environment>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectUUID := args[0]
			environment := args[1]

			force, _ := cmd.Flags().GetBool("force")
			if !force {
				var response string
				fmt.Printf("Are you sure you want to delete environment %s from project %s? This cannot be undone. (yes/no): ", environment, projectUUID)
				_, err := fmt.Scanln(&response)
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				if response != "yes" && response != "y" {
					fmt.Println("Delete cancelled.")
					return nil
				}
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := service.NewProjectService(client).DeleteEnvironment(cmd.Context(), projectUUID, environment); err != nil {
				return err
			}

			fmt.Printf("Environment %s deleted successfully.\n", environment)
			return nil
		},
	}
	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	return cmd
}

package cmd

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// EnvironmentRow represents an environment for display
type EnvironmentRow struct {
	UUID            string `json:"environment_uuid"`
	EnvironmentName string `json:"environment_name"`
}

// ProjectListRow represents a project for list display (without environments)
type ProjectListRow struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Project related commands",
}

var listProjectsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
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
		var rows []ProjectListRow
		for _, p := range projects {
			desc := ""
			if p.Description != nil {
				desc = *p.Description
			}
			rows = append(rows, ProjectListRow{
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

var oneProjectCmd = &cobra.Command{
	Use:   "get [uuid]",
	Short: "Get a project by uuid",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
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
		rows := expandProjectEnvironments(project)

		formatter, err := output.NewFormatter(format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return err
		}

		return formatter.Format(rows)
	},
}

// expandProjectEnvironments creates environment rows for display
func expandProjectEnvironments(project *models.Project) []EnvironmentRow {
	var rows []EnvironmentRow

	// If no environments, return empty list
	if len(project.Environments) == 0 {
		return rows
	}

	// Create one row per environment with just UUID and Name
	for _, env := range project.Environments {
		rows = append(rows, EnvironmentRow{
			UUID:            env.UUID,
			EnvironmentName: env.Name,
		})
	}

	return rows
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(listProjectsCmd)
	projectsCmd.AddCommand(oneProjectCmd)
}

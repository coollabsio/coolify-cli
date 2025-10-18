package database

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewUpdateCommand updates a database
func NewUpdateCommand() *cobra.Command {
	updateDatabaseCmd := &cobra.Command{
		Use:   "update <uuid>",
		Short: "Update a database",
		Long:  `Update a database's configuration by UUID.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			req := &models.DatabaseUpdateRequest{}
			hasChanges := false

			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = &name
				hasChanges = true
			}
			if cmd.Flags().Changed("description") {
				desc, _ := cmd.Flags().GetString("description")
				req.Description = &desc
				hasChanges = true
			}
			if cmd.Flags().Changed("image") {
				image, _ := cmd.Flags().GetString("image")
				req.Image = &image
				hasChanges = true
			}
			if cmd.Flags().Changed("is-public") {
				isPublic, _ := cmd.Flags().GetBool("is-public")
				req.IsPublic = &isPublic
				hasChanges = true
			}
			if cmd.Flags().Changed("public-port") {
				port, _ := cmd.Flags().GetInt("public-port")
				req.PublicPort = &port
				hasChanges = true
			}

			// Resource limits
			if cmd.Flags().Changed("limits-memory") {
				mem, _ := cmd.Flags().GetString("limits-memory")
				req.LimitsMemory = &mem
				hasChanges = true
			}
			if cmd.Flags().Changed("limits-cpus") {
				cpus, _ := cmd.Flags().GetString("limits-cpus")
				req.LimitsCpus = &cpus
				hasChanges = true
			}

			if !hasChanges {
				return fmt.Errorf("no fields to update")
			}

			// Validate is-public requires public-port
			if req.IsPublic != nil && *req.IsPublic {
				// If setting to public, check if port is provided or fetch current database to check existing port
				if req.PublicPort == nil || *req.PublicPort == 0 {
					client, err := cli.GetAPIClient(cmd)
					if err != nil {
						return fmt.Errorf("failed to get API client: %w", err)
					}

					dbService := service.NewDatabaseService(client)
					currentDB, err := dbService.Get(ctx, uuid)
					if err != nil {
						return fmt.Errorf("failed to get current database: %w", err)
					}

					// Check if database already has a public port
					if currentDB.PublicPort == nil || *currentDB.PublicPort == 0 {
						return fmt.Errorf("cannot set database as public without a public port. Please provide --public-port")
					}
				}
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			dbService := service.NewDatabaseService(client)
			err = dbService.Update(ctx, uuid, req)
			if err != nil {
				return fmt.Errorf("failed to update database: %w", err)
			}

			fmt.Println("Database updated successfully")
			return nil
		},
	}

	updateDatabaseCmd.Flags().String("name", "", "Database name")
	updateDatabaseCmd.Flags().String("description", "", "Database description")
	updateDatabaseCmd.Flags().String("image", "", "Docker image")
	updateDatabaseCmd.Flags().Bool("is-public", false, "Make database publicly accessible")
	updateDatabaseCmd.Flags().Int("public-port", 0, "Public port")
	updateDatabaseCmd.Flags().String("limits-memory", "", "Memory limit")
	updateDatabaseCmd.Flags().String("limits-cpus", "", "CPU limit")

	return updateDatabaseCmd
}

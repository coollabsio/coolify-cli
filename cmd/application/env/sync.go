package env

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/parser"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewSyncEnvCommand() *cobra.Command {
	syncEnvCmd := &cobra.Command{
		Use:   "sync <app_uuid>",
		Short: "Sync environment variables from a .env file",
		Long: `Sync environment variables from a .env file. This command intelligently:
- Updates existing environment variables with new values
- Creates new environment variables that don't exist yet
- Uses efficient bulk operations where possible

Example: coolify app env sync abc123 --file .env.production`,
		Args: cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			filePath, _ := cmd.Flags().GetString("file")
			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			isBuildTime, _ := cmd.Flags().GetBool("build-time")
			isPreview, _ := cmd.Flags().GetBool("preview")
			isLiteral, _ := cmd.Flags().GetBool("is-literal")

			// Parse the .env file
			envVars, err := parser.ParseEnvFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to parse .env file: %w", err)
			}

			if len(envVars) == 0 {
				fmt.Println("No environment variables found in file.")
				return nil
			}

			fmt.Printf("Found %d environment variables in file. Syncing...\n", len(envVars))

			// Fetch existing environment variables
			appSvc := service.NewApplicationService(client)
			existingEnvs, err := appSvc.ListEnvs(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to list existing environment variables: %w", err)
			}

			// Build a map of existing env vars by key
			existingMap := make(map[string]models.EnvironmentVariable)
			for _, env := range existingEnvs {
				existingMap[env.Key] = env
			}

			// Separate into updates and creates
			var toUpdate []models.EnvironmentVariableCreateRequest
			var toCreate []models.EnvironmentVariableCreateRequest

			for _, envVar := range envVars {
				req := models.EnvironmentVariableCreateRequest{
					Key:   envVar.Key,
					Value: envVar.Value,
				}

				// Apply flags if explicitly provided
				if cmd.Flags().Changed("build-time") {
					req.IsBuildTime = &isBuildTime
				}
				if cmd.Flags().Changed("preview") {
					req.IsPreview = &isPreview
				}
				if cmd.Flags().Changed("is-literal") {
					req.IsLiteral = &isLiteral
				}

				// Auto-detect multiline values
				if strings.Contains(envVar.Value, "\n") {
					multiline := true
					req.IsMultiline = &multiline
				}

				if _, exists := existingMap[envVar.Key]; exists {
					toUpdate = append(toUpdate, req)
				} else {
					toCreate = append(toCreate, req)
				}
			}

			updateCount := 0
			createCount := 0
			failCount := 0

			// Perform bulk update if there are vars to update
			if len(toUpdate) > 0 {
				fmt.Printf("Updating %d existing variables...\n", len(toUpdate))
				bulkReq := &service.BulkUpdateEnvsRequest{
					Data: toUpdate,
				}
				_, err := appSvc.BulkUpdateEnvs(ctx, uuid, bulkReq)
				if err != nil {
					fmt.Printf("  ✗ Bulk update failed: %v\n", err)
					failCount += len(toUpdate)
				} else {
					updateCount = len(toUpdate)
					fmt.Printf("  ✓ Successfully updated %d variables\n", updateCount)
				}
			}

			// Create new variables one by one
			if len(toCreate) > 0 {
				fmt.Printf("Creating %d new variables...\n", len(toCreate))
				for _, req := range toCreate {
					_, err := appSvc.CreateEnv(ctx, uuid, &req)
					if err != nil {
						fmt.Printf("  ✗ Failed to create '%s': %v\n", req.Key, err)
						failCount++
					} else {
						fmt.Printf("  ✓ Created '%s'\n", req.Key)
						createCount++
					}
				}
			}

			fmt.Printf("\nSync complete: %d updated, %d created, %d failed\n", updateCount, createCount, failCount)

			if failCount > 0 {
				return fmt.Errorf("some environment variables failed to sync")
			}

			return nil
		},
	}

	syncEnvCmd.Flags().StringP("file", "f", "", "Path to .env file (required)")
	syncEnvCmd.Flags().Bool("build-time", false, "Make all variables available at build time")
	syncEnvCmd.Flags().Bool("preview", false, "Make all variables available in preview deployments")
	syncEnvCmd.Flags().Bool("is-literal", false, "Treat all values as literal (don't interpolate variables)")
	return syncEnvCmd
}

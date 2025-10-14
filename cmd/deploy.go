package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy related commands",
}

var deployByUuidCmd = &cobra.Command{
	Use:   "uuid <uuid>",
	Short: "Deploy by uuid",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		force, _ := cmd.Flags().GetBool("force")
		deploySvc := service.NewDeploymentService(client)
		result, err := deploySvc.Deploy(ctx, uuid, force)
		if err != nil {
			return fmt.Errorf("failed to deploy resource: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return err
		}

		return formatter.Format(result)
	},
}

var deployByNameCmd = &cobra.Command{
	Use:   "name <name>",
	Short: "Deploy by resource name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		name := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		// Find resource by name
		resourceSvc := service.NewResourceService(client)
		resources, err := resourceSvc.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}

		var matchedUUID string
		for _, r := range resources {
			if r.Name == name {
				matchedUUID = r.UUID
				break
			}
		}

		if matchedUUID == "" {
			return fmt.Errorf("resource with name '%s' not found", name)
		}

		// Deploy using the found UUID
		force, _ := cmd.Flags().GetBool("force")
		deploySvc := service.NewDeploymentService(client)
		result, err := deploySvc.Deploy(ctx, matchedUUID, force)
		if err != nil {
			return fmt.Errorf("failed to deploy resource: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return err
		}

		return formatter.Format(result)
	},
}

var deployBatchCmd = &cobra.Command{
	Use:   "batch <name1,name2,...>",
	Short: "Deploy multiple resources by name",
	Long: `Deploy multiple resources at once.
Provide resource names as comma-separated values.
Example: coolify deploy batch app1,app2,app3`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		namesStr := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		// Parse comma-separated names
		names := make([]string, 0)
		for _, name := range strings.Split(namesStr, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				names = append(names, name)
			}
		}

		if len(names) == 0 {
			return fmt.Errorf("no resource names provided")
		}

		// Find resources by name
		resourceSvc := service.NewResourceService(client)
		resources, err := resourceSvc.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list resources: %w", err)
		}

		// Build map of name -> UUID
		nameToUUID := make(map[string]string)
		for _, r := range resources {
			nameToUUID[r.Name] = r.UUID
		}

		// Validate all names exist
		var notFound []string
		for _, name := range names {
			if _, exists := nameToUUID[name]; !exists {
				notFound = append(notFound, name)
			}
		}
		if len(notFound) > 0 {
			return fmt.Errorf("resources not found: %v", notFound)
		}

		// Deploy all resources
		force, _ := cmd.Flags().GetBool("force")
		deploySvc := service.NewDeploymentService(client)

		type result struct {
			Name    string
			UUID    string
			Success bool
			Message string
			Error   string
		}

		results := make([]result, 0, len(names))

		for _, name := range names {
			uuid := nameToUUID[name]
			fmt.Printf("Deploying %s...\n", name)

			res, err := deploySvc.Deploy(ctx, uuid, force)
			if err != nil {
				results = append(results, result{
					Name:    name,
					UUID:    uuid,
					Success: false,
					Error:   err.Error(),
				})
				fmt.Printf("  ❌ Failed: %v\n", err)
			} else {
				results = append(results, result{
					Name:    name,
					UUID:    uuid,
					Success: true,
					Message: res.Message,
				})
				fmt.Printf("  ✅ Success: %s\n", res.Message)
			}
		}

		// Summary
		successCount := 0
		for _, r := range results {
			if r.Success {
				successCount++
			}
		}

		fmt.Printf("\nBatch deployment complete: %d/%d succeeded\n", successCount, len(results))

		if successCount < len(results) {
			return fmt.Errorf("some deployments failed")
		}

		return nil
	},
}

func init() {
	deployByUuidCmd.Flags().Bool("force", false, "Force deployment")
	deployByNameCmd.Flags().Bool("force", false, "Force deployment")
	deployBatchCmd.Flags().Bool("force", false, "Force deployment")

	rootCmd.AddCommand(deployCmd)
	deployCmd.AddCommand(deployByUuidCmd)
	deployCmd.AddCommand(deployByNameCmd)
	deployCmd.AddCommand(deployBatchCmd)
}

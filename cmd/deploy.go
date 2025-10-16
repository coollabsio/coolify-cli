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

// DeployResultDisplay represents a deploy result for table display
type DeployResultDisplay struct {
	Message        string `json:"message"`
	DeploymentUUID string `json:"deployment_uuid"`
}

var deployByUuidCmd = &cobra.Command{
	Use:   "uuid <uuid>",
	Short: "Deploy by uuid",
	Args:  exactArgs(1, "<uuid>"),
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

		// For table format, convert deployment info array to display format
		if format == output.FormatTable {
			displays := make([]DeployResultDisplay, len(result.Deployments))
			for i, dep := range result.Deployments {
				displays[i] = DeployResultDisplay{
					Message:        dep.Message,
					DeploymentUUID: dep.DeploymentUUID,
				}
			}
			return formatter.Format(displays)
		}

		return formatter.Format(result)
	},
}

var deployByNameCmd = &cobra.Command{
	Use:   "name <name>",
	Short: "Deploy by resource name",
	Args:  exactArgs(1, "<uuid>"),
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

		// For table format, convert deployment info array to display format
		if format == output.FormatTable {
			displays := make([]DeployResultDisplay, len(result.Deployments))
			for i, dep := range result.Deployments {
				displays[i] = DeployResultDisplay{
					Message:        dep.Message,
					DeploymentUUID: dep.DeploymentUUID,
				}
			}
			return formatter.Format(displays)
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
	Args: exactArgs(1, "<uuid>"),
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
				// Get first deployment message from the array
				message := ""
				if len(res.Deployments) > 0 {
					message = res.Deployments[0].Message
				}
				results = append(results, result{
					Name:    name,
					UUID:    uuid,
					Success: true,
					Message: message,
				})
				fmt.Printf("  ✅ Success: %s\n", message)
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

var listDeploymentsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all deployments",
	Long:  `List all currently running deployments across all resources.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		deploySvc := service.NewDeploymentService(client)
		deployments, err := deploySvc.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list deployments: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(deployments)
	},
}

var getDeploymentCmd = &cobra.Command{
	Use:   "get <uuid>",
	Short: "Get deployment details by UUID",
	Long:  `Get detailed information about a specific deployment by its UUID.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		deploySvc := service.NewDeploymentService(client)
		deployment, err := deploySvc.Get(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(deployment)
	},
}

var cancelDeploymentCmd = &cobra.Command{
	Use:   "cancel <uuid>",
	Short: "Cancel a deployment by UUID",
	Long:  `Cancel an in-progress deployment. This will stop the deployment process and clean up any temporary resources.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		force, _ := cmd.Flags().GetBool("force")

		// Prompt for confirmation unless --force is used
		if !force {
			var response string
			fmt.Printf("Are you sure you want to cancel deployment %s? (yes/no): ", uuid)
			fmt.Scanln(&response)

			if response != "yes" && response != "y" {
				fmt.Println("Cancel aborted.")
				return nil
			}
		}

		deploySvc := service.NewDeploymentService(client)
		result, err := deploySvc.Cancel(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to cancel deployment: %w", err)
		}

		format, _ := cmd.Flags().GetString("format")
		formatter, err := output.NewFormatter(format, output.Options{})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(result)
	},
}

func init() {
	deployByUuidCmd.Flags().Bool("force", false, "Force deployment")
	deployByNameCmd.Flags().Bool("force", false, "Force deployment")
	deployBatchCmd.Flags().Bool("force", false, "Force deployment")
	cancelDeploymentCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	rootCmd.AddCommand(deployCmd)
	deployCmd.AddCommand(deployByUuidCmd)
	deployCmd.AddCommand(deployByNameCmd)
	deployCmd.AddCommand(deployBatchCmd)
	deployCmd.AddCommand(listDeploymentsCmd)
	deployCmd.AddCommand(getDeploymentCmd)
	deployCmd.AddCommand(cancelDeploymentCmd)
}

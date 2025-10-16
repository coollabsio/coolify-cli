package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/parser"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

var servicesCmd = &cobra.Command{
	Use:     "service",
	Aliases: []string{"services", "svc"},
	Short:   "Service related commands",
	Long:    `Manage Coolify one-click services (databases, Redis, PostgreSQL, etc.).`,
}

var listServicesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all services",
	Long:  `List all services in your Coolify instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		serviceSvc := service.NewServiceService(client)
		services, err := serviceSvc.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list services: %w", err)
		}

		formatter, err := output.NewFormatter(Format, output.Options{})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(services)
	},
}

var getServiceCmd = &cobra.Command{
	Use:   "get <uuid>",
	Short: "Get service details",
	Long:  `Get detailed information about a specific service.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		serviceSvc := service.NewServiceService(client)
		svc, err := serviceSvc.Get(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to get service: %w", err)
		}

		formatter, err := output.NewFormatter(Format, output.Options{})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(svc)
	},
}

var startServiceCmd = &cobra.Command{
	Use:   "start <uuid>",
	Short: "Start a service",
	Long:  `Start a service (deploy all containers).`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		serviceSvc := service.NewServiceService(client)
		resp, err := serviceSvc.Start(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}

		fmt.Println(resp.Message)
		return nil
	},
}

var stopServiceCmd = &cobra.Command{
	Use:   "stop <uuid>",
	Short: "Stop a service",
	Long:  `Stop a service (stop all containers).`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		serviceSvc := service.NewServiceService(client)
		resp, err := serviceSvc.Stop(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}

		fmt.Println(resp.Message)
		return nil
	},
}

var restartServiceCmd = &cobra.Command{
	Use:   "restart <uuid>",
	Short: "Restart a service",
	Long:  `Restart a service (restart all containers).`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		serviceSvc := service.NewServiceService(client)
		resp, err := serviceSvc.Restart(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}

		fmt.Println(resp.Message)
		return nil
	},
}

var deleteServiceCmd = &cobra.Command{
	Use:   "delete <uuid>",
	Short: "Delete a service",
	Long:  `Delete a service and optionally clean up its configurations, volumes, and networks.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		force, _ := cmd.Flags().GetBool("force")
		deleteConfigurations, _ := cmd.Flags().GetBool("delete-configurations")
		deleteVolumes, _ := cmd.Flags().GetBool("delete-volumes")
		dockerCleanup, _ := cmd.Flags().GetBool("docker-cleanup")
		deleteConnectedNetworks, _ := cmd.Flags().GetBool("delete-connected-networks")

		// Prompt for confirmation unless --force is used
		if !force {
			var response string
			fmt.Printf("Are you sure you want to delete this service? (yes/no): ")
			fmt.Scanln(&response)

			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled.")
				return nil
			}
		}

		serviceSvc := service.NewServiceService(client)
		err = serviceSvc.Delete(ctx, uuid, deleteConfigurations, deleteVolumes, dockerCleanup, deleteConnectedNetworks)
		if err != nil {
			return fmt.Errorf("failed to delete service: %w", err)
		}

		fmt.Println("Service deletion request queued.")
		return nil
	},
}

var envsServiceCmd = &cobra.Command{
	Use:     "env",
	Aliases: []string{"envs", "environment"},
	Short:   "Manage service environment variables",
	Long:    `Manage environment variables for a service. All commands require the service UUID first to establish context.`,
}

var listServiceEnvsCmd = &cobra.Command{
	Use:   "list <service_uuid>",
	Short: "List all environment variables for a service",
	Long:  `List all environment variables for a specific service.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		serviceSvc := service.NewServiceService(client)
		envs, err := serviceSvc.ListEnvs(ctx, uuid)
		if err != nil {
			return fmt.Errorf("failed to list environment variables: %w", err)
		}

		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

		// Mask sensitive values unless --show-sensitive is used
		if !showSensitive {
			for i := range envs {
				envs[i].Value = "********"
				if envs[i].RealValue != nil {
					masked := "********"
					envs[i].RealValue = &masked
				}
			}
		}

		formatter, err := output.NewFormatter(Format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(envs)
	},
}

var getServiceEnvCmd = &cobra.Command{
	Use:   "get <service_uuid> <env_uuid_or_key>",
	Short: "Get environment variable details",
	Long:  `Get detailed information about a specific environment variable. First UUID is the service, second is the environment variable UUID or key name.`,
	Args:  exactArgs(2, "<uuid1> <uuid2>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		serviceUUID := args[0]
		envUUID := args[1]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		serviceSvc := service.NewServiceService(client)
		env, err := serviceSvc.GetEnv(ctx, serviceUUID, envUUID)
		if err != nil {
			return fmt.Errorf("failed to get environment variable: %w", err)
		}

		showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

		// Mask sensitive value unless --show-sensitive is used
		if !showSensitive {
			env.Value = "********"
			if env.RealValue != nil {
				masked := "********"
				env.RealValue = &masked
			}
		}

		formatter, err := output.NewFormatter(Format, output.Options{
			ShowSensitive: showSensitive,
		})
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		return formatter.Format(env)
	},
}

var createServiceEnvCmd = &cobra.Command{
	Use:   "create <service_uuid>",
	Short: "Create an environment variable for a service",
	Long:  `Create a new environment variable for a specific service. Use --key and --value flags to specify the variable.`,
	Args:  exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		isBuildTime, _ := cmd.Flags().GetBool("build-time")
		isPreview, _ := cmd.Flags().GetBool("preview")
		isLiteral, _ := cmd.Flags().GetBool("is-literal")
		isMultiline, _ := cmd.Flags().GetBool("is-multiline")

		if key == "" {
			return fmt.Errorf("--key is required")
		}
		if value == "" {
			return fmt.Errorf("--value is required")
		}

		req := &models.EnvironmentVariableCreateRequest{
			Key:   key,
			Value: value,
		}

		// Only set flags if they were explicitly provided
		if cmd.Flags().Changed("build-time") {
			req.IsBuildTime = &isBuildTime
		}
		if cmd.Flags().Changed("preview") {
			req.IsPreview = &isPreview
		}
		if cmd.Flags().Changed("is-literal") {
			req.IsLiteral = &isLiteral
		}
		if cmd.Flags().Changed("is-multiline") {
			req.IsMultiline = &isMultiline
		}

		serviceSvc := service.NewServiceService(client)
		env, err := serviceSvc.CreateEnv(ctx, uuid, req)
		if err != nil {
			return fmt.Errorf("failed to create environment variable: %w", err)
		}

		fmt.Printf("Environment variable '%s' created successfully.\n", env.Key)
		return nil
	},
}

var updateServiceEnvCmd = &cobra.Command{
	Use:   "update <service_uuid> <env_uuid>",
	Short: "Update an environment variable",
	Long:  `Update an existing environment variable. First UUID is the service, second is the specific environment variable to update.`,
	Args:  exactArgs(2, "<uuid1> <uuid2>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		serviceUUID := args[0]
		envUUID := args[1]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		req := &models.EnvironmentVariableUpdateRequest{
			UUID: envUUID,
		}

		// Only set fields that were provided
		if cmd.Flags().Changed("key") {
			key, _ := cmd.Flags().GetString("key")
			req.Key = &key
		}
		if cmd.Flags().Changed("value") {
			value, _ := cmd.Flags().GetString("value")
			req.Value = &value
		}
		if cmd.Flags().Changed("build-time") {
			isBuildTime, _ := cmd.Flags().GetBool("build-time")
			req.IsBuildTime = &isBuildTime
		}
		if cmd.Flags().Changed("preview") {
			isPreview, _ := cmd.Flags().GetBool("preview")
			req.IsPreview = &isPreview
		}
		if cmd.Flags().Changed("is-literal") {
			isLiteral, _ := cmd.Flags().GetBool("is-literal")
			req.IsLiteral = &isLiteral
		}
		if cmd.Flags().Changed("is-multiline") {
			isMultiline, _ := cmd.Flags().GetBool("is-multiline")
			req.IsMultiline = &isMultiline
		}

		// Check if at least one field is being updated
		if req.Key == nil && req.Value == nil && req.IsBuildTime == nil && req.IsPreview == nil && req.IsLiteral == nil && req.IsMultiline == nil {
			return fmt.Errorf("at least one field must be provided to update (--key, --value, --build-time, --preview, --is-literal, or --is-multiline)")
		}

		serviceSvc := service.NewServiceService(client)
		env, err := serviceSvc.UpdateEnv(ctx, serviceUUID, req)
		if err != nil {
			return fmt.Errorf("failed to update environment variable: %w", err)
		}

		fmt.Printf("Environment variable '%s' updated successfully.\n", env.Key)
		return nil
	},
}

var deleteServiceEnvCmd = &cobra.Command{
	Use:   "delete <service_uuid> <env_uuid>",
	Short: "Delete an environment variable",
	Long:  `Delete an environment variable from a service. First UUID is the service, second is the specific environment variable to delete.`,
	Args:  exactArgs(2, "<uuid1> <uuid2>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		serviceUUID := args[0]
		envUUID := args[1]

		client, err := getAPIClient(cmd)
		if err != nil {
			return fmt.Errorf("failed to get API client: %w", err)
		}

		force, _ := cmd.Flags().GetBool("force")

		// Prompt for confirmation unless --force is used
		if !force {
			var response string
			fmt.Printf("Are you sure you want to delete this environment variable? (yes/no): ")
			fmt.Scanln(&response)

			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled.")
				return nil
			}
		}

		serviceSvc := service.NewServiceService(client)
		err = serviceSvc.DeleteEnv(ctx, serviceUUID, envUUID)
		if err != nil {
			return fmt.Errorf("failed to delete environment variable: %w", err)
		}

		fmt.Println("Environment variable deleted successfully.")
		return nil
	},
}

var syncServiceEnvCmd = &cobra.Command{
	Use:   "sync <service_uuid>",
	Short: "Sync environment variables from a .env file",
	Long: `Sync environment variables from a .env file. This command intelligently:
- Updates existing environment variables with new values
- Creates new environment variables that don't exist yet
- Uses efficient bulk operations where possible

Example: coolify service env sync abc123 --file .env.production`,
	Args: exactArgs(1, "<uuid>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		uuid := args[0]

		client, err := getAPIClient(cmd)
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
		serviceSvc := service.NewServiceService(client)
		existingEnvs, err := serviceSvc.ListEnvs(ctx, uuid)
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
			_, err := serviceSvc.BulkUpdateEnvs(ctx, uuid, bulkReq)
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
				_, err := serviceSvc.CreateEnv(ctx, uuid, &req)
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

func init() {
	// Define delete command flags
	deleteServiceCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	deleteServiceCmd.Flags().Bool("delete-configurations", true, "Delete configurations")
	deleteServiceCmd.Flags().Bool("delete-volumes", true, "Delete volumes")
	deleteServiceCmd.Flags().Bool("docker-cleanup", true, "Run docker cleanup")
	deleteServiceCmd.Flags().Bool("delete-connected-networks", true, "Delete connected networks")

	// Define envs create command flags
	createServiceEnvCmd.Flags().String("key", "", "Environment variable key (required)")
	createServiceEnvCmd.Flags().String("value", "", "Environment variable value (required)")
	createServiceEnvCmd.Flags().Bool("build-time", false, "Available at build time")
	createServiceEnvCmd.Flags().Bool("preview", false, "Available in preview deployments")
	createServiceEnvCmd.Flags().Bool("is-literal", false, "Treat value as literal (don't interpolate variables)")
	createServiceEnvCmd.Flags().Bool("is-multiline", false, "Value is multiline")

	// Define envs update command flags
	updateServiceEnvCmd.Flags().String("key", "", "New environment variable key")
	updateServiceEnvCmd.Flags().String("value", "", "New environment variable value")
	updateServiceEnvCmd.Flags().Bool("build-time", false, "Available at build time")
	updateServiceEnvCmd.Flags().Bool("preview", false, "Available in preview deployments")
	updateServiceEnvCmd.Flags().Bool("is-literal", false, "Treat value as literal (don't interpolate variables)")
	updateServiceEnvCmd.Flags().Bool("is-multiline", false, "Value is multiline")

	// Define envs delete command flags
	deleteServiceEnvCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	// Define envs sync command flags
	syncServiceEnvCmd.Flags().StringP("file", "f", "", "Path to .env file (required)")
	syncServiceEnvCmd.Flags().Bool("build-time", false, "Make all variables available at build time")
	syncServiceEnvCmd.Flags().Bool("preview", false, "Make all variables available in preview deployments")
	syncServiceEnvCmd.Flags().Bool("is-literal", false, "Treat all values as literal (don't interpolate variables)")

	rootCmd.AddCommand(servicesCmd)
	servicesCmd.AddCommand(listServicesCmd)
	servicesCmd.AddCommand(getServiceCmd)
	servicesCmd.AddCommand(startServiceCmd)
	servicesCmd.AddCommand(stopServiceCmd)
	servicesCmd.AddCommand(restartServiceCmd)
	servicesCmd.AddCommand(deleteServiceCmd)
	servicesCmd.AddCommand(envsServiceCmd)
	envsServiceCmd.AddCommand(listServiceEnvsCmd)
	envsServiceCmd.AddCommand(getServiceEnvCmd)
	envsServiceCmd.AddCommand(createServiceEnvCmd)
	envsServiceCmd.AddCommand(updateServiceEnvCmd)
	envsServiceCmd.AddCommand(deleteServiceEnvCmd)
	envsServiceCmd.AddCommand(syncServiceEnvCmd)
}

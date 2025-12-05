package env

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <service_uuid>",
		Short: "Create an environment variable for a service",
		Long:  `Create a new environment variable for a specific service. Use --key and --value flags to specify the variable.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			key, _ := cmd.Flags().GetString("key")
			value, _ := cmd.Flags().GetString("value")
			isBuildTime, _ := cmd.Flags().GetBool("build-time")
			isLiteral, _ := cmd.Flags().GetBool("is-literal")
			isMultiline, _ := cmd.Flags().GetBool("is-multiline")
			isRuntime, _ := cmd.Flags().GetBool("runtime")

			if key == "" {
				return fmt.Errorf("--key is required")
			}
			if value == "" {
				return fmt.Errorf("--value is required")
			}

			req := &models.ServiceEnvironmentVariableCreateRequest{
				Key:   key,
				Value: value,
			}

			// Only set flags if they were explicitly provided
			if cmd.Flags().Changed("build-time") {
				req.IsBuildTime = &isBuildTime
			}
			if cmd.Flags().Changed("is-literal") {
				req.IsLiteral = &isLiteral
			}
			if cmd.Flags().Changed("is-multiline") {
				req.IsMultiline = &isMultiline
			}
			if cmd.Flags().Changed("runtime") {
				req.IsRuntime = &isRuntime
			}

			serviceSvc := service.NewService(client)
			env, err := serviceSvc.CreateEnv(ctx, uuid, req)
			if err != nil {
				return fmt.Errorf("failed to create environment variable: %w", err)
			}

			fmt.Printf("Environment variable '%s' created successfully.\n", env.Key)
			return nil
		},
	}

	cmd.Flags().String("key", "", "Environment variable key (required)")
	cmd.Flags().String("value", "", "Environment variable value (required)")
	cmd.Flags().Bool("build-time", true, "Available at build time (default: true)")
	cmd.Flags().Bool("is-literal", false, "Treat value as literal (don't interpolate variables)")
	cmd.Flags().Bool("is-multiline", false, "Value is multiline")
	cmd.Flags().Bool("runtime", true, "Available at runtime (default: true)")

	return cmd
}

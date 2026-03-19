package env

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <service_uuid>",
		Short: "Update an environment variable",
		Long:  `Update an existing environment variable. UUID is the service.`,
		Args:  cli.ExactArgs(1, "<service_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			serviceUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Check minimum version requirement
			if err := cli.CheckMinimumVersion(ctx, client, "4.0.0-beta.469"); err != nil {
				return err
			}

			req := &models.ServiceEnvironmentVariableUpdateRequest{}

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
			if cmd.Flags().Changed("is-literal") {
				isLiteral, _ := cmd.Flags().GetBool("is-literal")
				req.IsLiteral = &isLiteral
			}
			if cmd.Flags().Changed("is-multiline") {
				isMultiline, _ := cmd.Flags().GetBool("is-multiline")
				req.IsMultiline = &isMultiline
			}
			if cmd.Flags().Changed("runtime") {
				isRuntime, _ := cmd.Flags().GetBool("runtime")
				req.IsRuntime = &isRuntime
			}

			if req.Key == nil {
				return fmt.Errorf("--key is required")
			}
			if req.Value == nil {
				return fmt.Errorf("--value is required")
			}

			serviceSvc := service.NewService(client)
			env, err := serviceSvc.UpdateEnv(ctx, serviceUUID, req)
			if err != nil {
				return fmt.Errorf("failed to update environment variable: %w", err)
			}

			fmt.Printf("Environment variable '%s' updated successfully.\n", env.Key)
			return nil
		},
	}

	cmd.Flags().String("key", "", "New environment variable key")
	cmd.Flags().String("value", "", "New environment variable value")
	cmd.Flags().Bool("build-time", true, "Available at build time (default: true)")
	cmd.Flags().Bool("is-literal", false, "Treat value as literal (don't interpolate variables)")
	cmd.Flags().Bool("is-multiline", false, "Value is multiline")
	cmd.Flags().Bool("runtime", true, "Available at runtime (default: true)")

	return cmd
}

package env

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <service_uuid> <env_uuid>",
		Short: "Update an environment variable",
		Long:  `Update an existing environment variable. First UUID is the service, second is the specific environment variable to update.`,
		Args:  cli.ExactArgs(2, "<uuid1> <uuid2>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			serviceUUID := args[0]
			envUUID := args[1]

			client, err := cli.GetAPIClient(cmd)
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
	cmd.Flags().Bool("build-time", false, "Available at build time")
	cmd.Flags().Bool("preview", false, "Available in preview deployments")
	cmd.Flags().Bool("is-literal", false, "Treat value as literal (don't interpolate variables)")
	cmd.Flags().Bool("is-multiline", false, "Value is multiline")

	return cmd
}

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
		Use:   "create <database_uuid>",
		Short: "Create an environment variable for a database",
		Long:  `Create a new environment variable for a specific database. Use --key and --value flags to specify the variable.`,
		Args:  cli.ExactArgs(1, "<database_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			key, _ := cmd.Flags().GetString("key")
			value, _ := cmd.Flags().GetString("value")

			if key == "" {
				return fmt.Errorf("--key is required")
			}
			if value == "" {
				return fmt.Errorf("--value is required")
			}

			req := &models.DatabaseEnvironmentVariableCreateRequest{
				Key:   key,
				Value: value,
			}

			if cmd.Flags().Changed("is-literal") {
				isLiteral, _ := cmd.Flags().GetBool("is-literal")
				req.IsLiteral = &isLiteral
			}
			if cmd.Flags().Changed("is-multiline") {
				isMultiline, _ := cmd.Flags().GetBool("is-multiline")
				req.IsMultiline = &isMultiline
			}
			if cmd.Flags().Changed("is-shown-once") {
				isShownOnce, _ := cmd.Flags().GetBool("is-shown-once")
				req.IsShownOnce = &isShownOnce
			}
			if cmd.Flags().Changed("comment") {
				comment, _ := cmd.Flags().GetString("comment")
				req.Comment = &comment
			}

			dbSvc := service.NewDatabaseService(client)
			_, err = dbSvc.CreateEnv(ctx, uuid, req)
			if err != nil {
				return fmt.Errorf("failed to create environment variable: %w", err)
			}

			fmt.Printf("Environment variable '%s' created successfully.\n", key)
			return nil
		},
	}

	cmd.Flags().String("key", "", "Environment variable key (required)")
	cmd.Flags().String("value", "", "Environment variable value (required)")
	cmd.Flags().Bool("is-literal", false, "Treat value as literal (don't interpolate variables)")
	cmd.Flags().Bool("is-multiline", false, "Value is multiline")
	cmd.Flags().Bool("is-shown-once", false, "Only show value once")
	cmd.Flags().String("comment", "", "Comment for the environment variable")

	return cmd
}

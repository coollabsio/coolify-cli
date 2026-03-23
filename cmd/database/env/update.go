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
		Use:   "update <database_uuid>",
		Short: "Update an environment variable",
		Long:  `Update an existing environment variable. UUID is the database.`,
		Args:  cli.ExactArgs(1, "<database_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			dbUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Check minimum version requirement
			if err := cli.CheckMinimumVersion(ctx, client, "4.0.0-beta.469"); err != nil {
				return err
			}

			req := &models.DatabaseEnvironmentVariableUpdateRequest{}

			if cmd.Flags().Changed("key") {
				key, _ := cmd.Flags().GetString("key")
				req.Key = &key
			}
			if cmd.Flags().Changed("value") {
				value, _ := cmd.Flags().GetString("value")
				req.Value = &value
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

			if req.Key == nil {
				return fmt.Errorf("--key is required")
			}
			if req.Value == nil {
				return fmt.Errorf("--value is required")
			}

			dbSvc := service.NewDatabaseService(client)
			env, err := dbSvc.UpdateEnv(ctx, dbUUID, req)
			if err != nil {
				return fmt.Errorf("failed to update environment variable: %w", err)
			}

			fmt.Printf("Environment variable '%s' updated successfully.\n", env.Key)
			return nil
		},
	}

	cmd.Flags().String("key", "", "New environment variable key")
	cmd.Flags().String("value", "", "New environment variable value")
	cmd.Flags().Bool("is-literal", false, "Treat value as literal")
	cmd.Flags().Bool("is-multiline", false, "Value is multiline")
	cmd.Flags().Bool("is-shown-once", false, "Only show value once")
	cmd.Flags().String("comment", "", "Comment for the environment variable")

	return cmd
}

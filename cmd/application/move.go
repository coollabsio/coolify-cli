package application

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewMoveCommand creates the application move command.
func NewMoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <uuid>",
		Short: "Move an application to another environment",
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			environmentUUID, _ := cmd.Flags().GetString("environment-uuid")
			if environmentUUID == "" {
				return fmt.Errorf("--environment-uuid is required")
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			response, err := service.NewApplicationService(client).Move(cmd.Context(), args[0], models.ApplicationMoveRequest{
				EnvironmentUUID: environmentUUID,
			})
			if err != nil {
				return fmt.Errorf("failed to move application: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(response)
		},
	}
	cmd.Flags().String("environment-uuid", "", "Target environment UUID (required)")
	return cmd
}

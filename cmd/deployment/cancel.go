package deployment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewCancelCommand cancels a deployment
func NewCancelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <uuid>",
		Short: "Cancel a deployment by UUID",
		Long:  `Cancel an in-progress deployment. This will stop the deployment process and clean up any temporary resources.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Check minimum version requirement
			if err := cli.CheckMinimumVersion(ctx, client, "4.0.0-beta.436"); err != nil {
				return err
			}

			force, _ := cmd.Flags().GetBool("force")

			// Prompt for confirmation unless --force is used
			if !force {
				fmt.Printf("Are you sure you want to cancel deployment %s? (yes/no): ", uuid)
				var response string
				if _, err := fmt.Scanln(&response); err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

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

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	return cmd
}

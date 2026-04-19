package previews

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewDeletePreviewCommand() *cobra.Command {
	deletePreviewCmd := &cobra.Command{
		Use:   "delete <app_uuid> <pr_id>",
		Short: "Delete a preview deployment",
		Long:  `Delete a preview deployment for an application. First argument is the application UUID, second is the pull request ID.`,
		Args:  cli.ExactArgs(2, "<app_uuid> <pr_id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			appUUID := args[0]
			prID := args[1]

			prIDInt, err := strconv.Atoi(prID)
			if err != nil {
				return fmt.Errorf("invalid pr_id: must be an integer")
			}
			if prIDInt <= 0 {
				return fmt.Errorf("invalid pr_id: must be a positive integer")
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := cli.CheckMinimumVersion(ctx, client, "4.0.0-beta.474"); err != nil {
				return err
			}

			force, _ := cmd.Flags().GetBool("force")

			// Prompt for confirmation unless --force is used
			if !force {
				var response string
				fmt.Printf("Are you sure you want to delete the preview deployment for PR %s? (yes/no): ", prID)
				_, err := fmt.Scanln(&response)

				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				if response != "yes" && response != "y" {
					fmt.Println("Delete cancelled.")
					return nil
				}
			}

			appSvc := service.NewApplicationService(client)
			err = appSvc.DeletePreview(ctx, appUUID, prID)
			if err != nil {
				return fmt.Errorf("failed to delete preview deployment: %w", err)
			}

			fmt.Printf("Preview deployment for PR %s deleted successfully.\n", prID)
			return nil
		},
	}

	deletePreviewCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	return deletePreviewCmd
}

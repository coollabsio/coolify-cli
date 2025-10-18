package github

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewDeleteCommand() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <app_uuid>",
		Short: "Delete a GitHub App integration",
		Long:  `Delete a GitHub App integration. The app must not be used by any applications.`,
		Args:  cli.ExactArgs(1, "<app_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			appUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			force, _ := cmd.Flags().GetBool("force")

			// Prompt for confirmation unless --force is used
			if !force {
				var response string
				fmt.Printf("Are you sure you want to delete GitHub App %s? This cannot be undone. (yes/no): ", appUUID)
				_, err := fmt.Scanln(&response)

				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				if response != "yes" && response != "y" {
					fmt.Println("Delete cancelled.")
					return nil
				}
			}

			svc := service.NewGitHubAppService(client)
			err = svc.Delete(ctx, appUUID)
			if err != nil {
				return fmt.Errorf("failed to delete GitHub App: %w", err)
			}

			fmt.Println("GitHub App deleted successfully")
			return nil
		},
	}

	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	return deleteCmd
}

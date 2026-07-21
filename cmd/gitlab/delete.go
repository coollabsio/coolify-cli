package gitlab

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDeleteCommand deletes a GitLab App integration
func NewDeleteCommand() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <app_id_or_uuid>",
		Short: "Delete a GitLab App integration",
		Long:  `Delete a GitLab App integration. The app must not be used by any applications.`,
		Args:  cli.ExactArgs(1, "<app_id_or_uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			idOrUUID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			force, _ := cmd.Flags().GetBool("force")

			if !force {
				var response string
				fmt.Printf("Are you sure you want to delete GitLab App %s? This cannot be undone. (yes/no): ", idOrUUID)
				_, err := fmt.Scanln(&response)
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				if response != "yes" && response != "y" {
					fmt.Println("Delete cancelled.")
					return nil
				}
			}

			svc := service.NewGitLabAppService(client)
			err = svc.Delete(ctx, idOrUUID)
			if err != nil {
				return fmt.Errorf("failed to delete GitLab App: %w", err)
			}

			fmt.Println("GitLab App deleted successfully")
			return nil
		},
	}

	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	return deleteCmd
}

package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDeleteCommand returns the delete project command
func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <uuid>",
		Short: "Delete a project",
		Long: `Delete a project by UUID. The project must be empty (no environments with resources).

Examples:
  coolify project delete <uuid>
  coolify project delete <uuid> --force`,
		Args: cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			force, _ := cmd.Flags().GetBool("force")
			if !force {
				var response string
				fmt.Printf("Are you sure you want to delete project %s? This cannot be undone. (yes/no): ", uuid)
				_, err := fmt.Scanln(&response)
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				if response != "yes" && response != "y" {
					fmt.Println("Delete cancelled.")
					return nil
				}
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			if err := service.NewProjectService(client).Delete(ctx, uuid); err != nil {
				return err
			}

			fmt.Printf("Project %s deleted successfully.\n", uuid)
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	return cmd
}

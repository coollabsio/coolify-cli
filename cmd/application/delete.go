package application

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <uuid>",
		Short: "Delete an application",
		Long:  `Delete an application. This action cannot be undone.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			force, _ := cmd.Flags().GetBool("force")

			if !force {
				var response string
				fmt.Printf("Are you sure you want to delete application %s? This cannot be undone. (yes/no): ", uuid)
				_, err := fmt.Scanln(&response)

				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
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

			appSvc := service.NewApplicationService(client)
			err = appSvc.Delete(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to delete application: %w", err)
			}

			fmt.Printf("Application %s deleted successfully.\n", uuid)
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	return cmd
}

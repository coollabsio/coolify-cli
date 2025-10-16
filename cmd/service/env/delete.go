package env

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <service_uuid> <env_uuid>",
		Short: "Delete an environment variable",
		Long:  `Delete an environment variable from a service. First UUID is the service, second is the specific environment variable to delete.`,
		Args:  cli.ExactArgs(2, "<uuid1> <uuid2>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			serviceUUID := args[0]
			envUUID := args[1]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			force, _ := cmd.Flags().GetBool("force")

			// Prompt for confirmation unless --force is used
			if !force {
				var response string
				fmt.Printf("Are you sure you want to delete this environment variable? (yes/no): ")
				fmt.Scanln(&response)

				if response != "yes" && response != "y" {
					fmt.Println("Delete cancelled.")
					return nil
				}
			}

			serviceSvc := service.NewServiceService(client)
			err = serviceSvc.DeleteEnv(ctx, serviceUUID, envUUID)
			if err != nil {
				return fmt.Errorf("failed to delete environment variable: %w", err)
			}

			fmt.Println("Environment variable deleted successfully.")
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

package service

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDeleteCommand deletes a service
func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <uuid>",
		Short: "Delete a service",
		Long:  `Delete a service and optionally clean up its configurations, volumes, and networks.`,
		Args:  cli.ExactArgs(1, "<uuid>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			force, _ := cmd.Flags().GetBool("force")
			deleteConfigurations, _ := cmd.Flags().GetBool("delete-configurations")
			deleteVolumes, _ := cmd.Flags().GetBool("delete-volumes")
			dockerCleanup, _ := cmd.Flags().GetBool("docker-cleanup")
			deleteConnectedNetworks, _ := cmd.Flags().GetBool("delete-connected-networks")

			// Prompt for confirmation unless --force is used
			if !force {
				var response string
				fmt.Printf("Are you sure you want to delete this service? (yes/no): ")
				_, err := fmt.Scanln(&response)

				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				if response != "yes" && response != "y" {
					fmt.Println("Delete cancelled.")
					return nil
				}
			}

			serviceSvc := service.NewService(client)
			err = serviceSvc.Delete(ctx, uuid, deleteConfigurations, deleteVolumes, dockerCleanup, deleteConnectedNetworks)
			if err != nil {
				return fmt.Errorf("failed to delete service: %w", err)
			}

			fmt.Println("Service deletion request queued.")
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	cmd.Flags().Bool("delete-configurations", true, "Delete configurations")
	cmd.Flags().Bool("delete-volumes", true, "Delete volumes")
	cmd.Flags().Bool("docker-cleanup", true, "Run docker cleanup")
	cmd.Flags().Bool("delete-connected-networks", true, "Delete connected networks")

	return cmd
}

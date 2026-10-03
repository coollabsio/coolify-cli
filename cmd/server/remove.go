package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewRemoveCommand creates the remove command
func NewRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <uuid>",
		Args:  cli.ExactArgs(1, "<uuid>"),
		Short: "Remove a server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Get API client
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Use service layer
			serverSvc := service.NewServerService(client)
			uuid := args[0]

			force, _ := cmd.Flags().GetBool("force")
			deleteFromProvider, _ := cmd.Flags().GetBool("delete-from-provider")

			if err := serverSvc.Delete(ctx, uuid, force, deleteFromProvider); err != nil {
				return fmt.Errorf("failed to delete server: %w", err)
			}

			fmt.Printf("Server %s deleted successfully\n", uuid)
			return nil
		},
	}

	cmd.Flags().Bool("force", false, "Also delete all resources on the server")
	cmd.Flags().Bool("delete-from-provider", false, "Also delete the server from its cloud provider (Hetzner, Vultr or DigitalOcean). This cannot be undone")

	return cmd
}

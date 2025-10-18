package server

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewRemoveCommand creates the remove command
func NewRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <uuid>",
		Args:  cli.ExactArgs(1, "<uuid>"),
		Short: "Remove a server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get API client
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Use service layer
			serverSvc := service.NewServerService(client)
			uuid := args[0]

			if err := serverSvc.Delete(ctx, uuid); err != nil {
				return fmt.Errorf("failed to delete server: %w", err)
			}

			fmt.Printf("Server %s deleted successfully\n", uuid)
			return nil
		},
	}
}

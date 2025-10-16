package server

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewValidateCommand creates the validate command
func NewValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <uuid>",
		Args:  cli.ExactArgs(1, "<uuid>"),
		Short: "Validate a server",
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

			response, err := serverSvc.Validate(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to validate server: %w", err)
			}

			if response.Message != "" {
				fmt.Println(response.Message)
			} else {
				fmt.Printf("Server %s validated successfully\n", uuid)
			}

			return nil
		},
	}
}

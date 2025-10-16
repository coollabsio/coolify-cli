package server

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewAddCommand creates the add command
func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <ip> <private_key_uuid>",
		Args:  cli.ExactArgs(3, "<name> <ip> <private_key_uuid>"),
		Short: "Add a server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Get API client
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Parse arguments and flags
			name := args[0]
			ip := args[1]
			privateKeyUuid := args[2]
			port, _ := cmd.Flags().GetInt("port")
			user, _ := cmd.Flags().GetString("user")
			validate, _ := cmd.Flags().GetBool("validate")

			// Create request
			req := models.ServerCreateRequest{
				Name:            name,
				IP:              ip,
				Port:            port,
				User:            user,
				PrivateKeyUUID:  privateKeyUuid,
				InstantValidate: validate,
			}

			// Use service layer
			serverSvc := service.NewServerService(client)
			response, err := serverSvc.Create(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create server: %w", err)
			}

			if validate {
				fmt.Printf("Server added successfully with uuid %s\n", response.UUID)
			} else {
				fmt.Printf("Server added successfully with uuid %s. Server is not validated. Use 'servers validate %s' to validate the server.\n", response.UUID, response.UUID)
			}

			return nil
		},
	}

	cmd.Flags().IntP("port", "p", 22, "Port")
	cmd.Flags().StringP("user", "u", "root", "User")
	cmd.Flags().Bool("validate", false, "Validate the server")

	return cmd
}

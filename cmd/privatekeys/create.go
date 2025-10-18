package privatekeys

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewCreateCommand creates the create command
func NewCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "add <key_name> <private_key_or_file>",
		Example: `add mykey ~/.ssh/id_rsa`,
		Args:    cli.ExactArgs(2, "<key_name> <private_key_or_file>"),
		Short:   "Add a private key",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			name := args[0]
			privateKeyInput := args[1]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			var privateKey string
			// Check if input is a file path
			if _, err := os.Stat(privateKeyInput); err == nil {
				keyBytes, err := os.ReadFile(privateKeyInput)
				if err != nil {
					return fmt.Errorf("error reading private key file: %w", err)
				}
				privateKey = string(keyBytes)
			} else {
				privateKey = privateKeyInput
			}

			keySvc := service.NewPrivateKeyService(client)
			req := models.PrivateKeyCreateRequest{
				Name:       name,
				PrivateKey: privateKey,
			}

			key, err := keySvc.Create(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to add private key: %w", err)
			}

			fmt.Printf("Private key '%s' added successfully (UUID: %s)\n", key.Name, key.UUID)
			return nil
		},
	}
}

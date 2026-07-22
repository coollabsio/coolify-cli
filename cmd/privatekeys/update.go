package privatekeys

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewUpdateCommand creates the update command
func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <uuid>",
		Args:  cli.ExactArgs(1, "<uuid>"),
		Short: "Update a private key",
		Long: `Update a private key by UUID.

The --private-key flag is required by the API. It accepts either the key content
or a path to a key file (resolved the same way as private-key add).`,
		Example: `  coolify private-key update <uuid> --private-key ~/.ssh/id_rsa
  coolify private-key update <uuid> --private-key ~/.ssh/id_rsa --name mykey
  coolify private-key update <uuid> --private-key ~/.ssh/id_rsa --description "Deploy key"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			privateKeyInput, err := cmd.Flags().GetString("private-key")
			if err != nil {
				return err
			}
			if privateKeyInput == "" {
				return fmt.Errorf("--private-key is required")
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

			req := models.PrivateKeyUpdateRequest{
				PrivateKey: privateKey,
			}

			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = &name
			}
			if cmd.Flags().Changed("description") {
				desc, _ := cmd.Flags().GetString("description")
				req.Description = &desc
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			keySvc := service.NewPrivateKeyService(client)
			key, err := keySvc.Update(ctx, uuid, req)
			if err != nil {
				return fmt.Errorf("failed to update private key: %w", err)
			}

			fmt.Printf("Private key updated successfully (UUID: %s)\n", key.UUID)
			return nil
		},
	}

	cmd.Flags().String("private-key", "", "Private key content or path to key file (required)")
	cmd.Flags().String("name", "", "Private key name")
	cmd.Flags().String("description", "", "Private key description")
	_ = cmd.MarkFlagRequired("private-key")

	return cmd
}

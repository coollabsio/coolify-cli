package privatekeys

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewDeleteCommand creates the delete command
func NewDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <uuid>",
		Args:  cli.ExactArgs(1, "<uuid>"),
		Short: "Remove a private key",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			keySvc := service.NewPrivateKeyService(client)
			err = keySvc.Delete(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to remove private key: %w", err)
			}

			fmt.Println("Private key removed successfully")
			return nil
		},
	}
}

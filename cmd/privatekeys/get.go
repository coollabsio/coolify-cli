package privatekeys

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewGetCommand creates the get command
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <uuid>",
		Args:  cli.ExactArgs(1, "<uuid>"),
		Short: "Get a private key by UUID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			uuid := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			keySvc := service.NewPrivateKeyService(client)
			key, err := keySvc.Get(ctx, uuid)
			if err != nil {
				return fmt.Errorf("failed to get private key: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			if err := formatter.Format(key); err != nil {
				return err
			}

			if !showSensitive && format == output.FormatTable {
				fmt.Println("\nNote: Use -s to show sensitive information.")
			}

			return nil
		},
	}
}

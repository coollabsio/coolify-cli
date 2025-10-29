package privatekeys

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewListCommand creates the list command
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all private keys",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			keySvc := service.NewPrivateKeyService(client)
			keys, err := keySvc.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list private keys: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			if err := formatter.Format(keys); err != nil {
				return err
			}

			if !showSensitive && format == output.FormatTable {
				fmt.Println("\nNote: Use -s to show sensitive information.")
			}

			return nil
		},
	}
}

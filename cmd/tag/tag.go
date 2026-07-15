package tag

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewTagCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "tag", Aliases: []string{"tags"}, Short: "Manage resource tags"}
	cmd.AddCommand(&cobra.Command{
		Use: "list", Short: "List all team tags",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}
			tags, err := service.NewTagService(client).List(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list tags: %w", err)
			}
			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return err
			}
			return formatter.Format(tags)
		},
	})
	return cmd
}

package teams

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewListCommand creates the list command
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all teams",
		Long:  `List all teams you have access to.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			teamSvc := service.NewTeamService(client)
			teams, err := teamSvc.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list teams: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(teams)
		},
	}
}

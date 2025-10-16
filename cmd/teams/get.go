package teams

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the get command
func NewGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get team details by ID",
		Long:  `Get detailed information about a specific team by its ID.`,
		Args:  cli.ExactArgs(1, "<id>"),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			teamID := args[0]

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			teamSvc := service.NewTeamService(client)
			team, err := teamSvc.Get(ctx, teamID)
			if err != nil {
				return fmt.Errorf("failed to get team: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			return formatter.Format(team)
		},
	}
}

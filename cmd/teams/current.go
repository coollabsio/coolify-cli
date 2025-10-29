package teams

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewCurrentCommand creates the current command
func NewCurrentCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Get currently authenticated team",
		Long:  `Get details of the team associated with the current authentication token.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			teamSvc := service.NewTeamService(client)
			team, err := teamSvc.Current(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current team: %w", err)
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

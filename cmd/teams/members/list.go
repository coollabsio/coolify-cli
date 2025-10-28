package members

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
		Use:   "list [team_id]",
		Short: "List team members",
		Long:  `List members of a specific team by ID, or list members of the current team if no ID is provided.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			teamSvc := service.NewTeamService(client)

			// If team ID provided, get members of that team
			// Otherwise get members of current team
			var members interface{}
			var membersErr error

			if len(args) > 0 {
				teamID := args[0]
				members, membersErr = teamSvc.ListMembers(ctx, teamID)
			} else {
				members, membersErr = teamSvc.CurrentMembers(ctx)
			}

			if membersErr != nil {
				return fmt.Errorf("failed to list team members: %w", membersErr)
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")

			formatter, err := output.NewFormatter(format, output.Options{
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}

			if err := formatter.Format(members); err != nil {
				return err
			}

			if !showSensitive && format == output.FormatTable {
				fmt.Println("\nNote: Use -s to show sensitive information.")
			}

			return nil
		},
	}
}

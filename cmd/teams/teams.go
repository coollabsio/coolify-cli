package teams

import (
	"github.com/coollabsio/coolify-cli/cmd/teams/members"
	"github.com/spf13/cobra"
)

// NewTeamsCommand creates the teams parent command
func NewTeamsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "teams",
		Aliases: []string{"team"},
		Short:   "Team related commands",
		Long:    `Manage Coolify teams - list all teams, get team details, view current team, and manage team members.`,
	}

	// Add subcommands
	cmd.AddCommand(NewCurrentCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewListCommand())

	membersCmd := &cobra.Command{
		Use:     "members",
		Aliases: []string{"member"},
		Short:   "Manage team members",
		Long:    `List and manage members of a specific team or the current team.`,
	}
	membersCmd.AddCommand(members.NewListCommand())
	cmd.AddCommand(membersCmd)

	return cmd
}

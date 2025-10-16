package cmd

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
	"github.com/spf13/cobra"
)

var teamsCmd = &cobra.Command{
	Use:     "team",
	Aliases: []string{"teams"},
	Short:   "Team related commands",
	Long:    `Manage Coolify teams - list all teams, get team details, view current team, and manage team members.`,
}

var listTeamsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	Long:  `List all teams you have access to.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
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

var getTeamCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get team details by ID",
	Long:  `Get detailed information about a specific team by its ID.`,
	Args:  exactArgs(1, "<id>"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		teamID := args[0]

		client, err := getAPIClient(cmd)
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

var currentTeamCmd = &cobra.Command{
	Use:   "current",
	Short: "Get currently authenticated team",
	Long:  `Get details of the team associated with the current authentication token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
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

var teamMembersCmd = &cobra.Command{
	Use:   "members",
	Short: "Team members related commands",
	Long:  `Manage team members - list members of a specific team or the current team.`,
}

var listTeamMembersCmd = &cobra.Command{
	Use:   "list [team_id]",
	Short: "List team members",
	Long:  `List members of a specific team by ID, or list members of the current team if no ID is provided.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := getAPIClient(cmd)
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

func init() {
	rootCmd.AddCommand(teamsCmd)
	teamsCmd.AddCommand(listTeamsCmd)
	teamsCmd.AddCommand(getTeamCmd)
	teamsCmd.AddCommand(currentTeamCmd)
	teamsCmd.AddCommand(teamMembersCmd)
	teamMembersCmd.AddCommand(listTeamMembersCmd)
}

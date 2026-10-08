package audit

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

func NewAuditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "audit",
		Aliases: []string{"audit-logs"},
		Short:   "View team audit logs",
	}
	cmd.AddCommand(newListCommand())
	return cmd
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List audit events for the current team",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			source, _ := cmd.Flags().GetString("source")
			action, _ := cmd.Flags().GetString("action")
			search, _ := cmd.Flags().GetString("search")
			perPage, _ := cmd.Flags().GetInt("per-page")
			page, _ := cmd.Flags().GetInt("page")

			result, err := service.NewAuditEventService(client).List(cmd.Context(), service.AuditEventListOptions{
				Source: source, Action: action, Search: search, PerPage: perPage, Page: page,
			})
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			formatter, err := output.NewFormatter(format, output.Options{})
			if err != nil {
				return fmt.Errorf("failed to create formatter: %w", err)
			}
			return formatter.Format(result.Data)
		},
	}

	cmd.Flags().String("source", "", "Filter by source (ui, api, mcp, webhook)")
	cmd.Flags().String("action", "", "Filter by action")
	cmd.Flags().String("search", "", "Search events, actors, and resources")
	cmd.Flags().Int("per-page", 25, "Events per page (maximum 100)")
	cmd.Flags().Int("page", 1, "Page number")
	return cmd
}

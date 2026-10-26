package audit

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewAuditCommand creates the audit parent command
func NewAuditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "audit",
		Aliases: []string{"audit-events", "audit-log"},
		Short:   "Team audit log",
		Long:    `View the audit log of the team bound to the API token. Only team admins and owners can view audit events.`,
	}
	cmd.AddCommand(NewListCommand())
	return cmd
}

// NewListCommand lists audit events (GET /audit-events)
func NewListCommand() *cobra.Command {
	var opts models.AuditEventListOptions
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List audit events, newest first",
		Long: `List audit events of the team bound to the API token, newest first (API: GET /audit-events).

Only team admins and owners can view audit events. Actor e-mail, token, metadata,
changes, IP address and user agent are only returned for tokens with the
read:sensitive or root permission; they are included in json/pretty output.`,
		Example: `  coolify audit list
  coolify audit list --source api --per-page 50
  coolify audit list --action deleted --search my-app --page 2
  coolify audit list --format json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := service.ValidateAuditListOptions(opts); err != nil {
				return err
			}
			// The service treats 0 as "not set"; on the command line it is invalid.
			if opts.PerPage == 0 {
				return fmt.Errorf("--per-page must be between 1 and 100")
			}
			if opts.Page == 0 {
				return fmt.Errorf("--page must be 1 or greater")
			}

			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			page, err := service.NewAuditService(client).List(cmd.Context(), opts)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			showSensitive, _ := cmd.Flags().GetBool("show-sensitive")
			formatter, err := output.NewFormatter(format, output.Options{
				Writer:        cmd.OutOrStdout(),
				ShowSensitive: showSensitive,
			})
			if err != nil {
				return err
			}

			if format != output.FormatTable {
				return formatter.Format(page)
			}

			if err := formatter.Format(page.Data); err != nil {
				return err
			}
			if page.Total > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\nPage %d of %d (%d events).", page.CurrentPage, page.LastPage, page.Total)
				if page.CurrentPage < page.LastPage {
					fmt.Fprintf(cmd.OutOrStdout(), " Use --page %d for more.", page.CurrentPage+1)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&opts.PerPage, "per-page", 25, "Events per page (1-100)")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number")
	cmd.Flags().StringVar(&opts.Search, "search", "", "Search text")
	cmd.Flags().StringVar(&opts.Action, "action", "", "Filter by action, e.g. created, updated, deleted (all = every action)")
	cmd.Flags().StringVar(&opts.Source, "source", "", "Filter by source: "+strings.Join(models.AuditEventSources, ", "))
	return cmd
}

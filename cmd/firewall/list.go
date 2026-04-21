package firewall

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	ifw "github.com/coollabsio/coolify-cli/internal/firewall"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// newListCommand builds `coolify firewall list`.
func newListCommand(flags *FirewallFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed allow rules across all servers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runList(cmd.Context(), cmd, flags)
		},
	}
}

func runList(ctx context.Context, cmd *cobra.Command, flags *FirewallFlags) error {
	if err := flags.SSHMeshFlags.Validate(); err != nil {
		return err
	}
	runner, err := flags.BuildSSHClient()
	if err != nil {
		return fmt.Errorf("SSH client: %w", err)
	}
	return emitList(ctx, cmd, flags, runner)
}

func emitList(
	ctx context.Context,
	cmd *cobra.Command,
	flags *FirewallFlags,
	runner ssh.Runner,
) error {
	tokenFor := tokenResolver(ctx, runner, flags)

	// --all-namespaces → omit the query param so coold returns the union.
	ns := flags.Namespace
	if flags.AllNamespaces {
		ns = ""
	}
	all, results := ifw.CooldListAll(ctx, runner, flags.Servers, flags.SSHUser,
		flags.SSHPort, flags.CooldPort, flags.WGInterface, tokenFor,
		flags.Concurrency, ns)

	rows := make([]models.AllowRuleRow, 0, len(all))
	for _, r := range all {
		rows = append(rows, models.AllowRuleRow{
			Host:      r.Host,
			Namespace: r.Namespace,
			ID:        r.Comment,
			Src:       r.Src.String(),
			Dst:       r.Dst.String(),
			Proto:     r.Proto,
			Port:      r.Port,
			Comment:   r.Comment,
		})
	}

	var errs []string
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", r.Host, r.Err))
		}
	}
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, "Warning:", e)
	}

	format, _ := cmd.Root().PersistentFlags().GetString("format")
	if format == "" {
		format = output.FormatTable
	}
	formatter, err := output.NewFormatter(format, output.Options{Writer: os.Stdout})
	if err != nil {
		return err
	}
	if format == output.FormatJSON || format == output.FormatPretty {
		return formatter.Format(models.FirewallListOutput{
			Rules: rows, Errors: errs,
		})
	}
	if len(rows) == 0 {
		fmt.Fprintln(os.Stderr, "No allow rules found. Run `coolify firewall allow ...` to add one.")
		return nil
	}
	return formatter.Format(rows)
}

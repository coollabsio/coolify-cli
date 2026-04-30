package initcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/wireguard"
)

// NewPlanCommand creates the `coolify init plan` subcommand.
func NewPlanCommand(flags *InitFlags) *cobra.Command {
	var intentFlag string
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Show WireGuard mesh changes without applying them",
		Long: `Reconstruct the current WireGuard state from each server via SSH and
show the actions that apply would execute.  Nothing is changed.

Pass --intent to preview a specific subcommand's behavior (bootstrap, extend,
upgrade). bootstrap is the default and matches the pre-split behavior.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprint(os.Stderr, alphaBanner)
			flags.Intent = intentFlag
			return runPlan(cmd.Context(), cmd, flags)
		},
	}
	cmd.Flags().StringVar(&intentFlag, "intent", "bootstrap",
		`Preview filter: "bootstrap" (all actions), "extend" (treat --new-hosts as fresh, existing hosts peer-refresh only), "upgrade" (version bumps only).`)
	return cmd
}

func runPlan(ctx context.Context, cmd *cobra.Command, flags *InitFlags) error {
	if err := validatePlanFlags(flags); err != nil {
		return err
	}

	desired, err := buildDesired(flags)
	if err != nil {
		return err
	}
	if err := wireguard.ValidateIntent(desired); err != nil {
		return err
	}

	sshClient, err := flags.BuildSSHClient()
	if err != nil {
		return fmt.Errorf("SSH client: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Probing %d server(s)...\n", len(flags.Servers))

	current, err := wireguard.Reconstruct(ctx, sshClient, flags.Servers,
		flags.SSHUser, flags.SSHPort, flags.WGInterface,
		flags.Namespaces, flags.Concurrency)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	plan, err := wireguard.BuildPlan(desired, current)
	if err != nil {
		return fmt.Errorf("build plan: %w", err)
	}

	for _, w := range plan.Warnings {
		fmt.Fprintf(os.Stderr, "Warning [%s]: %s\n", w.Host, w.Reason)
	}

	format, _ := cmd.Root().PersistentFlags().GetString("format")
	intent := intentLabel(flags.Intent)

	if plan.IsEmpty() && len(plan.Skipped) == 0 {
		msg := "No changes needed. Mesh is already converged."
		if format == output.FormatJSON {
			out := models.PlanOutput{
				Servers:  flags.Servers,
				Intent:   intent,
				Actions:  []models.PlanActionRow{},
				Warnings: warningsToStrings(plan.Warnings),
			}
			formatter, _ := output.NewFormatter(format, output.Options{Writer: os.Stdout})
			return formatter.Format(out)
		}
		fmt.Println(msg)
		return nil
	}

	rows := make([]models.PlanActionRow, len(plan.Actions))
	for i, a := range plan.Actions {
		rows[i] = models.PlanActionRow{
			Server: a.Host,
			Action: string(a.Type),
			Detail: a.Detail,
		}
	}
	skipped := skippedRows(plan.Skipped)

	formatter, err := output.NewFormatter(format, output.Options{Writer: os.Stdout})
	if err != nil {
		return err
	}

	if format == output.FormatJSON || format == output.FormatPretty {
		return formatter.Format(models.PlanOutput{
			Servers:  flags.Servers,
			Intent:   intent,
			Actions:  rows,
			Skipped:  skipped,
			Warnings: warningsToStrings(plan.Warnings),
		})
	}

	if len(rows) > 0 {
		if err := formatter.Format(rows); err != nil {
			return err
		}
	} else {
		fmt.Println("No actions scheduled.")
	}
	if len(skipped) > 0 {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Skipped by intent filter:")
		for _, s := range skipped {
			fmt.Fprintf(os.Stderr, "  [%s] %s — %s\n", s.Server, s.Action, s.Reason)
		}
	}
	return nil
}

func validatePlanFlags(f *InitFlags) error {
	if err := f.Validate(); err != nil {
		return err
	}
	return f.ValidateNamespaces()
}

// warningsToStrings formats allocator warnings as human-readable strings.
func warningsToStrings(ws []wireguard.Warning) []string {
	if len(ws) == 0 {
		return nil
	}
	out := make([]string, len(ws))
	for i, w := range ws {
		out[i] = fmt.Sprintf("[%s] %s", w.Host, w.Reason)
	}
	return out
}

// skippedRows converts the plan's intent-filtered actions into render rows.
func skippedRows(ss []wireguard.SkippedAction) []models.PlanSkippedRow {
	if len(ss) == 0 {
		return nil
	}
	out := make([]models.PlanSkippedRow, len(ss))
	for i, s := range ss {
		out[i] = models.PlanSkippedRow{
			Server: s.Action.Host,
			Action: string(s.Action.Type),
			Reason: s.Reason,
		}
	}
	return out
}

// intentLabel normalizes an empty or zero intent to "bootstrap" for display.
func intentLabel(raw string) string {
	if raw == "" {
		return "bootstrap"
	}
	return raw
}

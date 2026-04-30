package initcmd

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	internalssh "github.com/coollabsio/coolify-cli/internal/ssh"
	"github.com/coollabsio/coolify-cli/internal/wireguard"
)

// Ensure internalssh is used (for *internalssh.Client in signatures).
var _ *internalssh.Client

// applyOptions tweaks runApply per subcommand.
type applyOptions struct {
	// SkipAlphaGate, when true, bypasses the interactive "press enter"
	// confirmation. upgrade/extend set it because those are called from the
	// Coolify backend in production, not a human at a terminal.
	SkipAlphaGate bool

	// Header is a one-line banner describing the intent (e.g. "extending
	// mesh with 1 new host"). Printed to stderr before the plan.
	Header string
}

func runApply(ctx context.Context, cmd *cobra.Command, flags *InitFlags, opts applyOptions) error {
	fmt.Fprint(os.Stderr, alphaBanner)

	if err := validatePlanFlags(flags); err != nil {
		return err
	}

	if !opts.SkipAlphaGate && !shouldSkipGate(flags) {
		fmt.Fprintln(os.Stderr, "This command will modify network configuration on the listed servers.")
		fmt.Fprint(os.Stderr, "Press Enter to continue, or Ctrl+C to abort... ")
		reader := bufio.NewReader(os.Stdin)
		if _, err := reader.ReadString('\n'); err != nil {
			return fmt.Errorf("read confirmation: %w", err)
		}
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

	if opts.Header != "" {
		fmt.Fprintln(os.Stderr, opts.Header)
	}
	fmt.Fprintf(os.Stderr, "Probing %d server(s)...\n", len(flags.Servers))

	current, probeErr := wireguard.Reconstruct(ctx, sshClient, flags.Servers,
		flags.SSHUser, flags.SSHPort, flags.WGInterface,
		flags.Namespaces, flags.Concurrency)
	if probeErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", probeErr)
	}

	plan, err := wireguard.BuildPlan(desired, current)
	if err != nil {
		return fmt.Errorf("build plan: %w", err)
	}

	for _, w := range plan.Warnings {
		fmt.Fprintf(os.Stderr, "Warning [%s]: %s\n", w.Host, w.Reason)
	}

	format, _ := cmd.Root().PersistentFlags().GetString("format")

	if plan.IsEmpty() {
		fmt.Fprintln(os.Stderr, "No changes needed. Mesh is already converged.")
	} else {
		fmt.Fprintln(os.Stderr, "Plan:")
		for _, a := range plan.Actions {
			fmt.Fprintf(os.Stderr, "  [%s] %s  %s\n", a.Host, a.Type, a.Detail)
		}
		fmt.Fprintln(os.Stderr)
	}
	if len(plan.Skipped) > 0 {
		fmt.Fprintln(os.Stderr, "Skipped by intent filter:")
		for _, s := range plan.Skipped {
			fmt.Fprintf(os.Stderr, "  [%s] %s — %s\n", s.Action.Host, s.Action.Type, s.Reason)
		}
		fmt.Fprintln(os.Stderr)
	}

	if plan.IsEmpty() {
		return runVerify(ctx, sshClient, flags, desired, format)
	}

	fmt.Fprintln(os.Stderr, "Applying...")
	actionResults, applyErr := wireguard.ApplyMesh(ctx, sshClient,
		flags.SSHUser, flags.SSHPort, desired, current, flags.Concurrency)

	rows := make([]models.ApplyResultRow, len(actionResults))
	for i, r := range actionResults {
		status := "ok"
		detail := r.Action.Detail
		if r.Err != nil {
			status = "error"
			if detail == "" {
				detail = r.Err.Error()
			}
		}
		rows[i] = models.ApplyResultRow{
			Server: r.Action.Host,
			Action: string(r.Action.Type),
			Status: status,
			Detail: detail,
		}
	}

	if format == output.FormatJSON || format == output.FormatPretty {
		verifyRows := collectVerifyRows(ctx, sshClient, flags, desired)
		out := models.ApplyOutput{Results: rows, Verified: verifyRows}
		formatter, ferr := output.NewFormatter(format, output.Options{Writer: os.Stdout})
		if ferr != nil {
			return ferr
		}
		if err := formatter.Format(out); err != nil {
			return err
		}
		return applyErr
	}

	if len(rows) > 0 {
		formatter, _ := output.NewFormatter(output.FormatTable, output.Options{Writer: os.Stdout})
		_ = formatter.Format(rows)
	}

	if err := runVerify(ctx, sshClient, flags, desired, format); err != nil {
		return err
	}

	return applyErr
}

// shouldSkipGate returns true when the interactive alpha gate should be bypassed.
func shouldSkipGate(flags *InitFlags) bool {
	if flags.Yes {
		return true
	}
	if os.Getenv("COOLIFY_NON_INTERACTIVE") == "1" {
		return true
	}
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return true
	}
	return false
}

func runVerify(ctx context.Context, sshClient *internalssh.Client, flags *InitFlags, desired *wireguard.DesiredMesh, format string) error {
	fmt.Fprintln(os.Stderr, "Verifying...")
	vrows := collectVerifyRows(ctx, sshClient, flags, desired)

	formatter, err := output.NewFormatter(format, output.Options{Writer: os.Stdout})
	if err != nil {
		return err
	}
	return formatter.Format(vrows)
}

func collectVerifyRows(ctx context.Context, sshClient *internalssh.Client, flags *InitFlags, desired *wireguard.DesiredMesh) []models.VerifyResultRow {
	vresults := wireguard.Verify(ctx, sshClient,
		flags.Servers, flags.SSHUser, flags.SSHPort, desired.Interface, flags.Concurrency)

	rows := make([]models.VerifyResultRow, len(vresults))
	for i, v := range vresults {
		status := "ok"
		wgIP := ""
		if v.WireGuardIP != nil {
			wgIP = v.WireGuardIP.String()
		}
		if v.Err != nil || !v.Active {
			status = "error"
		}
		rows[i] = models.VerifyResultRow{
			Server:      v.Host,
			WireGuardIP: wgIP,
			PeerCount:   v.PeerCount,
			Status:      status,
		}
	}
	return rows
}

package initcmd

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/services"
	internalssh "github.com/coollabsio/coolify-cli/internal/ssh"
	"github.com/coollabsio/coolify-cli/internal/wireguard"
)

// Ensure internalssh is used (for *internalssh.Client in signatures).
var _ *internalssh.Client

// NewApplyCommand creates the `coolify init apply` subcommand.
func NewApplyCommand(flags *InitFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Bootstrap the WireGuard mesh (executes the plan)",
		Long: `Reconstruct the current state, compute the plan, then execute it.
Idempotent: re-running with no changes produces an empty plan.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runApply(cmd.Context(), cmd, flags)
		},
	}
}

func runApply(ctx context.Context, cmd *cobra.Command, flags *InitFlags) error {
	// Always print the alpha banner to stderr.
	fmt.Fprint(os.Stderr, alphaBanner)

	if err := validatePlanFlags(flags); err != nil {
		return err
	}

	var corrosionSha, cooldSha string
	for _, bp := range []struct {
		label, path string
		out         *string
	}{
		{"corrosion", flags.CorrosionBinaryPath, &corrosionSha},
		{"coold", flags.CooldBinaryPath, &cooldSha},
	} {
		if bp.path == "" {
			return fmt.Errorf("--%s-binary is required", bp.label)
		}
		if _, err := os.Stat(bp.path); err != nil {
			return fmt.Errorf("%s binary %q: %w", bp.label, bp.path, err)
		}
		if err := services.VerifyLinuxARM64(bp.path); err != nil {
			return fmt.Errorf("%s binary: %w", bp.label, err)
		}
		sum, err := wireguard.FileSha256(bp.path)
		if err != nil {
			return fmt.Errorf("hash %s binary: %w", bp.label, err)
		}
		*bp.out = sum
	}

	// Alpha gate: block unless bypassed.
	if !shouldSkipGate(flags) {
		fmt.Fprintln(os.Stderr, "This command will modify network configuration on the listed servers.")
		fmt.Fprint(os.Stderr, "Press Enter to continue, or Ctrl+C to abort... ")
		reader := bufio.NewReader(os.Stdin)
		if _, err := reader.ReadString('\n'); err != nil {
			return fmt.Errorf("read confirmation: %w", err)
		}
	}

	_, mgmtPool, err := net.ParseCIDR(flags.WGMgmtPool)
	if err != nil {
		return fmt.Errorf("invalid --wg-mgmt-pool %q: %w", flags.WGMgmtPool, err)
	}
	_, contPool, err := net.ParseCIDR(flags.ContainerPool)
	if err != nil {
		return fmt.Errorf("invalid --container-pool %q: %w", flags.ContainerPool, err)
	}

	desired := &wireguard.DesiredMesh{
		Hosts:                 flags.Servers,
		Interface:             flags.WGInterface,
		MgmtPool:              mgmtPool,
		ContainerPool:         contPool,
		ContainerPrefix:       flags.ContainerPrefix,
		ListenPort:            flags.WGListenPort,
		InstallPodman:         true,
		PodmanNetworkName:     flags.PodmanNetworkName,
		DefaultDenyContainers: !flags.SkipDefaultDeny,
		InstallCoold:          true,
		CooldBinaryPath:       flags.CooldBinaryPath,
		CorrosionBinaryPath:   flags.CorrosionBinaryPath,
		CorrosionGossipPort:   flags.CorrosionGossipPort,
		CorrosionAPIPort:      flags.CorrosionAPIPort,
		CorrosionBinarySha256: corrosionSha,
		CooldBinarySha256:     cooldSha,
	}

	sshClient, err := flags.BuildSSHClient()
	if err != nil {
		return fmt.Errorf("SSH client: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Probing %d server(s)...\n", len(flags.Servers))

	current, probeErr := wireguard.Reconstruct(ctx, sshClient, flags.Servers,
		flags.SSHUser, flags.SSHPort, flags.WGInterface,
		flags.PodmanNetworkName, flags.Concurrency)
	if probeErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", probeErr)
	}

	plan, err := wireguard.BuildPlan(desired, current)
	if err != nil {
		return fmt.Errorf("build plan: %w", err)
	}

	// Surface allocator warnings.
	for _, w := range plan.Warnings {
		fmt.Fprintf(os.Stderr, "Warning [%s]: %s\n", w.Host, w.Reason)
	}

	format, _ := cmd.Root().PersistentFlags().GetString("format")

	// Print the plan before executing.
	if plan.IsEmpty() {
		fmt.Fprintln(os.Stderr, "No changes needed. Mesh is already converged.")
	} else {
		fmt.Fprintln(os.Stderr, "Plan:")
		for _, a := range plan.Actions {
			fmt.Fprintf(os.Stderr, "  [%s] %s  %s\n", a.Host, a.Type, a.Detail)
		}
		fmt.Fprintln(os.Stderr)
	}

	if plan.IsEmpty() {
		// Nothing to do; still emit verify output.
		return runVerify(ctx, sshClient, flags, desired, format)
	}

	// Execute the plan.
	fmt.Fprintln(os.Stderr, "Applying...")
	actionResults, applyErr := wireguard.ApplyMesh(ctx, sshClient, sshClient,
		flags.SSHUser, flags.SSHPort, desired, current, flags.Concurrency)

	// Render action results.
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
		// Collect verify results first, then emit combined JSON.
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

	// Table output: print results, then verify.
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
	// Skip when stdin is not a terminal (CI/piped).
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

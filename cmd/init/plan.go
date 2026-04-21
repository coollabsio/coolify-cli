package initcmd

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/wireguard"
)

// NewPlanCommand creates the `coolify init plan` subcommand.
func NewPlanCommand(flags *InitFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Show WireGuard mesh changes without applying them",
		Long: `Reconstruct the current WireGuard state from each server via SSH and
show the actions that apply would execute.  Nothing is changed.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Print alpha banner to stderr (non-blocking).
			fmt.Fprint(os.Stderr, alphaBanner)
			return runPlan(cmd.Context(), cmd, flags)
		},
	}
}

func runPlan(ctx context.Context, cmd *cobra.Command, flags *InitFlags) error {
	if err := validatePlanFlags(flags); err != nil {
		return err
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
		Namespaces:            flags.Namespaces,
		DefaultDenyContainers: !flags.SkipDefaultDeny,
		InstallCoold:        true,
		CooldVersion:        flags.CooldVersion,
		CorrosionVersion:    flags.CorrosionVersion,
		CorrosionGossipPort: flags.CorrosionGossipPort,
		CorrosionAPIPort:    flags.CorrosionAPIPort,
	}

	// Build SSH runner (handles passphrase resolution).
	sshClient, err := flags.BuildSSHClient()
	if err != nil {
		return fmt.Errorf("SSH client: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Probing %d server(s)...\n", len(flags.Servers))

	current, err := wireguard.Reconstruct(ctx, sshClient, flags.Servers,
		flags.SSHUser, flags.SSHPort, flags.WGInterface,
		flags.Namespaces, flags.Concurrency)
	if err != nil {
		// Non-fatal: partial state is still usable for plan display.
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
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

	if plan.IsEmpty() {
		msg := "No changes needed. Mesh is already converged."
		if format == output.FormatJSON {
			out := models.PlanOutput{
				Servers:  flags.Servers,
				Actions:  []models.PlanActionRow{},
				Warnings: warningsToStrings(plan.Warnings),
			}
			formatter, _ := output.NewFormatter(format, output.Options{Writer: os.Stdout})
			return formatter.Format(out)
		}
		fmt.Println(msg)
		return nil
	}

	// Render.
	rows := make([]models.PlanActionRow, len(plan.Actions))
	for i, a := range plan.Actions {
		rows[i] = models.PlanActionRow{
			Server: a.Host,
			Action: string(a.Type),
			Detail: a.Detail,
		}
	}

	formatter, err := output.NewFormatter(format, output.Options{Writer: os.Stdout})
	if err != nil {
		return err
	}

	if format == output.FormatJSON || format == output.FormatPretty {
		return formatter.Format(models.PlanOutput{
			Servers:  flags.Servers,
			Actions:  rows,
			Warnings: warningsToStrings(plan.Warnings),
		})
	}

	return formatter.Format(rows)
}

func validatePlanFlags(f *InitFlags) error {
	if err := f.SSHMeshFlags.Validate(); err != nil {
		return err
	}
	return f.MeshNetFlags.ValidateNamespaces()
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

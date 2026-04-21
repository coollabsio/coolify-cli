package firewall

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/common"
	ifw "github.com/coollabsio/coolify-cli/internal/firewall"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/output"
	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// allowRevokeFlags are the per-subcommand flags for `allow` / `revoke`.
type allowRevokeFlags struct {
	From          string
	To            string
	Port          int
	Proto         string
	Bidirectional bool
}

// newAllowCommand builds `coolify firewall allow`.
func newAllowCommand(parent *FirewallFlags) *cobra.Command {
	local := &allowRevokeFlags{}
	cmd := &cobra.Command{
		Use:   "allow",
		Short: "Add an allow rule (from container → to container:port)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAllowRevoke(cmd.Context(), cmd, parent, local, false)
		},
	}
	bindAllowRevokeFlags(cmd, local)
	return cmd
}

// newRevokeCommand builds `coolify firewall revoke`.
func newRevokeCommand(parent *FirewallFlags) *cobra.Command {
	local := &allowRevokeFlags{}
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Remove an allow rule",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAllowRevoke(cmd.Context(), cmd, parent, local, true)
		},
	}
	bindAllowRevokeFlags(cmd, local)
	return cmd
}

func bindAllowRevokeFlags(cmd *cobra.Command, f *allowRevokeFlags) {
	pf := cmd.Flags()
	pf.StringVar(&f.From, "from", "",
		"Source container (name, short-id, raw IP, or host:name) — required")
	pf.StringVar(&f.To, "to", "",
		"Destination container (name, short-id, raw IP, or host:name) — required")
	pf.IntVar(&f.Port, "port", 0,
		"Destination port (required unless --proto is empty)")
	pf.StringVar(&f.Proto, "proto", "tcp",
		"Protocol (tcp, udp, or empty for any)")
	pf.BoolVar(&f.Bidirectional, "bidirectional", false,
		"Also install the reverse rule on the source host (default: one-way; conntrack handles replies)")
}

func validateAllowRevokeFlags(f *allowRevokeFlags) error {
	if f.From == "" {
		return fmt.Errorf("--from is required")
	}
	if f.To == "" {
		return fmt.Errorf("--to is required")
	}
	if f.Proto != "" && f.Proto != "tcp" && f.Proto != "udp" {
		return fmt.Errorf("--proto must be tcp, udp, or empty (got %q)", f.Proto)
	}
	if f.Proto != "" && f.Port <= 0 {
		return fmt.Errorf("--port is required when --proto is set")
	}
	return nil
}

func runAllowRevoke(
	ctx context.Context,
	cmd *cobra.Command,
	parent *FirewallFlags,
	local *allowRevokeFlags,
	revoke bool,
) error {
	if err := parent.SSHMeshFlags.Validate(); err != nil {
		return err
	}
	if err := common.ValidateNamespace(parent.Namespace); err != nil {
		return err
	}
	if err := validateAllowRevokeFlags(local); err != nil {
		return err
	}
	runner, err := parent.BuildSSHClient()
	if err != nil {
		return fmt.Errorf("SSH client: %w", err)
	}
	return emitAllowRevoke(ctx, cmd, parent, local, runner, revoke)
}

// emitAllowRevoke is the core path: discover → resolve → build rule → apply.
// Split from the cobra wrapper so tests inject a fake ssh.Runner.
func emitAllowRevoke(
	ctx context.Context,
	cmd *cobra.Command,
	parent *FirewallFlags,
	local *allowRevokeFlags,
	runner ssh.Runner,
	revoke bool,
) error {
	all, results := discoverAllViaPkg(ctx, runner, parent)
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "Warning: discover %s: %v\n", r.Host, r.Err)
		}
	}

	from, err := resolveEndpoint(local.From, all)
	if err != nil {
		return fmt.Errorf("--from: %w", err)
	}
	to, err := resolveEndpoint(local.To, all)
	if err != nil {
		return fmt.Errorf("--to: %w", err)
	}
	if from.IP == nil || to.IP == nil {
		return fmt.Errorf("failed to resolve endpoint IPs (from=%s to=%s)", local.From, local.To)
	}

	// Determine destination host (rule owner). If `to` was resolved from a
	// raw IP with no container match, try to map it via discovery first.
	dstHost := to.Host
	if dstHost == "" {
		if h, ok := findHostForIP(to.IP, all); ok {
			dstHost = h
		}
	}
	if dstHost == "" {
		return fmt.Errorf("cannot determine destination host for IP %s — no container on the mesh owns it", to.IP)
	}

	srcHost := from.Host
	if srcHost == "" {
		if h, ok := findHostForIP(from.IP, all); ok {
			srcHost = h
		}
	}

	ns := parent.Namespace
	primary := ifw.AllowRule{
		Host:      dstHost,
		Namespace: ns,
		Src:       from.IP,
		Dst:       to.IP,
		Proto:     local.Proto,
		Port:      local.Port,
		Comment:   "cid:" + ifw.ComputeID(ns, from.IP, to.IP, local.Proto, local.Port),
	}
	rules := []ifw.AllowRule{primary}

	if local.Bidirectional {
		if srcHost == "" {
			return fmt.Errorf("--bidirectional requires the source endpoint to belong to a mesh host")
		}
		reverse := ifw.AllowRule{
			Host:      srcHost,
			Namespace: ns,
			Src:       to.IP,
			Dst:       from.IP,
			Proto:     local.Proto,
			Port:      local.Port,
			Comment:   "cid:" + ifw.ComputeID(ns, to.IP, from.IP, local.Proto, local.Port),
		}
		rules = append(rules, reverse)
	}

	action := "allow"
	past := "allowed"
	if revoke {
		action = "revoke"
		past = "revoked"
	}
	tokenFor := tokenResolver(ctx, runner, parent)
	for _, r := range rules {
		token, terr := tokenFor(r.Host)
		if terr != nil {
			return fmt.Errorf("%s on %s: %w", action, r.Host, terr)
		}
		var rerr error
		if revoke {
			// Revoke by id — coold is idempotent (204 even on unknown id).
			id := strings.TrimPrefix(r.Comment, "cid:")
			rerr = ifw.CooldRevoke(ctx, runner, r.Host, parent.SSHUser,
				parent.SSHPort, parent.CooldPort, parent.WGInterface, token, id)
		} else {
			rerr = ifw.CooldApply(ctx, runner, r.Host, parent.SSHUser,
				parent.SSHPort, parent.CooldPort, parent.WGInterface, token, r)
		}
		if rerr != nil {
			return fmt.Errorf("%s on %s: %w", action, r.Host, rerr)
		}
		fmt.Fprintf(os.Stderr, "%s on %s: %s → %s %s/%d\n",
			past, r.Host, ipOrAny(r.Src), ipOrAny(r.Dst),
			protoOrAny(r.Proto), r.Port)
	}

	rows := make([]models.AllowRuleRow, 0, len(rules))
	for _, r := range rules {
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

	format, _ := cmd.Root().PersistentFlags().GetString("format")
	if format == "" {
		format = output.FormatTable
	}
	formatter, err := output.NewFormatter(format, output.Options{Writer: os.Stdout})
	if err != nil {
		return err
	}
	if format == output.FormatJSON || format == output.FormatPretty {
		return formatter.Format(models.FirewallAllowOutput{Rules: rows})
	}
	return formatter.Format(rows)
}

func ipOrAny(ip net.IP) string {
	if ip == nil {
		return "any"
	}
	return ip.String()
}

func protoOrAny(p string) string {
	if p == "" {
		return "any"
	}
	return p
}

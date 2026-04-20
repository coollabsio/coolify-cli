package firewall

import (
	"context"
	"fmt"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// ApplyAllow idempotently appends r to COOLIFY-ALLOW on r.Host, installs the
// reboot-persistence unit on first call, and snapshots the chain to
// /etc/coolify/allow.rules so the rule survives a reboot.
//
// The operation is a single SSH session chained with `&&`:
//
//	ensure chain exists → install persistence unit → -C || -A → save
//
// The chain-exists guard (`iptables -N COOLIFY-ALLOW || true`) covers the
// case where the user ran `coolify init` WITHOUT --default-deny: the chain
// simply won't exist. Without the guard every allow would fail.
func ApplyAllow(
	ctx context.Context,
	runner ssh.Runner,
	r AllowRule,
	user string,
	port int,
) error {
	if r.Host == "" {
		return fmt.Errorf("AllowRule.Host must be set")
	}
	cmd := buildApplyCmd(r)
	stdout, stderr, err := runner.Run(ctx, r.Host, user, port, cmd)
	if err != nil {
		return fmt.Errorf("apply allow on %s: %w (stderr: %s, stdout: %s)",
			r.Host, err, strings.TrimSpace(stderr), strings.TrimSpace(stdout))
	}
	return nil
}

// RevokeAllow idempotently deletes r from COOLIFY-ALLOW on r.Host, then
// re-snapshots the chain. A missing rule is a no-op — the `-C` guard
// short-circuits before `-D` is attempted.
func RevokeAllow(
	ctx context.Context,
	runner ssh.Runner,
	r AllowRule,
	user string,
	port int,
) error {
	if r.Host == "" {
		return fmt.Errorf("AllowRule.Host must be set")
	}
	cmd := buildRevokeCmd(r)
	stdout, stderr, err := runner.Run(ctx, r.Host, user, port, cmd)
	if err != nil {
		return fmt.Errorf("revoke allow on %s: %w (stderr: %s, stdout: %s)",
			r.Host, err, strings.TrimSpace(stderr), strings.TrimSpace(stdout))
	}
	return nil
}

// buildApplyCmd composes the idempotent apply script. Kept separate from
// ApplyAllow so tests can assert on the exact command without SSH mocking.
func buildApplyCmd(r AllowRule) string {
	var b strings.Builder
	// Guarantee chain exists even if --default-deny wasn't set at init.
	b.WriteString(`iptables -N `)
	b.WriteString(ChainName)
	b.WriteString(` 2>/dev/null || true`)
	b.WriteString(" && ")
	// Install persistence unit. Idempotent; no-op on repeat.
	b.WriteString(InstallPersistenceCommand())
	b.WriteString(" && ")
	// Idempotent append.
	b.WriteString("( ")
	b.WriteString(r.RenderCheck())
	b.WriteString(" 2>/dev/null || ")
	b.WriteString(r.RenderAppend())
	b.WriteString(" )")
	b.WriteString(" && ")
	// Snapshot for reboot-persistence.
	b.WriteString(SaveRulesCommand())
	return b.String()
}

// buildRevokeCmd composes the idempotent revoke script.
func buildRevokeCmd(r AllowRule) string {
	var b strings.Builder
	// If the chain doesn't exist there's nothing to revoke.
	b.WriteString("iptables -S ")
	b.WriteString(ChainName)
	b.WriteString(" >/dev/null 2>&1 || exit 0")
	b.WriteString(" ; ")
	// Delete if present (no-op on missing).
	b.WriteString("( ")
	b.WriteString(r.RenderCheck())
	b.WriteString(" 2>/dev/null && ")
	b.WriteString(r.RenderDelete())
	b.WriteString(" || true")
	b.WriteString(" )")
	b.WriteString(" && ")
	b.WriteString(SaveRulesCommand())
	return b.String()
}

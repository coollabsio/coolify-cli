package wireguard

import (
	"fmt"
	"net"
	"strings"
)

const firewallUnitPath = "/etc/systemd/system/coolify-mesh-fw.service"
const firewallServiceName = "coolify-mesh-fw.service"

// FirewallServiceUnit returns the systemd unit text that installs the
// idempotent iptables rules required for cross-host container traffic over WG.
//
// Two modes:
//
//   - defaultDeny == false (mode A, blanket allow): installs FORWARD ACCEPT
//     rules for the container subnet. Tears down any default-deny scaffold
//     left over from a prior --default-deny run.
//
//   - defaultDeny == true  (mode B, default deny): removes blanket ACCEPT,
//     installs COOLIFY-INTRA + COOLIFY-ALLOW chains, and adds FORWARD jumps
//     so any traffic with the container subnet as source OR destination
//     traverses the deny chain. Conntrack ESTABLISHED/RELATED is accepted
//     early so reply traffic for already-allowed flows bypasses the chain.
//
// Note: default-deny only enforces CROSS-HOST container traffic (which
// crosses wg0 ↔ bridge interfaces and so traverses iptables FORWARD).
// Intra-host (same bridge) traffic stays at L2 and bypasses iptables in
// Linux + netavark — use per-app podman networks for intra-host isolation.
//
// Both modes preserve the POSTROUTING RETURN rule that prevents podman's
// MASQUERADE from rewriting container egress to wg0's IP.
func FirewallServiceUnit(iface string, containerSubnet *net.IPNet, defaultDeny bool) string {
	subnet := containerSubnet.String()
	var b strings.Builder

	fmt.Fprintf(&b, `[Unit]
Description=Coolify mesh firewall rules
After=wg-quick@%[1]s.service network-online.target
Wants=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes

# POSTROUTING RETURN — needed in both modes (skip podman MASQUERADE on wg0).
ExecStart=/bin/sh -c "/usr/sbin/iptables -t nat -C POSTROUTING -s %[2]s -o %[1]s -j RETURN 2>/dev/null || /usr/sbin/iptables -t nat -I POSTROUTING -s %[2]s -o %[1]s -j RETURN"
`, iface, subnet)

	if !defaultDeny {
		// Mode A: tear down any default-deny scaffold + install blanket ACCEPT.
		fmt.Fprintf(&b, `# Tear down default-deny scaffold from prior --default-deny run.
ExecStart=/bin/sh -c "/usr/sbin/iptables -D FORWARD -d %[2]s -j COOLIFY-INTRA 2>/dev/null || true"
ExecStart=/bin/sh -c "/usr/sbin/iptables -D FORWARD -s %[2]s -j COOLIFY-INTRA 2>/dev/null || true"
ExecStart=/bin/sh -c "/usr/sbin/iptables -F COOLIFY-INTRA 2>/dev/null || true"
ExecStart=/bin/sh -c "/usr/sbin/iptables -X COOLIFY-INTRA 2>/dev/null || true"
# COOLIFY-ALLOW intentionally NOT removed — preserves runtime allows for re-enable.

# Blanket ACCEPT — allow all traffic to/from the container subnet.
ExecStart=/bin/sh -c "/usr/sbin/iptables -C FORWARD -s %[2]s -j ACCEPT 2>/dev/null || /usr/sbin/iptables -I FORWARD -s %[2]s -j ACCEPT"
ExecStart=/bin/sh -c "/usr/sbin/iptables -C FORWARD -d %[2]s -j ACCEPT 2>/dev/null || /usr/sbin/iptables -I FORWARD -d %[2]s -j ACCEPT"
`, iface, subnet)
	} else {
		// Mode B: remove blanket ACCEPT + install default-deny scaffold.
		fmt.Fprintf(&b, `# Remove blanket ACCEPT from prior mode-A run.
ExecStart=/bin/sh -c "/usr/sbin/iptables -D FORWARD -s %[2]s -j ACCEPT 2>/dev/null || true"
ExecStart=/bin/sh -c "/usr/sbin/iptables -D FORWARD -d %[2]s -j ACCEPT 2>/dev/null || true"

# Create chains (idempotent).
ExecStart=/bin/sh -c "/usr/sbin/iptables -N COOLIFY-ALLOW 2>/dev/null || true"
ExecStart=/bin/sh -c "/usr/sbin/iptables -N COOLIFY-INTRA 2>/dev/null || true"

# Flush COOLIFY-INTRA so order is deterministic on every restart.
# COOLIFY-ALLOW is NOT flushed — preserves runtime allow rules added
# by the v5 control plane between service restarts.
ExecStart=/usr/sbin/iptables -F COOLIFY-INTRA
ExecStart=/usr/sbin/iptables -A COOLIFY-INTRA -j COOLIFY-ALLOW
ExecStart=/usr/sbin/iptables -A COOLIFY-INTRA -j DROP

# Conntrack early-accept at top of FORWARD (idempotent).
ExecStart=/bin/sh -c "/usr/sbin/iptables -C FORWARD -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT 2>/dev/null || /usr/sbin/iptables -I FORWARD 1 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT"

# Top-level FORWARD jumps for both directions of container subnet traffic.
ExecStart=/bin/sh -c "/usr/sbin/iptables -C FORWARD -d %[2]s -j COOLIFY-INTRA 2>/dev/null || /usr/sbin/iptables -A FORWARD -d %[2]s -j COOLIFY-INTRA"
ExecStart=/bin/sh -c "/usr/sbin/iptables -C FORWARD -s %[2]s -j COOLIFY-INTRA 2>/dev/null || /usr/sbin/iptables -A FORWARD -s %[2]s -j COOLIFY-INTRA"
`, iface, subnet)
	}

	b.WriteString(`
[Install]
WantedBy=multi-user.target
`)

	return b.String()
}

// InstallFirewallCommand returns a shell command that atomically writes the
// service unit, reloads systemd, and enables/starts (or restarts) it.
func InstallFirewallCommand(iface string, containerSubnet *net.IPNet, defaultDeny bool) string {
	unit := FirewallServiceUnit(iface, containerSubnet, defaultDeny)
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`cat > %s.tmp <<'COOLIFY_FW_EOF'
%sCOOLIFY_FW_EOF
mv %s.tmp %s && `, firewallUnitPath, unit, firewallUnitPath, firewallUnitPath))
	b.WriteString(`systemctl daemon-reload && `)
	// Use restart so a flag flip re-runs ExecStart= even if the unit is
	// already active (Type=oneshot with RemainAfterExit=yes blocks plain
	// "start" from running again).
	b.WriteString(fmt.Sprintf(`systemctl enable %s && systemctl restart %s`, firewallServiceName, firewallServiceName))
	return b.String()
}

package wireguard

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

const firewallUnitPath = "/etc/systemd/system/coolify-mesh-fw.service"
const firewallServiceName = "coolify-mesh-fw.service"

// AllowRulesPath is the on-disk location where coold snapshots the
// COOLIFY-ALLOW chain as an iptables-restore fragment on every rule mutate.
// The firewall unit reads this file at boot/restart to repopulate the chain
// after the kernel tables are cleared.
const AllowRulesPath = "/etc/coolify/allow.rules"

// BridgeTableName is the nftables table name owned by the CLI scaffold.
const BridgeTableName = "coolify_bridge"

// BridgeAllowRulesPath is where coold writes the nft bridge-family allow
// fragment. The firewall unit replays it at start/restart.
const BridgeAllowRulesPath = "/etc/coolify/allow.nft"

// BridgeScaffoldPath is where the CLI writes the static bridge chain
// scaffold (forward + coolify_intra chains). Applied at unit start/restart.
const BridgeScaffoldPath = "/etc/coolify/bridge-fw.nft"

// FirewallServiceUnit returns the systemd unit text that installs the
// idempotent iptables rules required for cross-host container traffic over WG.
//
// containerSubnets is the per-namespace list of subnets on this host (one
// /<prefix> per namespace). Rules are emitted once per subnet so every
// namespace is covered by the same host-global COOLIFY-INTRA / COOLIFY-ALLOW
// chain pair.
//
// Two modes:
//
//   - defaultDeny == false (mode A, blanket allow): installs FORWARD ACCEPT
//     rules for every subnet. Tears down any default-deny scaffold left over
//     from a prior --default-deny run.
//
//   - defaultDeny == true  (mode B, default deny): removes blanket ACCEPT,
//     installs COOLIFY-INTRA + COOLIFY-ALLOW chains, and adds FORWARD jumps
//     so any traffic with a container subnet as source OR destination
//     traverses the deny chain. Conntrack ESTABLISHED/RELATED is accepted
//     early so reply traffic for already-allowed flows bypasses the chain.
//
// Note: default-deny only enforces CROSS-HOST container traffic. Same-
// namespace intra-host traffic stays at L2 and bypasses iptables; cross-
// namespace intra-host traffic is blocked at L2 anyway because each namespace
// has its own podman bridge.
//
// Both modes preserve the POSTROUTING RETURN rule that prevents podman's
// MASQUERADE from rewriting container egress to wg0's IP.
func FirewallServiceUnit(iface string, namespaces []string, containerSubnets []*net.IPNet, defaultDeny bool) string {
	var b strings.Builder

	fmt.Fprintf(&b, `[Unit]
Description=Coolify mesh firewall rules
After=wg-quick@%[1]s.service network-online.target
Wants=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes

`, iface)

	// POSTROUTING RETURN — needed in both modes, once per subnet.
	for _, sn := range containerSubnets {
		fmt.Fprintf(&b,
			`ExecStart=/bin/sh -c "/usr/sbin/iptables -t nat -C POSTROUTING -s %[2]s -o %[1]s -j RETURN 2>/dev/null || /usr/sbin/iptables -t nat -I POSTROUTING -s %[2]s -o %[1]s -j RETURN"
`, iface, sn.String())
	}

	if !defaultDeny {
		fmt.Fprint(&b, `# Tear down default-deny scaffold from prior --default-deny run.
`)
		for _, sn := range containerSubnets {
			fmt.Fprintf(&b,
				`ExecStart=/bin/sh -c "/usr/sbin/iptables -D FORWARD -d %[1]s -j COOLIFY-INTRA 2>/dev/null || true"
ExecStart=/bin/sh -c "/usr/sbin/iptables -D FORWARD -s %[1]s -j COOLIFY-INTRA 2>/dev/null || true"
`, sn.String())
		}
		fmt.Fprintf(&b, `ExecStart=/bin/sh -c "/usr/sbin/iptables -F COOLIFY-INTRA 2>/dev/null || true"
ExecStart=/bin/sh -c "/usr/sbin/iptables -X COOLIFY-INTRA 2>/dev/null || true"
# COOLIFY-ALLOW intentionally NOT removed — preserves runtime allows for re-enable.
# Remove bridge-family scaffold (permissive mode) before installing blanket ACCEPT.
ExecStart=/bin/sh -c "nft delete table bridge %[1]s 2>/dev/null || true"

# Blanket ACCEPT — allow all traffic to/from every namespace's container subnet.
`, BridgeTableName)
		for _, sn := range containerSubnets {
			fmt.Fprintf(&b,
				`ExecStart=/bin/sh -c "/usr/sbin/iptables -C FORWARD -s %[1]s -j ACCEPT 2>/dev/null || /usr/sbin/iptables -I FORWARD -s %[1]s -j ACCEPT"
ExecStart=/bin/sh -c "/usr/sbin/iptables -C FORWARD -d %[1]s -j ACCEPT 2>/dev/null || /usr/sbin/iptables -I FORWARD -d %[1]s -j ACCEPT"
`, sn.String())
		}
	} else {
		fmt.Fprint(&b, `# Remove blanket ACCEPT from prior mode-A run.
`)
		for _, sn := range containerSubnets {
			fmt.Fprintf(&b,
				`ExecStart=/bin/sh -c "/usr/sbin/iptables -D FORWARD -s %[1]s -j ACCEPT 2>/dev/null || true"
ExecStart=/bin/sh -c "/usr/sbin/iptables -D FORWARD -d %[1]s -j ACCEPT 2>/dev/null || true"
`, sn.String())
		}
		fmt.Fprintf(&b, `
# Create chains (idempotent).
ExecStart=/bin/sh -c "/usr/sbin/iptables -N COOLIFY-ALLOW 2>/dev/null || true"
ExecStart=/bin/sh -c "/usr/sbin/iptables -N COOLIFY-INTRA 2>/dev/null || true"

# Flush COOLIFY-INTRA so order is deterministic on every restart.
ExecStart=/usr/sbin/iptables -F COOLIFY-INTRA
ExecStart=/usr/sbin/iptables -A COOLIFY-INTRA -j COOLIFY-ALLOW
ExecStart=/usr/sbin/iptables -A COOLIFY-INTRA -j DROP

# Repopulate COOLIFY-ALLOW from coold's canonical snapshot. File is rewritten
# by coold on every rule mutate, so it is the source of truth across reboots
# and service restarts. Flush first because 'iptables-restore --noflush'
# leaves existing chain contents in place and would otherwise duplicate every
# rule on re-run.
ExecStart=/bin/sh -c "[ -s %[1]s ] && /usr/sbin/iptables -F COOLIFY-ALLOW && /usr/sbin/iptables-restore --noflush < %[1]s || true"

# Conntrack early-accept at top of FORWARD (idempotent).
ExecStart=/bin/sh -c "/usr/sbin/iptables -C FORWARD -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT 2>/dev/null || /usr/sbin/iptables -I FORWARD 1 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT"

# Top-level FORWARD jumps for every namespace's subnet (both directions).
`, AllowRulesPath)
		for _, sn := range containerSubnets {
			fmt.Fprintf(&b,
				`ExecStart=/bin/sh -c "/usr/sbin/iptables -C FORWARD -d %[1]s -j COOLIFY-INTRA 2>/dev/null || /usr/sbin/iptables -A FORWARD -d %[1]s -j COOLIFY-INTRA"
ExecStart=/bin/sh -c "/usr/sbin/iptables -C FORWARD -s %[1]s -j COOLIFY-INTRA 2>/dev/null || /usr/sbin/iptables -A FORWARD -s %[1]s -j COOLIFY-INTRA"
`, sn.String())
		}
		fmt.Fprintf(&b, `# Bridge-family nft scaffold — intra-namespace default-deny.
ExecStart=/bin/sh -c "nft list table bridge %[1]s >/dev/null 2>&1 || nft add table bridge %[1]s"
ExecStart=/bin/sh -c "nft add chain bridge %[1]s coolify_allow '{ }' 2>/dev/null || true"
ExecStart=/bin/sh -c "nft delete chain bridge %[1]s forward 2>/dev/null || true"
ExecStart=/bin/sh -c "nft delete chain bridge %[1]s coolify_intra 2>/dev/null || true"
ExecStart=/bin/sh -c "nft -f %[2]s"
ExecStart=/bin/sh -c "[ -s %[3]s ] && nft -f %[3]s || true"
`, BridgeTableName, BridgeScaffoldPath, BridgeAllowRulesPath)
	}

	b.WriteString(`
[Install]
WantedBy=multi-user.target
`)

	return b.String()
}

// InstallFirewallCommand returns a shell command that atomically writes the
// service unit, reloads systemd, and enables/starts (or restarts) it.
func InstallFirewallCommand(iface string, namespaces []string, containerSubnets []*net.IPNet, defaultDeny bool) string {
	unit := FirewallServiceUnit(iface, namespaces, containerSubnets, defaultDeny)
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`cat > %s.tmp <<'COOLIFY_FW_EOF'
%sCOOLIFY_FW_EOF
mv %s.tmp %s && `, firewallUnitPath, unit, firewallUnitPath, firewallUnitPath))
	if defaultDeny {
		scaffold := renderBridgeScaffold(namespaces)
		b.WriteString(fmt.Sprintf(`cat > %s.tmp <<'COOLIFY_BR_EOF'
%sCOOLIFY_BR_EOF
mv %s.tmp %s && `, BridgeScaffoldPath, scaffold, BridgeScaffoldPath, BridgeScaffoldPath))
	} else {
		b.WriteString(fmt.Sprintf("rm -f %s && ", BridgeScaffoldPath))
	}
	b.WriteString(`systemctl daemon-reload && `)
	// Use restart so a flag flip re-runs ExecStart= even if the unit is
	// already active (Type=oneshot with RemainAfterExit=yes blocks plain
	// "start" from running again).
	b.WriteString(fmt.Sprintf(`systemctl enable %s && systemctl restart %s`, firewallServiceName, firewallServiceName))
	return b.String()
}

// renderBridgeScaffold builds the nft file-format content for the bridge
// scaffold. Uses `add table` + `add chain` (idempotent) then `flush chain` +
// `add rule` so forward and coolify_intra are atomically replaced on every
// apply without touching coolify_allow (owned by coold).
func renderBridgeScaffold(namespaces []string) string {
	sorted := make([]string, len(namespaces))
	copy(sorted, namespaces)
	sort.Strings(sorted)

	// Build quoted iifname/oifname set: { "coolify-ns1-mesh", "coolify-ns2-mesh" }
	ifNames := make([]string, 0, len(sorted))
	for _, ns := range sorted {
		ifNames = append(ifNames, fmt.Sprintf(`"coolify-%s-mesh"`, ns))
	}
	ifaceSet := "{ " + strings.Join(ifNames, ", ") + " }"

	var b strings.Builder
	b.WriteString("# Managed by coolify init — do not edit manually.\n")
	b.WriteString("# Replaces forward + coolify_intra chains on restart; never touches coolify_allow.\n")
	fmt.Fprintf(&b, "add table bridge %s\n", BridgeTableName)
	fmt.Fprintf(&b, "add chain bridge %s forward { type filter hook forward priority -200; policy accept; }\n", BridgeTableName)
	fmt.Fprintf(&b, "flush chain bridge %s forward\n", BridgeTableName)
	fmt.Fprintf(&b, "add rule bridge %s forward meta protocol != ip accept\n", BridgeTableName)
	fmt.Fprintf(&b, "add rule bridge %s forward ct state established,related accept\n", BridgeTableName)
	fmt.Fprintf(&b, "add rule bridge %s forward iifname %s jump coolify_intra\n", BridgeTableName, ifaceSet)
	fmt.Fprintf(&b, "add rule bridge %s forward oifname %s jump coolify_intra\n", BridgeTableName, ifaceSet)
	fmt.Fprintf(&b, "add chain bridge %s coolify_intra\n", BridgeTableName)
	fmt.Fprintf(&b, "flush chain bridge %s coolify_intra\n", BridgeTableName)
	fmt.Fprintf(&b, "add rule bridge %s coolify_intra jump coolify_allow\n", BridgeTableName)
	fmt.Fprintf(&b, "add rule bridge %s coolify_intra drop\n", BridgeTableName)
	return b.String()
}

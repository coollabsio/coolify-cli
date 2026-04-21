package wireguard

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirewallServiceUnit_DefaultDenyOff(t *testing.T) {
	subnets := []*net.IPNet{mustParseCIDR("10.210.0.0/24")}
	got := FirewallServiceUnit("wg0", []string{"default"}, subnets, false)

	assert.Contains(t, got, "[Unit]")
	assert.Contains(t, got, "Description=Coolify mesh firewall rules")
	assert.Contains(t, got, "After=wg-quick@wg0.service")
	assert.Contains(t, got, "Type=oneshot")
	assert.Contains(t, got, "RemainAfterExit=yes")

	// Blanket allow rules present.
	assert.Contains(t, got, "/usr/sbin/iptables -I FORWARD -s 10.210.0.0/24 -j ACCEPT")
	assert.Contains(t, got, "/usr/sbin/iptables -I FORWARD -d 10.210.0.0/24 -j ACCEPT")

	// Teardown of default-deny scaffold present (idempotent cleanup).
	assert.Contains(t, got, "/usr/sbin/iptables -X COOLIFY-INTRA")
	assert.Contains(t, got, "/usr/sbin/iptables -D FORWARD -s 10.210.0.0/24 -j COOLIFY-INTRA")
	assert.Contains(t, got, "/usr/sbin/iptables -D FORWARD -d 10.210.0.0/24 -j COOLIFY-INTRA")

	// Default-deny chain rules MUST NOT be present.
	assert.NotContains(t, got, "-A COOLIFY-INTRA -j COOLIFY-ALLOW")
	assert.NotContains(t, got, "-A COOLIFY-INTRA -j DROP")

	// COOLIFY-ALLOW chain is never destroyed.
	assert.NotContains(t, got, "-X COOLIFY-ALLOW")

	// POSTROUTING RETURN preserved (needed in both modes).
	assert.Contains(t, got, "/usr/sbin/iptables -t nat -I POSTROUTING -s 10.210.0.0/24 -o wg0 -j RETURN")

	assert.Contains(t, got, "WantedBy=multi-user.target")
}

func TestFirewallServiceUnit_DefaultDenyOn(t *testing.T) {
	subnets := []*net.IPNet{mustParseCIDR("10.210.0.0/24")}
	got := FirewallServiceUnit("wg0", []string{"default"}, subnets, true)

	// Chains created.
	assert.Contains(t, got, "/usr/sbin/iptables -N COOLIFY-ALLOW")
	assert.Contains(t, got, "/usr/sbin/iptables -N COOLIFY-INTRA")

	// Conntrack early-accept.
	assert.Contains(t, got, "-m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT")

	// COOLIFY-INTRA flush + jump to ALLOW + DROP.
	assert.Contains(t, got, "/usr/sbin/iptables -F COOLIFY-INTRA")
	assert.Contains(t, got, "/usr/sbin/iptables -A COOLIFY-INTRA -j COOLIFY-ALLOW")
	assert.Contains(t, got, "/usr/sbin/iptables -A COOLIFY-INTRA -j DROP")

	// FORWARD jumps for both directions of container subnet traffic.
	assert.Contains(t, got, "/usr/sbin/iptables -A FORWARD -d 10.210.0.0/24 -j COOLIFY-INTRA")
	assert.Contains(t, got, "/usr/sbin/iptables -A FORWARD -s 10.210.0.0/24 -j COOLIFY-INTRA")

	// Teardown of blanket ACCEPT from prior mode-A run.
	assert.Contains(t, got, "/usr/sbin/iptables -D FORWARD -s 10.210.0.0/24 -j ACCEPT")
	assert.Contains(t, got, "/usr/sbin/iptables -D FORWARD -d 10.210.0.0/24 -j ACCEPT")

	// Blanket ACCEPT rules MUST NOT be installed in default-deny mode.
	assert.NotContains(t, got, "/usr/sbin/iptables -I FORWARD -s 10.210.0.0/24 -j ACCEPT")
	assert.NotContains(t, got, "/usr/sbin/iptables -I FORWARD -d 10.210.0.0/24 -j ACCEPT")

	// COOLIFY-ALLOW chain is never destroyed. It IS flushed-and-restored at
	// boot/restart from the canonical snapshot — that's how runtime allow
	// rules survive reboots.
	assert.NotContains(t, got, "-X COOLIFY-ALLOW")
	assert.Contains(t, got, "/usr/sbin/iptables -F COOLIFY-ALLOW")
	assert.Contains(t, got, "/usr/sbin/iptables-restore --noflush < "+AllowRulesPath)
	assert.Contains(t, got, "[ -s "+AllowRulesPath+" ]")

	// POSTROUTING RETURN preserved.
	assert.Contains(t, got, "/usr/sbin/iptables -t nat -I POSTROUTING -s 10.210.0.0/24 -o wg0 -j RETURN")
}

func TestFirewallServiceUnit_DefaultDenyOff_NoAllowRestore(t *testing.T) {
	subnets := []*net.IPNet{mustParseCIDR("10.210.0.0/24")}
	got := FirewallServiceUnit("wg0", []string{"default"}, subnets, false)

	// Blanket-allow mode bypasses COOLIFY-ALLOW entirely — no restore.
	assert.NotContains(t, got, "iptables-restore")
	assert.NotContains(t, got, AllowRulesPath)
}

func TestInstallFirewallCommand_AtomicWriteAndEnable(t *testing.T) {
	subnets := []*net.IPNet{mustParseCIDR("10.210.5.0/24")}
	cmd := InstallFirewallCommand("wg0", []string{"default"}, subnets, false)

	// Atomic write via .tmp + mv.
	assert.Contains(t, cmd, "/etc/systemd/system/coolify-mesh-fw.service.tmp")
	assert.Contains(t, cmd, "mv /etc/systemd/system/coolify-mesh-fw.service.tmp /etc/systemd/system/coolify-mesh-fw.service")

	// systemd reload + enable + restart (so a flag flip re-runs ExecStart).
	assert.Contains(t, cmd, "systemctl daemon-reload")
	assert.Contains(t, cmd, "systemctl enable coolify-mesh-fw.service")
	assert.Contains(t, cmd, "systemctl restart coolify-mesh-fw.service")

	// Subnet baked into command.
	assert.Contains(t, cmd, "10.210.5.0/24")
}

func TestInstallFirewallCommand_DefaultDenyEmbedded(t *testing.T) {
	subnets := []*net.IPNet{mustParseCIDR("10.210.5.0/24")}
	cmd := InstallFirewallCommand("wg0", []string{"default"}, subnets, true)

	// Default-deny variant of unit must be embedded in the heredoc.
	assert.Contains(t, cmd, "-A COOLIFY-INTRA -j DROP")
}

func TestFirewallServiceUnit_BridgeScaffold_DefaultDenyOn(t *testing.T) {
	subnets := []*net.IPNet{mustParseCIDR("10.210.0.0/24")}
	got := FirewallServiceUnit("wg0", []string{"default"}, subnets, true)

	assert.Contains(t, got, "nft list table bridge coolify_bridge")
	assert.Contains(t, got, "nft add table bridge coolify_bridge")
	assert.Contains(t, got, "nft add chain bridge coolify_bridge coolify_allow")
	assert.Contains(t, got, "nft delete chain bridge coolify_bridge forward")
	assert.Contains(t, got, "nft delete chain bridge coolify_bridge coolify_intra")
	assert.Contains(t, got, "nft -f /etc/coolify/bridge-fw.nft")
	assert.Contains(t, got, "/etc/coolify/allow.nft")
	assert.NotContains(t, got, "-X COOLIFY-ALLOW")
}

func TestFirewallServiceUnit_BridgeScaffold_DefaultDenyOff(t *testing.T) {
	subnets := []*net.IPNet{mustParseCIDR("10.210.0.0/24")}
	got := FirewallServiceUnit("wg0", []string{"default"}, subnets, false)

	assert.Contains(t, got, "nft delete table bridge coolify_bridge")
	assert.NotContains(t, got, "nft add table bridge coolify_bridge")
	assert.NotContains(t, got, "nft -f /etc/coolify/bridge-fw.nft")
}

func TestFirewallServiceUnit_BridgeSetStableSortedSubnets(t *testing.T) {
	// Pass subnets in reverse-sorted order — scaffold must sort them.
	subnets := []*net.IPNet{
		mustParseCIDR("10.220.1.0/24"),
		mustParseCIDR("10.210.1.0/24"),
	}
	// renderBridgeScaffold is embedded in InstallFirewallCommand, so check that.
	cmd := InstallFirewallCommand("wg0", []string{"alpha", "default"}, subnets, true)

	// Assert the nft scaffold set contains both, sorted:
	//   `ip saddr { 10.210.1.0/24, 10.220.1.0/24 } jump coolify_intra`
	assert.Contains(t, cmd, "ip saddr { 10.210.1.0/24, 10.220.1.0/24 } jump coolify_intra")
	assert.Contains(t, cmd, "ip daddr { 10.210.1.0/24, 10.220.1.0/24 } jump coolify_intra")
}

func TestFirewallServiceUnit_BridgeScaffold_UsesIPSaddrNotIifname(t *testing.T) {
	subnets := []*net.IPNet{mustParseCIDR("10.210.0.0/24")}
	cmd := InstallFirewallCommand("wg0", []string{"default"}, subnets, true)

	// Podman bridge names exceed IFNAMSIZ=16 (e.g. "coolify-default-mesh" = 20
	// chars). Scaffold MUST key dispatch on ip saddr/daddr, never iifname.
	assert.Contains(t, cmd, "ip saddr")
	assert.Contains(t, cmd, "ip daddr")
	assert.NotContains(t, cmd, "iifname")
	assert.NotContains(t, cmd, "oifname")
	assert.NotContains(t, cmd, "coolify-default-mesh\"")
}

func TestInstallFirewallCommand_WritesBridgeScaffoldFile(t *testing.T) {
	subnets := []*net.IPNet{mustParseCIDR("10.210.0.0/24")}
	cmd := InstallFirewallCommand("wg0", []string{"default"}, subnets, true)

	assert.Contains(t, cmd, "/etc/coolify/bridge-fw.nft")
	assert.Contains(t, cmd, "COOLIFY_BR_EOF")
	assert.Contains(t, cmd, "bridge-fw.nft.tmp")

	// /etc/coolify must be created before bridge-fw.nft.tmp is written —
	// without it, `cat > .tmp` fails on fresh hosts.
	mkdirIdx := strings.Index(cmd, "mkdir -p /etc/coolify")
	tmpIdx := strings.Index(cmd, "bridge-fw.nft.tmp")
	assert.True(t, mkdirIdx >= 0, "mkdir -p /etc/coolify must be present")
	assert.True(t, mkdirIdx < tmpIdx, "mkdir must run before bridge-fw.nft.tmp write")
}

func TestInstallFirewallCommand_DefaultDenyOff_RemovesBridgeScaffold(t *testing.T) {
	subnets := []*net.IPNet{mustParseCIDR("10.210.0.0/24")}
	cmd := InstallFirewallCommand("wg0", []string{"default"}, subnets, false)

	assert.Contains(t, cmd, "rm -f /etc/coolify/bridge-fw.nft")
	assert.NotContains(t, cmd, "COOLIFY_BR_EOF")
}

func TestFirewallServiceUnit_GoldenFixture_TwoNamespaces(t *testing.T) {
	subnets := []*net.IPNet{
		mustParseCIDR("10.210.0.0/24"),
		mustParseCIDR("10.220.0.0/24"),
	}
	got := FirewallServiceUnit("wg0", []string{"alpha", "default"}, subnets, true)

	fixturePath := filepath.Join("..", "..", "test", "fixtures", "firewall_unit_deny_two_ns.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		err := os.WriteFile(fixturePath, []byte(got), 0o644)
		require.NoError(t, err, "failed to write golden fixture")
		t.Logf("golden fixture updated: %s", fixturePath)
		return
	}

	b, err := os.ReadFile(fixturePath)
	require.NoError(t, err, "golden fixture missing — run with UPDATE_GOLDEN=1 to create it")
	assert.Equal(t, string(b), got)
}

func TestFirewallServiceUnit_MultipleNamespacesEmitPerSubnetRules(t *testing.T) {
	subnets := []*net.IPNet{
		mustParseCIDR("10.210.1.0/24"),
		mustParseCIDR("10.220.1.0/24"),
	}
	got := FirewallServiceUnit("wg0", []string{"default"}, subnets, true)

	// Each namespace subnet gets its own POSTROUTING RETURN + FORWARD jumps.
	for _, sub := range []string{"10.210.1.0/24", "10.220.1.0/24"} {
		assert.Contains(t, got, "/usr/sbin/iptables -t nat -I POSTROUTING -s "+sub+" -o wg0 -j RETURN")
		assert.Contains(t, got, "/usr/sbin/iptables -A FORWARD -d "+sub+" -j COOLIFY-INTRA")
		assert.Contains(t, got, "/usr/sbin/iptables -A FORWARD -s "+sub+" -j COOLIFY-INTRA")
	}
}

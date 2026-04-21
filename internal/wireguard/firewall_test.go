package wireguard

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
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

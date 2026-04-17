package services

import (
	"fmt"
	"net"
)

// DefaultCooldDNSZone is the DNS zone served by coold's embedded resolver.
// `.internal` is RFC 6761 reserved — safe from public-TLD collisions.
const DefaultCooldDNSZone = "coolify.internal"

// CooldServiceUnit returns the systemd unit text for coold.
//
// mgmtIP is this host's wg0 management IP (coold writes rows scoped to it and
// binds its REST API to mgmtIP:8443).
//
// bridgeGatewayIP is the .1 of this host's container subnet — coold's embedded
// DNS server (CONTROL_PLANE.md §5) binds UDP/TCP :53 here. Pass nil to skip
// DNS env injection (e.g. in tests that don't care about DNS).
func CooldServiceUnit(mgmtIP, bridgeGatewayIP net.IP) string {
	// Wants (not Requires) on corrosion: if corrosion crashes/restarts we want
	// coold to stay up and retry — reconcile_once already backs off for 1s on
	// error, so it self-heals once corrosion is back. Requires would cascade
	// stop coold and leave it down until someone restarted it.
	dnsEnv := ""
	if bridgeGatewayIP != nil {
		dnsEnv = fmt.Sprintf(`Environment=COOLD_BRIDGE_GATEWAY_IP=%s
Environment=COOLD_DNS_ZONE=%s
`, bridgeGatewayIP, DefaultCooldDNSZone)
	}
	return fmt.Sprintf(`[Unit]
Description=Coolify host agent
Wants=corrosion.service
After=corrosion.service network-online.target podman.socket

[Service]
Environment=COOLD_HOST_MGMT_IP=%s
%sExecStart=/usr/local/bin/coold
Restart=on-failure
RestartSec=2s

[Install]
WantedBy=multi-user.target
`, mgmtIP, dnsEnv)
}

package services

import (
	"fmt"
	"net"
)

// DefaultCooldDNSZone is the DNS zone served by coold's embedded resolver.
// `.internal` is RFC 6761 reserved — safe from public-TLD collisions.
const DefaultCooldDNSZone = "coolify.internal"

// CooldAPIPort is the TCP port coold's firewall REST API binds on wg0.
const CooldAPIPort = 8443

// CooldAPITokenPath is the on-host path where coold reads the bearer token
// for the firewall REST API. The file is generated once by `coolify init
// apply --install-coold` (random 32-byte hex via `openssl rand`) and kept
// mode 0600.
const CooldAPITokenPath = "/etc/coolify/api-token"

// CooldServiceUnit returns the systemd unit text for coold.
//
// mgmtIP is this host's wg0 management IP (coold writes rows scoped to it and
// binds its REST API to mgmtIP:CooldAPIPort).
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
	// Firewall REST API binds wg0-only (never a public interface) and requires
	// a bearer token. Plain HTTP for alpha — TLS material is managed by the
	// central Coolify control plane and will be wired in a follow-up.
	apiEnv := fmt.Sprintf(`Environment=COOLD_API_BIND=%s:%d
Environment=COOLD_API_TOKEN_FILE=%s
`, mgmtIP, CooldAPIPort, CooldAPITokenPath)
	return fmt.Sprintf(`[Unit]
Description=Coolify host agent
Wants=corrosion.service
After=corrosion.service network-online.target podman.socket

[Service]
Environment=COOLD_HOST_MGMT_IP=%s
%s%sExecStart=/usr/local/bin/coold
AmbientCapabilities=CAP_NET_BIND_SERVICE CAP_NET_ADMIN CAP_NET_RAW
Restart=on-failure
RestartSec=2s

[Install]
WantedBy=multi-user.target
`, mgmtIP, dnsEnv, apiEnv)
}

// EnsureCooldAPITokenCommand returns a shell snippet that creates the
// CooldAPITokenPath file with a random 32-byte hex token if it does not
// already exist. Idempotent: repeated runs preserve the existing token so
// clients already trusting it keep working.
func EnsureCooldAPITokenCommand() string {
	return fmt.Sprintf(
		`mkdir -p /etc/coolify && `+
			`if [ ! -s %[1]s ]; then `+
			`openssl rand -hex 32 > %[1]s.tmp && `+
			`chmod 0600 %[1]s.tmp && `+
			`mv %[1]s.tmp %[1]s; `+
			`fi`,
		CooldAPITokenPath,
	)
}

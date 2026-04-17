package services

import (
	"fmt"
	"net"
)

// CooldServiceUnit returns the systemd unit text for coold.
// mgmtIP is this host's wg0 management IP (coold writes rows scoped to it).
func CooldServiceUnit(mgmtIP net.IP) string {
	// Wants (not Requires) on corrosion: if corrosion crashes/restarts we want
	// coold to stay up and retry — reconcile_once already backs off for 1s on
	// error, so it self-heals once corrosion is back. Requires would cascade
	// stop coold and leave it down until someone restarted it.
	return fmt.Sprintf(`[Unit]
Description=Coolify host agent
Wants=corrosion.service
After=corrosion.service network-online.target podman.socket

[Service]
Environment=COOLD_HOST_MGMT_IP=%s
ExecStart=/usr/local/bin/coold
Restart=on-failure
RestartSec=2s

[Install]
WantedBy=multi-user.target
`, mgmtIP)
}

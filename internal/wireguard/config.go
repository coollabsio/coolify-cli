package wireguard

import (
	"fmt"
	"net"
	"strings"
)

// PeerConfig holds the information needed to write a [Peer] block.
type PeerConfig struct {
	// Endpoint is the SSH/public IP of the peer (used as WG endpoint).
	Endpoint string
	// PublicKey is the peer's WireGuard public key.
	PublicKey string
	// MgmtIP is the peer's /32 wg0 management IP.
	MgmtIP net.IP
	// ContainerSubnet is the peer's /24 container bridge subnet.
	// Both MgmtIP/32 and ContainerSubnet are listed in AllowedIPs so
	// management and container traffic both route via the tunnel.
	ContainerSubnet *net.IPNet
}

// RenderConfig returns the content of wg0.conf for one host.
//
// The host's own Address is the management IP /32 (e.g. 100.64.0.0/32). It
// lives in a separate pool from the container subnet, so the Podman bridge
// can own the full per-host /24 without conflict.
//
// The literal string __PRIVKEY__ is used as a placeholder; callers must
// substitute the actual key before (or during) writing to disk.
func RenderConfig(mgmtIP net.IP, listenPort int, peers []PeerConfig) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[Interface]\n")
	fmt.Fprintf(&b, "Address = %s/32\n", mgmtIP)
	fmt.Fprintf(&b, "ListenPort = %d\n", listenPort)
	fmt.Fprintf(&b, "PrivateKey = __PRIVKEY__\n")

	for _, p := range peers {
		fmt.Fprintf(&b, "\n[Peer]\n")
		fmt.Fprintf(&b, "# %s\n", p.Endpoint)
		fmt.Fprintf(&b, "PublicKey = %s\n", p.PublicKey)
		fmt.Fprintf(&b, "AllowedIPs = %s/32, %s\n", p.MgmtIP, p.ContainerSubnet)
		fmt.Fprintf(&b, "Endpoint = %s:%d\n", p.Endpoint, listenPort)
		fmt.Fprintf(&b, "PersistentKeepalive = 25\n")
	}
	return b.String()
}

// WriteConfigCommand returns the shell command that atomically writes
// /etc/wireguard/<iface>.conf on the remote host.
//
// The private key is read from /etc/wireguard/privatekey on the remote so it
// never traverses SSH.  The config is written to a .tmp file first and then
// moved into place so a killed session cannot leave a torn config.
func WriteConfigCommand(iface string, mgmtIP net.IP, listenPort int, peers []PeerConfig) string {
	var b strings.Builder

	b.WriteString(`PRIVKEY=$(cat /etc/wireguard/privatekey) && `)
	b.WriteString(`mkdir -p /etc/wireguard && `)
	b.WriteString(fmt.Sprintf(`{ echo "[Interface]"; `))
	b.WriteString(fmt.Sprintf(`echo "Address = %s/32"; `, mgmtIP))
	b.WriteString(fmt.Sprintf(`echo "ListenPort = %d"; `, listenPort))
	b.WriteString(`echo "PrivateKey = $PRIVKEY"; `)

	for _, p := range peers {
		b.WriteString(`echo ""; `)
		b.WriteString(`echo "[Peer]"; `)
		b.WriteString(fmt.Sprintf(`echo "# %s"; `, p.Endpoint))
		b.WriteString(fmt.Sprintf(`echo "PublicKey = %s"; `, p.PublicKey))
		b.WriteString(fmt.Sprintf(`echo "AllowedIPs = %s/32, %s"; `, p.MgmtIP, p.ContainerSubnet))
		b.WriteString(fmt.Sprintf(`echo "Endpoint = %s:%d"; `, p.Endpoint, listenPort))
		b.WriteString(`echo "PersistentKeepalive = 25"; `)
	}

	b.WriteString(fmt.Sprintf(`} > /etc/wireguard/%s.conf.tmp && `, iface))
	b.WriteString(fmt.Sprintf(`mv /etc/wireguard/%s.conf.tmp /etc/wireguard/%s.conf`, iface, iface))

	return b.String()
}

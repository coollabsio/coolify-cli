// Package wireguard implements the WireGuard mesh bootstrap logic for
// the coolify init command (alpha, Coolify v5).
package wireguard

import "net"

// Peer represents a single WireGuard peer as seen in the config or
// from `wg show <iface> dump`.
type Peer struct {
	PublicKey           string
	PresharedKey        string // "(none)" when absent
	Endpoint            string // "ip:port" or empty
	AllowedIPs          []string
	LatestHandshake     int64 // Unix timestamp; 0 means no handshake yet
	PersistentKeepalive int   // seconds; 0 means disabled
}

// ServerState holds the reconstructed WireGuard + Podman state for one server.
// It is built from live SSH probes and never cached to disk.
type ServerState struct {
	// Host is the SSH address used to reach this server.
	// It also serves as the WireGuard Endpoint value for peer configs.
	Host string

	// Installed is true when the wireguard package is present.
	Installed bool

	// KeysExist is true when /etc/wireguard/privatekey exists.
	KeysExist bool

	// PublicKey is the content of /etc/wireguard/publickey (trimmed).
	// Empty when KeysExist is false.
	PublicKey string

	// WireGuardMgmtIP is the /32 management IP assigned to wg0 (parsed from
	// the [Interface] Address line). Lives outside the container pool so the
	// Podman bridge can own the full per-host /24 without conflict.
	// nil when not yet assigned.
	WireGuardMgmtIP net.IP

	// ContainerSubnet is the per-host /24 owned by the Podman bridge
	// (read from `podman network inspect`). nil when not yet created.
	ContainerSubnet *net.IPNet

	// ListenPort is the WireGuard listen port from the config.
	ListenPort int

	// Interface is the WireGuard interface name (e.g., "wg0").
	Interface string

	// Active is true when `wg show <iface>` returns output (interface up).
	Active bool

	// Peers lists the peers currently present in the config file.
	Peers []Peer

	// PodmanInstalled is true when the podman package is present.
	PodmanInstalled bool

	// PodmanSocketActive is true when podman.socket systemd unit is active.
	PodmanSocketActive bool

	// PodmanNetExists is true when the configured Podman network already exists.
	PodmanNetExists bool

	// IPForwardEnabled is true when net.ipv4.ip_forward == 1.
	IPForwardEnabled bool

	// FirewallActive is true when coolify-mesh-fw.service is active.
	FirewallActive bool

	// DefaultDenyActive is true when the COOLIFY-INTRA chain exists and
	// terminates in DROP (the default-deny scaffold is in place).
	DefaultDenyActive bool
}

// MeshState is the reconstructed state across all servers in the mesh.
type MeshState struct {
	// Servers maps host → *ServerState.
	Servers map[string]*ServerState
}

// AssignedMgmtIPs returns a map of host → net.IP for all servers that
// already have a WG management IP assigned.
func (m *MeshState) AssignedMgmtIPs() map[string]net.IP {
	out := make(map[string]net.IP, len(m.Servers))
	for host, s := range m.Servers {
		if s.WireGuardMgmtIP != nil {
			out[host] = s.WireGuardMgmtIP
		}
	}
	return out
}

// AssignedContainerSubnets returns a map of host → *net.IPNet for all servers
// that already have a container subnet assigned (via the Podman bridge).
func (m *MeshState) AssignedContainerSubnets() map[string]*net.IPNet {
	out := make(map[string]*net.IPNet, len(m.Servers))
	for host, s := range m.Servers {
		if s.ContainerSubnet != nil {
			out[host] = s.ContainerSubnet
		}
	}
	return out
}

// DesiredMesh describes the target WireGuard + Podman configuration.
type DesiredMesh struct {
	// Hosts lists the SSH addresses of all servers (also used as WG endpoints).
	Hosts []string

	// Interface is the WireGuard interface name (default "wg0").
	Interface string

	// MgmtPool is the address pool from which per-host /32 management IPs
	// are carved and assigned to wg0 (default 100.64.0.0/16 — RFC 6598 CGNAT).
	MgmtPool *net.IPNet

	// ContainerPool is the address pool from which per-host container subnets
	// are carved and owned by the Podman bridge (default 10.210.0.0/16).
	ContainerPool *net.IPNet

	// ContainerPrefix is the prefix length of each per-host container subnet
	// (default 24, giving each host 254 usable container IPs).
	ContainerPrefix int

	// ListenPort is the WireGuard UDP listen port (default 51820).
	ListenPort int

	// InstallPodman, when true, installs Podman, enables its socket, creates
	// the per-host bridge network, installs firewall rules, and enables IP
	// forwarding on each server.
	InstallPodman bool

	// PodmanNetworkName is the name of the Podman bridge network to create
	// on each host (default "coolify-mesh").
	PodmanNetworkName string

	// DefaultDenyContainers, when true (and InstallPodman is true), installs
	// default-deny iptables rules for ALL container traffic on the host's
	// container subnet (intra-host AND cross-host via wg0). The v5 control
	// plane manages the explicit allow-list in the COOLIFY-ALLOW chain.
	DefaultDenyContainers bool
}

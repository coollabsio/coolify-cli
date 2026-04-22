// Package wireguard implements the WireGuard mesh bootstrap logic for
// the coolify init command (alpha, Coolify v5).
package wireguard

import (
	"net"
	"sort"
)

// DefaultNamespace is the namespace used when the user does not pass
// --namespaces. It is also always present even in a multi-namespace setup —
// coold's config assumes a `default` entry.
const DefaultNamespace = "default"

// PodmanNetworkFor returns the podman bridge name backing namespace ns on
// every host. Derived as `coolify-<ns>-mesh` so the namespace is visible
// directly in `podman network ls`.
func PodmanNetworkFor(ns string) string {
	return "coolify-" + ns + "-mesh"
}

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

// NamespaceServerState captures per-namespace podman state on one host. A
// ServerState carries one entry per namespace in the desired set.
type NamespaceServerState struct {
	// Namespace is the logical namespace name (e.g. "default", "alpha").
	Namespace string

	// NetworkExists is true when the per-namespace podman bridge
	// (coolify-<ns>-mesh) already exists on this host.
	NetworkExists bool

	// ContainerSubnet is the /<prefix> owned by the per-namespace bridge
	// (read from `podman network inspect`). nil when not yet created.
	ContainerSubnet *net.IPNet

	// DNSEnabled is true when the per-namespace network has `dns_enabled=true`
	// (netavark auto-starts aardvark-dns on the bridge gateway:53). coold owns
	// that socket, so drift triggers ActionRecreatePodmanNet.
	DNSEnabled bool

	// Label is the `io.coolify.namespace` label on the network. Used only as
	// an assertion that the network was created by us — label mismatch is
	// treated like "the network exists but is not ours" and triggers recreate.
	Label string
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

	// Namespaces maps namespace name → per-namespace podman state on this
	// host. Populated by Probe for every namespace in the desired set.
	Namespaces map[string]*NamespaceServerState

	// IPForwardEnabled is true when net.ipv4.ip_forward == 1.
	IPForwardEnabled bool

	// FirewallActive is true when coolify-mesh-fw.service is active.
	FirewallActive bool

	// DefaultDenyActive is true when the COOLIFY-INTRA chain exists and
	// terminates in DROP (the default-deny scaffold is in place).
	DefaultDenyActive bool

	// FirewallUnitSha256 is the sha256 of /etc/systemd/system/coolify-mesh-fw.service
	// (hex), or empty when absent. Used to detect unit drift when the desired
	// set of namespace subnets changes.
	FirewallUnitSha256 string

	// BridgeTableExists is true when `nft list table bridge coolify_bridge`
	// succeeds on this host (nft bridge-family deny scaffold is in place).
	BridgeTableExists bool

	// NftAvailable is true when `nft --version` exits 0 on this host.
	NftAvailable bool

	// CorrosionInstalled is true when /usr/local/bin/corrosion exists and is executable.
	CorrosionInstalled bool

	// CorrosionActive is true when the corrosion systemd service is active.
	CorrosionActive bool

	// CorrosionConfigHash is the sha256 of /etc/corrosion/config.toml, or empty
	// when the file is absent.  Used to detect drift when peer list changes.
	CorrosionConfigHash string

	// CorrosionSchemaExists is true when /etc/corrosion/schemas/coolify.sql exists.
	CorrosionSchemaExists bool

	// CorrosionSchemaSha256 is the sha256 of /etc/corrosion/schemas/coolify.sql
	// (hex), or empty when absent. Used by BuildPlan to detect schema drift so
	// a new schema revision triggers re-write + corrosion restart + DB reset.
	CorrosionSchemaSha256 string

	// CooldInstalled is true when /usr/local/bin/coold exists and is executable.
	CooldInstalled bool

	// CooldActive is true when the coold systemd service is active.
	CooldActive bool

	// CorrosionVersion is the content of /usr/local/bin/corrosion.version
	// (trimmed), or empty when absent. Matches the version tag passed to
	// CorrosionInstallCommand (e.g. "nightly", "v1.2.3").
	CorrosionVersion string

	// CooldVersion is the content of /usr/local/bin/coold.version (trimmed),
	// or empty when absent.
	CooldVersion string

	// CooldUnitSha256 is the sha256 of /etc/systemd/system/coold.service (hex),
	// or empty when absent. Used by BuildPlan to detect generator changes
	// (e.g. Requires→Wants) that would otherwise be invisible.
	CooldUnitSha256 string
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

// AssignedContainerSubnets returns the per-(namespace, host) subnets that are
// already assigned on remote podman networks. The result is nested:
// `out[namespace][host] = subnet`.
func (m *MeshState) AssignedContainerSubnets() map[string]map[string]*net.IPNet {
	out := map[string]map[string]*net.IPNet{}
	for host, s := range m.Servers {
		if s == nil {
			continue
		}
		for ns, nss := range s.Namespaces {
			if nss == nil || nss.ContainerSubnet == nil {
				continue
			}
			if out[ns] == nil {
				out[ns] = map[string]*net.IPNet{}
			}
			out[ns][host] = nss.ContainerSubnet
		}
	}
	return out
}

// FirewallSubnets returns the sorted-by-namespace list of this host's
// container subnets across all namespaces (one /prefix per namespace). Used
// by the firewall service unit generator.
func (s *ServerState) FirewallSubnets() []*net.IPNet {
	var out []*net.IPNet
	names := make([]string, 0, len(s.Namespaces))
	for n := range s.Namespaces {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if ns := s.Namespaces[n]; ns != nil && ns.ContainerSubnet != nil {
			out = append(out, ns.ContainerSubnet)
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

	// ContainerPool is the address pool from which per-(namespace, host)
	// container subnets are carved (default 10.210.0.0/16). One pool is
	// shared across all namespaces so subnets cannot overlap.
	ContainerPool *net.IPNet

	// ContainerPrefix is the prefix length of each per-host, per-namespace
	// container subnet (default 24, giving each host 254 usable container IPs
	// per namespace).
	ContainerPrefix int

	// ListenPort is the WireGuard UDP listen port (default 51820).
	ListenPort int

	// InstallPodman, when true, installs Podman, enables its socket, creates
	// the per-namespace bridge networks, installs firewall rules, and enables
	// IP forwarding on each server.
	InstallPodman bool

	// Namespaces lists every namespace the mesh should carry. Ordered —
	// deterministic iteration produces stable subnet assignments. At least
	// one entry (typically "default") is always expected.
	Namespaces []string

	// DefaultDenyContainers, when true (and InstallPodman is true), installs
	// default-deny iptables rules for ALL container traffic on the host's
	// container subnets (intra-host AND cross-host via wg0). The v5 control
	// plane manages the explicit allow-list in the COOLIFY-ALLOW chain.
	DefaultDenyContainers bool

	// InstallCoold, when true, downloads corrosion + coold from GitHub releases
	// to each host, writes their configs/unit files, and enables both services.
	// Requires InstallPodman (coold depends on podman.socket).
	InstallCoold bool

	// CooldVersion is the release tag to download (e.g. "nightly", "v1.2.3").
	CooldVersion string

	// CorrosionVersion is the release tag to download for corrosion.
	CorrosionVersion string

	// CorrosionGossipPort is the SWIM gossip UDP port (default 8787).
	CorrosionGossipPort int

	// CorrosionAPIPort is the corrosion HTTP API port bound to 127.0.0.1 (default 8080).
	CorrosionAPIPort int

	// CentralHost is the SSH address of the central VM that runs broker
	// and Redis. Empty string disables phases 4+5 (broker setup).
	// Must be an element of Hosts.
	CentralHost string

	// BrokerVersion is the release tag for broker (e.g. "nightly").
	BrokerVersion string

	// EnableBuilder, when true, installs buildah/git and the builder binary
	// on every host, advertises the "builder" capability in the host JWT,
	// and wires coold to accept `BuildRequest` frames from the broker on its
	// existing gRPC stream. Requires a non-empty CentralHost (broker issues
	// the JWT) and InstallPodman (coold's build subprocess shells out to
	// buildah which shares podman's containers-storage). Safe to leave false
	// until build-side rollout.
	EnableBuilder bool

	// BuilderCapacity caps concurrent builds per host. 0 falls back to 2 (the
	// coold builder adapter's own default).
	BuilderCapacity int
}

// SortedNamespaces returns the desired namespaces in deterministic order.
func (d *DesiredMesh) SortedNamespaces() []string {
	out := append([]string(nil), d.Namespaces...)
	sort.Strings(out)
	return out
}

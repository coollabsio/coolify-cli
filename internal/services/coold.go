package services

import (
	"fmt"
	"net"
	"strings"
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

// CooldNamespace describes one namespace for coold's env var. coold's
// embedded DNS binds <BridgeGateway>:53 per namespace, and its sync loop
// iterates `Network` to discover containers.
type CooldNamespace struct {
	Name          string // e.g. "default", "alpha"
	Network       string // e.g. "coolify-default-mesh" — podman bridge name
	BridgeGateway net.IP // the .1 of that namespace's per-host container subnet
}

// CooldNamespacesEnvValue renders the COOLD_NAMESPACES env value. Shape:
//
//	default:coolify-default-mesh:10.210.0.1,alpha:coolify-alpha-mesh:10.220.0.1
//
// Triples are comma-separated; fields within a triple are colon-separated.
// Empty slice yields empty string so callers can omit the env var entirely.
func CooldNamespacesEnvValue(ns []CooldNamespace) string {
	parts := make([]string, 0, len(ns))
	for _, n := range ns {
		parts = append(parts, fmt.Sprintf("%s:%s:%s", n.Name, n.Network, n.BridgeGateway))
	}
	return strings.Join(parts, ",")
}

// CooldServiceUnit returns the systemd unit text for coold.
//
// mgmtIP is this host's wg0 management IP (coold writes rows scoped to it and
// binds its REST API to mgmtIP:CooldAPIPort).
//
// namespaces is the ordered list of namespaces coold manages on this host.
// Each gets its own podman network (coolify-<ns>-mesh) and its own DNS bind
// (bridge gateway :53). Pass nil to skip namespace env injection (e.g. tests
// that don't care about namespaces); coold's config.rs defaults to a single
// `default` entry.
func CooldServiceUnit(mgmtIP net.IP, namespaces []CooldNamespace) string {
	// Wants (not Requires) on corrosion: if corrosion crashes/restarts we want
	// coold to stay up and retry — reconcile_once already backs off for 1s on
	// error, so it self-heals once corrosion is back. Requires would cascade
	// stop coold and leave it down until someone restarted it.
	nsEnv := ""
	if len(namespaces) > 0 {
		nsEnv = fmt.Sprintf(`Environment=COOLD_NAMESPACES=%s
Environment=COOLD_DNS_ZONE=%s
`, CooldNamespacesEnvValue(namespaces), DefaultCooldDNSZone)
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
`, mgmtIP, nsEnv, apiEnv)
}

// CooldInstallCommand returns a shell snippet that downloads and installs coold
// from the GitHub release for the given version tag (e.g. "nightly", "v1.2.3").
// Architecture is auto-detected on the remote host via uname -m.
// The version tag is written to /usr/local/bin/coold.version after install.
func CooldInstallCommand(version string) string {
	return fmt.Sprintf(`set -e
ARCH_RAW=$(uname -m)
case "$ARCH_RAW" in
  x86_64)  ARCH=amd64 ;;
  aarch64) ARCH=arm64 ;;
  *) echo "unsupported arch: $ARCH_RAW" >&2; exit 1 ;;
esac
URL="https://github.com/coollabsio/coold/releases/download/%s/coold-linux-${ARCH}.tar.gz"
DLDIR=$(mktemp -d)
trap 'rm -rf "$DLDIR"' EXIT
curl -fsSL --retry 3 --max-time 120 -o "$DLDIR/coold.tar.gz" "$URL"
tar -xzf "$DLDIR/coold.tar.gz" -C "$DLDIR"
test -f "$DLDIR/coold" || { echo "coold binary not found in tarball" >&2; exit 1; }
install -m 0755 "$DLDIR/coold" /usr/local/bin/coold.tmp
mv /usr/local/bin/coold.tmp /usr/local/bin/coold
echo '%s' > /usr/local/bin/coold.version`, version, version)
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

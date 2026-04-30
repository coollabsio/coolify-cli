package services

import "fmt"

// SchedulerGRPCPort is the TCP port scheduler listens on. coold dials this stream
// and carries both coold and builder traffic on the same connection — there
// is no longer a separate listener for builds.
const SchedulerGRPCPort = 6443

// SchedulerJWTPubPath is the on-host path where the scheduler reads the ES256 public key.
const SchedulerJWTPubPath = "/etc/coolify/jwt.pub"

// SchedulerJWTPrivPath is the on-central path for the EC private key (chmod 0600).
const SchedulerJWTPrivPath = "/etc/coolify/jwt.priv"

// HostJWTPath is the on-host path where coold reads its bearer JWT.
const HostJWTPath = "/etc/coolify/host-jwt"

// SchedulerUnixSocketPath is the on-host path of the scheduler's HTTP-over-UDS
// listener. The central-plane caller (Laravel) connects here. Access
// control is filesystem perms — see SchedulerServiceUnit.
const SchedulerUnixSocketPath = "/run/coolify/scheduler.sock"

// SchedulerServiceUnit returns the systemd unit text for scheduler.
//
// grpcBind is "ip:port" for the single gRPC listener (e.g. "100.64.0.1:6443").
// It binds on the central host's wg0 mgmt IP so the listener is unreachable
// outside the mesh.
//
// RuntimeDirectory=coolify creates /run/coolify owned by the scheduler user
// at unit start, which is where the UDS gets bound. Laravel group access
// is configured at deploy time via SCHEDULER_UNIX_SOCKET_GROUP once the
// PHP-FPM group is finalized; until then the socket stays 0600.
func SchedulerServiceUnit(grpcBind, jwtPubPath string) string {
	return fmt.Sprintf(`[Unit]
Description=Coolify scheduler
After=network-online.target wg-quick@wg0.service

[Service]
RuntimeDirectory=coolify
RuntimeDirectoryMode=0750
Environment=SCHEDULER_GRPC_BIND=%s
Environment=SCHEDULER_UNIX_SOCKET_PATH=%s
Environment=SCHEDULER_JWT_PUBLIC_KEY_PATH=%s
ExecStart=/usr/local/bin/scheduler
Restart=on-failure
RestartSec=2s

[Install]
WantedBy=multi-user.target
`, grpcBind, SchedulerUnixSocketPath, jwtPubPath)
}

// SchedulerInstallCommand returns a shell snippet that downloads and installs
// scheduler from the GitHub release for the given version tag.
func SchedulerInstallCommand(version string) string {
	return fmt.Sprintf(`set -e
ARCH_RAW=$(uname -m)
case "$ARCH_RAW" in
  x86_64)  ARCH=amd64 ;;
  aarch64) ARCH=arm64 ;;
  *) echo "unsupported arch: $ARCH_RAW" >&2; exit 1 ;;
esac
URL="https://github.com/coollabsio/coold/releases/download/%s/scheduler-linux-${ARCH}.tar.gz"
DLDIR=$(mktemp -d)
trap 'rm -rf "$DLDIR"' EXIT
curl -fsSL --retry 3 --max-time 120 -o "$DLDIR/scheduler.tar.gz" "$URL"
tar -xzf "$DLDIR/scheduler.tar.gz" -C "$DLDIR"
test -f "$DLDIR/scheduler" || { echo "scheduler binary not found in tarball" >&2; exit 1; }
install -m 0755 "$DLDIR/scheduler" /usr/local/bin/scheduler.tmp
mv /usr/local/bin/scheduler.tmp /usr/local/bin/scheduler
echo '%s' > /usr/local/bin/scheduler.version`, version, version)
}

// EnsureJWTKeypairCommand returns a shell snippet that generates an EC P-256
// keypair in PKCS8 format on the central host (idempotent).
func EnsureJWTKeypairCommand() string {
	return `mkdir -p /etc/coolify && ` +
		`if [ ! -f ` + SchedulerJWTPrivPath + ` ]; then ` +
		`openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 ` +
		`-out ` + SchedulerJWTPrivPath + `.tmp 2>&1 && ` +
		`chmod 0600 ` + SchedulerJWTPrivPath + `.tmp && ` +
		`mv ` + SchedulerJWTPrivPath + `.tmp ` + SchedulerJWTPrivPath + ` && ` +
		`openssl pkey -in ` + SchedulerJWTPrivPath + ` -pubout -out ` + SchedulerJWTPubPath + ` 2>&1 && ` +
		`chmod 0644 ` + SchedulerJWTPubPath + `; fi`
}

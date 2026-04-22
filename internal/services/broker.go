package services

import "fmt"

// BrokerGRPCPort is the TCP port broker listens on. coold dials this stream
// and carries both coold and builder traffic on the same connection — there
// is no longer a separate listener for builds.
const BrokerGRPCPort = 6443

// BrokerJWTPubPath is the on-host path where the broker reads the ES256 public key.
const BrokerJWTPubPath = "/etc/coolify/jwt.pub"

// BrokerJWTPrivPath is the on-central path for the EC private key (chmod 0600).
const BrokerJWTPrivPath = "/etc/coolify/jwt.priv"

// HostJWTPath is the on-host path where coold reads its bearer JWT.
const HostJWTPath = "/etc/coolify/host-jwt"

// BrokerServiceUnit returns the systemd unit text for broker.
//
// grpcBind is "ip:port" for the single gRPC listener (e.g. "100.64.0.1:6443").
// It binds on the central host's wg0 mgmt IP so the listener is unreachable
// outside the mesh.
func BrokerServiceUnit(grpcBind, redisURL, jwtPubPath string) string {
	return fmt.Sprintf(`[Unit]
Description=Coolify broker
After=network-online.target wg-quick@wg0.service redis-server.service redis.service
Wants=redis-server.service redis.service

[Service]
Environment=BROKER_GRPC_BIND=%s
Environment=BROKER_REDIS_URL=%s
Environment=BROKER_JWT_PUBLIC_KEY_PATH=%s
ExecStart=/usr/local/bin/broker
Restart=on-failure
RestartSec=2s

[Install]
WantedBy=multi-user.target
`, grpcBind, redisURL, jwtPubPath)
}

// BrokerInstallCommand returns a shell snippet that downloads and installs
// broker from the GitHub release for the given version tag.
func BrokerInstallCommand(version string) string {
	return fmt.Sprintf(`set -e
ARCH_RAW=$(uname -m)
case "$ARCH_RAW" in
  x86_64)  ARCH=amd64 ;;
  aarch64) ARCH=arm64 ;;
  *) echo "unsupported arch: $ARCH_RAW" >&2; exit 1 ;;
esac
URL="https://github.com/coollabsio/coold/releases/download/%s/broker-linux-${ARCH}.tar.gz"
DLDIR=$(mktemp -d)
trap 'rm -rf "$DLDIR"' EXIT
curl -fsSL --retry 3 --max-time 120 -o "$DLDIR/broker.tar.gz" "$URL"
tar -xzf "$DLDIR/broker.tar.gz" -C "$DLDIR"
test -f "$DLDIR/broker" || { echo "broker binary not found in tarball" >&2; exit 1; }
install -m 0755 "$DLDIR/broker" /usr/local/bin/broker.tmp
mv /usr/local/bin/broker.tmp /usr/local/bin/broker
echo '%s' > /usr/local/bin/broker.version`, version, version)
}

// EnsureJWTKeypairCommand returns a shell snippet that generates an EC P-256
// keypair in PKCS8 format on the central host (idempotent).
func EnsureJWTKeypairCommand() string {
	return `mkdir -p /etc/coolify && ` +
		`if [ ! -f ` + BrokerJWTPrivPath + ` ]; then ` +
		`openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:P-256 ` +
		`-out ` + BrokerJWTPrivPath + `.tmp 2>&1 && ` +
		`chmod 0600 ` + BrokerJWTPrivPath + `.tmp && ` +
		`mv ` + BrokerJWTPrivPath + `.tmp ` + BrokerJWTPrivPath + ` && ` +
		`openssl pkey -in ` + BrokerJWTPrivPath + ` -pubout -out ` + BrokerJWTPubPath + ` 2>&1 && ` +
		`chmod 0644 ` + BrokerJWTPubPath + `; fi`
}

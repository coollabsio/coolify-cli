package services

import "fmt"

// BuilderWorkDir is the scratch root coold creates per-build subdirectories
// in when it dispatches a `BuildRequest`. Cleaned per-request by coold.
const BuilderWorkDir = "/var/lib/coolify-builder/work"

// BuilderBinaryPath is the path to the builder binary coold spawns as a
// short-lived subprocess under a `systemd-run --scope` transient unit. No
// long-running builder daemon exists on the host.
const BuilderBinaryPath = "/usr/local/bin/builder"

// BuilderInstallCommand returns a shell snippet that installs buildah + git
// (required by the builder pipeline), ensures the work directory exists,
// and downloads the builder binary from the GitHub release for the given
// version tag. The version tag should track the coold release — builder
// and coold ship from the same workspace.
func BuilderInstallCommand(version string) string {
	return fmt.Sprintf(`set -e
DEBIAN_FRONTEND=noninteractive apt-get update -qq 2>/dev/null
DEBIAN_FRONTEND=noninteractive apt-get install -y \
  -o Dpkg::Options::="--force-confold" \
  buildah git ca-certificates 2>&1 >/dev/null
mkdir -p %[1]s
ARCH_RAW=$(uname -m)
case "$ARCH_RAW" in
  x86_64)  ARCH=amd64 ;;
  aarch64) ARCH=arm64 ;;
  *) echo "unsupported arch: $ARCH_RAW" >&2; exit 1 ;;
esac
URL="https://github.com/coollabsio/coold/releases/download/%[2]s/builder-linux-${ARCH}.tar.gz"
DLDIR=$(mktemp -d)
trap 'rm -rf "$DLDIR"' EXIT
curl -fsSL --retry 3 --max-time 120 -o "$DLDIR/builder.tar.gz" "$URL"
tar -xzf "$DLDIR/builder.tar.gz" -C "$DLDIR"
test -f "$DLDIR/builder" || { echo "builder binary not found in tarball" >&2; exit 1; }
install -m 0755 "$DLDIR/builder" %[3]s.tmp
mv %[3]s.tmp %[3]s
echo '%[2]s' > %[3]s.version`,
		BuilderWorkDir, version, BuilderBinaryPath)
}

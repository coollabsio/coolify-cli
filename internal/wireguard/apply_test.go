package wireguard

import (
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstLine(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"single line", "single line"},
		{"first\nsecond\nthird", "first"},
		{"  spaces  \nnext", "spaces  "},
		{"\nleading newline", "leading newline"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, firstLine(tt.input), "input: %q", tt.input)
	}
}

func TestPodmanNetCreateCmd_DisablesDNSAndLabels(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("10.210.0.0/24")
	gw := net.ParseIP("10.210.0.1")
	got := podmanNetCreateCmd("coolify-default-mesh", "default", subnet, gw)

	// Must pass --disable-dns so aardvark-dns never binds bridge gateway :53
	// (coold owns that socket).
	assert.Contains(t, got, "--disable-dns", "create must include --disable-dns")
	assert.Contains(t, got, "--subnet=10.210.0.0/24")
	assert.Contains(t, got, "--gateway=10.210.0.1")
	// Labels identify the network as ours + carry the namespace for drift checks.
	assert.Contains(t, got, "--label io.coolify.managed=true")
	assert.Contains(t, got, "--label io.coolify.namespace=default")
	// Idempotency guard must still be present.
	assert.Contains(t, got, "podman network exists coolify-default-mesh")
}

func TestPodmanNetRecreateCmd_DropsAndCreatesWithDisableDNS(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("10.220.0.0/24")
	gw := net.ParseIP("10.220.0.1")
	got := podmanNetRecreateCmd("coolify-alpha-mesh", "alpha", subnet, gw)

	assert.Contains(t, got, "podman network rm -f coolify-alpha-mesh")
	assert.Contains(t, got, "--disable-dns")
	assert.Contains(t, got, "--subnet=10.220.0.0/24")
	assert.Contains(t, got, "--label io.coolify.namespace=alpha")
	// rm must come before create so the ordering is unambiguous.
	rmIdx := strings.Index(got, "rm -f")
	createIdx := strings.Index(got, "network create")
	assert.True(t, rmIdx >= 0 && createIdx > rmIdx, "rm must precede create")
}

func TestHeredocWrite_EmitsChmodBeforeMv(t *testing.T) {
	got := heredocWrite("/etc/corrosion/config.toml", "body", "TAG", 0o600)

	assert.Contains(t, got, "cat > /etc/corrosion/config.toml.tmp <<'TAG'")
	assert.Contains(t, got, "\nbody")
	assert.Contains(t, got, "chmod 600 /etc/corrosion/config.toml.tmp")
	assert.Contains(t, got, "mv /etc/corrosion/config.toml.tmp /etc/corrosion/config.toml")

	chmodIdx := strings.Index(got, "chmod 600")
	mvIdx := strings.Index(got, "mv /etc/corrosion")
	assert.True(t, chmodIdx > 0 && mvIdx > chmodIdx,
		"chmod must precede mv so final rename is atomic with intended mode")
}

func TestHeredocWrite_DifferentModes(t *testing.T) {
	unit := heredocWrite("/etc/systemd/system/x.service", "b", "T", 0o644)
	assert.Contains(t, unit, "chmod 644 /etc/systemd/system/x.service.tmp")

	secret := heredocWrite("/etc/corrosion/schemas/coolify.sql", "b", "T", 0o600)
	assert.Contains(t, secret, "chmod 600 /etc/corrosion/schemas/coolify.sql.tmp")
}

func TestNonEmptyLines(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"line1\nline2", []string{"line1", "line2"}},
		{"line1\n\nline2", []string{"line1", "line2"}},
		{"  \n  \nactual", []string{"actual"}},
		{"only", []string{"only"}},
	}
	for _, tt := range tests {
		got := nonEmptyLines(tt.input)
		assert.Equal(t, tt.want, got, "input: %q", tt.input)
	}
}

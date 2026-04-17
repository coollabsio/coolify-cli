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

func TestPodmanNetCreateCmd_DisablesDNS(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("10.210.0.0/24")
	gw := net.ParseIP("10.210.0.1")
	got := podmanNetCreateCmd("coolify-mesh", subnet, gw)

	// Must pass --disable-dns so aardvark-dns never binds bridge gateway :53
	// (coold owns that socket).
	assert.Contains(t, got, "--disable-dns", "create must include --disable-dns")
	assert.Contains(t, got, "--subnet=10.210.0.0/24")
	assert.Contains(t, got, "--gateway=10.210.0.1")
	// Idempotency guard must still be present.
	assert.Contains(t, got, "podman network exists coolify-mesh")
}

func TestPodmanNetRecreateCmd_DropsAndCreatesWithDisableDNS(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("10.210.0.0/24")
	gw := net.ParseIP("10.210.0.1")
	got := podmanNetRecreateCmd("coolify-mesh", subnet, gw)

	assert.Contains(t, got, "podman network rm -f coolify-mesh")
	assert.Contains(t, got, "--disable-dns")
	assert.Contains(t, got, "--subnet=10.210.0.0/24")
	// rm must come before create so the ordering is unambiguous.
	rmIdx := strings.Index(got, "rm -f")
	createIdx := strings.Index(got, "network create")
	assert.True(t, rmIdx >= 0 && createIdx > rmIdx, "rm must precede create")
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

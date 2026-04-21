package firewall

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeID_Stable(t *testing.T) {
	a := ComputeID("default", net.ParseIP("10.210.0.10"), net.ParseIP("10.210.1.10"), "tcp", 80)
	b := ComputeID("default", net.ParseIP("10.210.0.10"), net.ParseIP("10.210.1.10"), "tcp", 80)
	assert.Equal(t, a, b)
	assert.Len(t, a, 12)
}

func TestComputeID_CaseInsensitiveProto(t *testing.T) {
	a := ComputeID("default", net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "TCP", 80)
	b := ComputeID("default", net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "tcp", 80)
	assert.Equal(t, a, b)
}

func TestComputeID_DifferentInputsDifferent(t *testing.T) {
	a := ComputeID("default", net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "tcp", 80)
	b := ComputeID("default", net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "tcp", 443)
	assert.NotEqual(t, a, b)
}

// TestComputeID_DifferentNamespacesDifferent verifies that identical
// src/dst/proto/port tuples in different namespaces produce different IDs —
// this is the whole point of per-namespace rule identity.
func TestComputeID_DifferentNamespacesDifferent(t *testing.T) {
	a := ComputeID("default", net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.2"), "tcp", 80)
	b := ComputeID("alpha", net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.2"), "tcp", 80)
	assert.NotEqual(t, a, b)
}

// TestComputeID_EmptyNamespaceMatchesDefault guards the wire-compat rule:
// an empty namespace must hash the same as "default" so older coold builds
// and newer CLI callers agree on the same ID.
func TestComputeID_EmptyNamespaceMatchesDefault(t *testing.T) {
	empty := ComputeID("", net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.2"), "tcp", 80)
	def := ComputeID("default", net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.2"), "tcp", 80)
	assert.Equal(t, empty, def)
}

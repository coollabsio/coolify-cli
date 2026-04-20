package firewall

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeID_Stable(t *testing.T) {
	a := ComputeID(net.ParseIP("10.210.0.10"), net.ParseIP("10.210.1.10"), "tcp", 80)
	b := ComputeID(net.ParseIP("10.210.0.10"), net.ParseIP("10.210.1.10"), "tcp", 80)
	assert.Equal(t, a, b)
	assert.Len(t, a, 12)
}

func TestComputeID_CaseInsensitiveProto(t *testing.T) {
	a := ComputeID(net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "TCP", 80)
	b := ComputeID(net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "tcp", 80)
	assert.Equal(t, a, b)
}

func TestComputeID_DifferentInputsDifferent(t *testing.T) {
	a := ComputeID(net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "tcp", 80)
	b := ComputeID(net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "tcp", 443)
	assert.NotEqual(t, a, b)
}

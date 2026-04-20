package firewall

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	ifw "github.com/coollabsio/coolify-cli/internal/firewall"
)

func cs() []ifw.Container {
	return []ifw.Container{
		{Host: "h1", ID: "aaa111111111", Name: "web", IP: net.ParseIP("10.210.0.10")},
		{Host: "h2", ID: "bbb222222222", Name: "api", IP: net.ParseIP("10.210.1.10")},
		{Host: "h3", ID: "ccc333333333", Name: "web", IP: net.ParseIP("10.210.2.10")},
	}
}

func TestResolveEndpoint_ByName_Unique(t *testing.T) {
	c, err := resolveEndpoint("api", cs())
	assert.NoError(t, err)
	assert.Equal(t, "h2", c.Host)
	assert.Equal(t, "10.210.1.10", c.IP.String())
}

func TestResolveEndpoint_ByName_Ambiguous(t *testing.T) {
	_, err := resolveEndpoint("web", cs())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
}

func TestResolveEndpoint_ByShortID(t *testing.T) {
	c, err := resolveEndpoint("bbb", cs())
	assert.NoError(t, err)
	assert.Equal(t, "h2", c.Host)
}

func TestResolveEndpoint_ByHostName(t *testing.T) {
	c, err := resolveEndpoint("h3:web", cs())
	assert.NoError(t, err)
	assert.Equal(t, "h3", c.Host)
	assert.Equal(t, "10.210.2.10", c.IP.String())
}

func TestResolveEndpoint_ByRawIP(t *testing.T) {
	c, err := resolveEndpoint("10.210.1.10", cs())
	assert.NoError(t, err)
	assert.Equal(t, "h2", c.Host)
}

func TestResolveEndpoint_UnknownRawIP_Synthetic(t *testing.T) {
	c, err := resolveEndpoint("10.99.99.99", cs())
	assert.NoError(t, err)
	assert.Equal(t, "", c.Host)
	assert.Equal(t, "10.99.99.99", c.IP.String())
}

func TestResolveEndpoint_NotFound(t *testing.T) {
	_, err := resolveEndpoint("nobody", cs())
	assert.Error(t, err)
}

func TestResolveEndpoint_Empty(t *testing.T) {
	_, err := resolveEndpoint("", cs())
	assert.Error(t, err)
}

func TestFindHostForIP(t *testing.T) {
	h, ok := findHostForIP(net.ParseIP("10.210.0.10"), cs())
	assert.True(t, ok)
	assert.Equal(t, "h1", h)
	_, ok = findHostForIP(net.ParseIP("1.2.3.4"), cs())
	assert.False(t, ok)
}

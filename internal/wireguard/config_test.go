package wireguard

import (
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderConfig_NoPeers(t *testing.T) {
	mgmtIP := net.ParseIP("100.64.0.1").To4()
	got := RenderConfig(mgmtIP, 51820, nil)

	assert.Contains(t, got, "[Interface]")
	assert.Contains(t, got, "Address = 100.64.0.1/32")
	assert.Contains(t, got, "ListenPort = 51820")
	assert.Contains(t, got, "PrivateKey = __PRIVKEY__")
	assert.NotContains(t, got, "[Peer]")
}

func TestRenderConfig_WithPeers(t *testing.T) {
	mgmtIP := net.ParseIP("100.64.0.1").To4()
	peers := []PeerConfig{
		{
			Endpoint:        "203.0.113.11",
			PublicKey:       "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
			MgmtIP:          net.ParseIP("100.64.0.1").To4(),
			ContainerSubnet: mustParseCIDR("10.210.1.0/24"),
		},
		{
			Endpoint:        "203.0.113.12",
			PublicKey:       "CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=",
			MgmtIP:          net.ParseIP("100.64.0.2").To4(),
			ContainerSubnet: mustParseCIDR("10.210.2.0/24"),
		},
	}

	got := RenderConfig(mgmtIP, 51820, peers)

	assert.Equal(t, 2, strings.Count(got, "[Peer]"))
	assert.Contains(t, got, "PublicKey = BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=")
	assert.Contains(t, got, "Endpoint = 203.0.113.11:51820")
	assert.Contains(t, got, "AllowedIPs = 100.64.0.1/32, 10.210.1.0/24")
	assert.Contains(t, got, "PersistentKeepalive = 25")
	assert.Contains(t, got, "PublicKey = CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=")
	assert.Contains(t, got, "AllowedIPs = 100.64.0.2/32, 10.210.2.0/24")
}

func TestWriteConfigCommand_ContainsPrivkeyRead(t *testing.T) {
	mgmtIP := net.ParseIP("100.64.0.1").To4()
	peers := []PeerConfig{
		{
			Endpoint:        "203.0.113.11",
			PublicKey:       "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
			MgmtIP:          net.ParseIP("100.64.0.1").To4(),
			ContainerSubnet: mustParseCIDR("10.210.1.0/24"),
		},
	}

	cmd := WriteConfigCommand("wg0", mgmtIP, 51820, peers)

	assert.Contains(t, cmd, "cat /etc/wireguard/privatekey")
	assert.Contains(t, cmd, "$PRIVKEY")
	assert.Contains(t, cmd, ".conf.tmp")
	assert.Contains(t, cmd, "mv /etc/wireguard/wg0.conf.tmp /etc/wireguard/wg0.conf")
	// Host Address is the mgmt /32 — outside the container pool.
	assert.Contains(t, cmd, "Address = 100.64.0.1/32")
	// Peer AllowedIPs lists peer mgmt /32 + peer container /24.
	assert.Contains(t, cmd, "100.64.0.1/32, 10.210.1.0/24")
}

func TestWriteConfigCommand_NoPeers(t *testing.T) {
	mgmtIP := net.ParseIP("100.64.0.1").To4()
	cmd := WriteConfigCommand("wg0", mgmtIP, 51820, nil)

	assert.Contains(t, cmd, "PRIVKEY")
	assert.Contains(t, cmd, "51820")
	assert.NotContains(t, cmd, "[Peer]")
}

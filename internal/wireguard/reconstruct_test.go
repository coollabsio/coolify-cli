package wireguard

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeReconRunner is a deterministic ssh.Runner for reconstruct unit tests.
type fakeReconRunner struct {
	responses map[string]string
}

func (f *fakeReconRunner) Run(_ context.Context, _, _ string, _ int, cmd string) (string, string, error) {
	for substr, resp := range f.responses {
		if strings.Contains(cmd, substr) {
			return resp, "", nil
		}
	}
	return "", "", nil
}

func readFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "test", "fixtures", "wg", name)
	b, err := os.ReadFile(path)
	require.NoError(t, err, "missing fixture %s", name)
	return string(b)
}

func TestParseConfigFile_Full(t *testing.T) {
	content := readFixture(t, "wg0.conf")
	state := &ServerState{}
	parseConfigFile(state, content)

	require.NotNil(t, state.WireGuardMgmtIP)
	assert.Equal(t, "100.64.0.1", state.WireGuardMgmtIP.String())
	assert.Equal(t, 51820, state.ListenPort)
	require.Len(t, state.Peers, 1)
	assert.Equal(t, "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBK=", state.Peers[0].PublicKey)
	assert.Equal(t, "203.0.113.11:51820", state.Peers[0].Endpoint)
	assert.Equal(t, 25, state.Peers[0].PersistentKeepalive)
}

func TestParseConfigFile_Empty(t *testing.T) {
	state := &ServerState{}
	parseConfigFile(state, "")
	assert.Nil(t, state.WireGuardMgmtIP)
	assert.Empty(t, state.Peers)
}

func TestParseConfigFile_MultiplePeers(t *testing.T) {
	content := `[Interface]
Address = 100.64.0.1/32
ListenPort = 51820
PrivateKey = aaa

[Peer]
PublicKey = BBB=
AllowedIPs = 100.64.0.2/32, 10.210.1.0/24
Endpoint = 1.2.3.4:51820
PersistentKeepalive = 25

[Peer]
PublicKey = CCC=
AllowedIPs = 100.64.0.2/32, 10.210.2.0/24
Endpoint = 1.2.3.5:51820
PersistentKeepalive = 25
`
	state := &ServerState{}
	parseConfigFile(state, content)

	require.Len(t, state.Peers, 2)
	assert.Equal(t, "BBB=", state.Peers[0].PublicKey)
	assert.Equal(t, "CCC=", state.Peers[1].PublicKey)
}

func TestParseConfigFile_IgnoresComments(t *testing.T) {
	content := `# This is a comment
[Interface]
# Another comment
Address = 100.64.0.5/32
ListenPort = 51820
PrivateKey = xxx
`
	state := &ServerState{}
	parseConfigFile(state, content)

	require.NotNil(t, state.WireGuardMgmtIP)
	assert.Equal(t, "100.64.0.5", state.WireGuardMgmtIP.String())
	assert.Empty(t, state.Peers)
}

func TestParseConfigFile_CaseInsensitiveKeys(t *testing.T) {
	content := `[interface]
address = 100.64.0.10/32
listenport = 12345
privatekey = xxx
`
	state := &ServerState{}
	parseConfigFile(state, content)

	require.NotNil(t, state.WireGuardMgmtIP)
	assert.Equal(t, "100.64.0.10", state.WireGuardMgmtIP.String())
	assert.Equal(t, 12345, state.ListenPort)
}

func TestMeshState_AssignedMgmtIPs(t *testing.T) {
	mesh := MeshState{
		Servers: map[string]*ServerState{
			"a": {Host: "a", WireGuardMgmtIP: []byte{100, 64, 0, 1}},
			"b": {Host: "b", WireGuardMgmtIP: nil},
			"c": {Host: "c", WireGuardMgmtIP: []byte{100, 64, 0, 3}},
		},
	}
	ips := mesh.AssignedMgmtIPs()
	assert.Len(t, ips, 2)
	assert.Contains(t, ips, "a")
	assert.NotContains(t, ips, "b")
	assert.Contains(t, ips, "c")
}

func TestMeshState_AssignedContainerSubnets(t *testing.T) {
	mesh := MeshState{
		Servers: map[string]*ServerState{
			"a": {Host: "a", Namespaces: map[string]*NamespaceServerState{
				DefaultNamespace: {Namespace: DefaultNamespace, ContainerSubnet: mustParseCIDR("10.210.0.0/24")},
				"alpha":          {Namespace: "alpha", ContainerSubnet: mustParseCIDR("10.220.0.0/24")},
			}},
			"b": {Host: "b", Namespaces: map[string]*NamespaceServerState{
				DefaultNamespace: {Namespace: DefaultNamespace}, // ContainerSubnet nil
			}},
			"c": {Host: "c", Namespaces: map[string]*NamespaceServerState{
				DefaultNamespace: {Namespace: DefaultNamespace, ContainerSubnet: mustParseCIDR("10.210.2.0/24")},
			}},
		},
	}
	subs := mesh.AssignedContainerSubnets()
	// Nested: namespace → host → subnet.
	assert.Contains(t, subs[DefaultNamespace], "a")
	assert.NotContains(t, subs[DefaultNamespace], "b")
	assert.Contains(t, subs[DefaultNamespace], "c")
	assert.Contains(t, subs["alpha"], "a")
}

func TestTruncateKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"short", "short"},
		{"12345678", "12345678"},
		{"123456789", "12345678..."},
		{"AAAAAAAABBBBBBBB", "AAAAAAAA..."},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, truncateKey(tt.input), "input: %q", tt.input)
	}
}

func TestProbe_NftAvailableAndBridgeTableExists_True(t *testing.T) {
	runner := &fakeReconRunner{
		responses: map[string]string{
			"dpkg-query":                          "1\n",
			"wg show":                             "",
			"cat /etc/wireguard/":                 "",
			"wg pubkey":                           "",
			"ip -4 -o addr show":                  "",
			"systemctl is-active wg-quick":        "active\n",
			"podman --version":                    "podman version 4.9.0\n",
			"systemctl is-active podman.socket":   "active\n",
			"sysctl net.ipv4.ip_forward":          "net.ipv4.ip_forward = 1\n",
			"podman network inspect":              `[{"name":"coolify-default-mesh","subnets":[{"subnet":"10.210.0.0/24","gateway":"10.210.0.1"}],"dns_enabled":false,"labels":{"io.coolify.managed":"true","io.coolify.namespace":"default"}}]` + "\n",
			"systemctl is-active coolify-mesh-fw": "active\n",
			"sha256sum /etc/systemd/system/coolify-mesh-fw.service": "",
			"iptables -nL COOLIFY-INTRA":                            "yes\n",
			"command -v nft":                                        "yes\n",
			"nft list table bridge coolify_bridge":                  "yes\n",
			"test -x /usr/local/bin/corrosion":                      "yes\n",
			"systemctl is-active corrosion":                         "active\n",
			"sha256sum /etc/corrosion/config.toml":                  "",
			"test -x /usr/local/bin/coold":                          "yes\n",
			"systemctl is-active coold":                             "active\n",
			"cat /etc/coolify/coold-version":                        "",
			"cat /etc/coolify/corrosion-version":                    "",
		},
	}

	state, err := Probe(context.Background(), runner, "1.1.1.1", "root", 22, "wg0", []string{"default"})
	require.NoError(t, err)
	assert.True(t, state.NftAvailable, "NftAvailable should be true")
	assert.True(t, state.BridgeTableExists, "BridgeTableExists should be true")
	// DefaultDenyActive = COOLIFY-INTRA DROP && BridgeTableExists
	assert.True(t, state.DefaultDenyActive, "DefaultDenyActive should be true when both conditions met")
}

func TestProbe_NftNotAvailable_BridgeTableAbsent(t *testing.T) {
	runner := &fakeReconRunner{
		responses: map[string]string{
			"dpkg-query":                           "1\n",
			"iptables -nL COOLIFY-INTRA":           "yes\n",
			"command -v nft":                       "no\n",
			"nft list table bridge coolify_bridge": "no\n",
		},
	}

	state, err := Probe(context.Background(), runner, "1.1.1.1", "root", 22, "wg0", []string{"default"})
	require.NoError(t, err)
	assert.False(t, state.NftAvailable, "NftAvailable should be false")
	assert.False(t, state.BridgeTableExists, "BridgeTableExists should be false")
	// DefaultDenyActive must be false even though COOLIFY-INTRA has DROP
	assert.False(t, state.DefaultDenyActive, "DefaultDenyActive should be false when BridgeTableExists is false")
}

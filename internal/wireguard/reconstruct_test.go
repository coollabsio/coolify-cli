package wireguard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			"a": {Host: "a", ContainerSubnet: mustParseCIDR("10.210.0.0/24")},
			"b": {Host: "b", ContainerSubnet: nil},
			"c": {Host: "c", ContainerSubnet: mustParseCIDR("10.210.2.0/24")},
		},
	}
	subs := mesh.AssignedContainerSubnets()
	assert.Len(t, subs, 2)
	assert.Contains(t, subs, "a")
	assert.NotContains(t, subs, "b")
	assert.Contains(t, subs, "c")
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

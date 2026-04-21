package firewall

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/coollabsio/coolify-cli/cmd/common"
)

func TestEmitContainers_RunsAndFormatsTable(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{
		"podman ps": "aaa111111111|web|10.210.0.10",
	}}
	parent := &FirewallFlags{
		SSHMeshFlags: common.SSHMeshFlags{
			Servers: []string{"h1"}, SSHUser: "root", SSHPort: 22, Concurrency: 1,
		},
		Namespace: common.DefaultNamespace,
	}
	inner := &cobra.Command{Use: "containers"}
	rootCmdFor(inner)

	err := emitContainers(context.Background(), inner, parent, fr)
	assert.NoError(t, err)
	// Discovery command was issued, targeting the default-namespace bridge.
	assert.Len(t, fr.calls, 1)
	assert.Contains(t, fr.calls[0], "podman ps")
	assert.Contains(t, fr.calls[0], "coolify-default-mesh")
}

func TestEmitContainers_EmptyOutput(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{}}
	parent := &FirewallFlags{
		SSHMeshFlags: common.SSHMeshFlags{
			Servers: []string{"h1"}, SSHUser: "root", SSHPort: 22, Concurrency: 1,
		},
		Namespace: common.DefaultNamespace,
	}
	inner := &cobra.Command{Use: "containers"}
	rootCmdFor(inner)

	err := emitContainers(context.Background(), inner, parent, fr)
	assert.NoError(t, err)
}

// TestEmitContainers_AllNamespaces_FansOutAcrossNetworks verifies that with
// --all-namespaces the CLI first enumerates managed networks on every host
// and then issues one podman-ps per namespace it found.
func TestEmitContainers_AllNamespaces_FansOutAcrossNetworks(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{
		// Host reports two managed namespaces via label inspection.
		"podman network ls": "default\nalpha\n",
		// Every subsequent podman-ps returns the same container.
		"podman ps": "aaa111111111|web|10.210.0.10",
	}}
	parent := &FirewallFlags{
		SSHMeshFlags: common.SSHMeshFlags{
			Servers: []string{"h1"}, SSHUser: "root", SSHPort: 22, Concurrency: 1,
		},
		Namespace:     common.DefaultNamespace,
		AllNamespaces: true,
	}
	inner := &cobra.Command{Use: "containers"}
	rootCmdFor(inner)

	err := emitContainers(context.Background(), inner, parent, fr)
	assert.NoError(t, err)

	// Expect one `podman network ls` discovery call + one `podman ps` per
	// discovered namespace (default + alpha = 2).
	var ls, ps int
	for _, c := range fr.calls {
		switch {
		case containsAll(c, "podman network ls", "io.coolify.managed=true"):
			ls++
		case containsAll(c, "podman ps"):
			ps++
		}
	}
	assert.Equal(t, 1, ls, "one namespace-discovery call per host")
	assert.Equal(t, 2, ps, "one container-discovery call per namespace per host")
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	// Tiny local wrapper so tests stay readable without importing strings
	// twice — the test file already uses it elsewhere via cmdFakeRunner.
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

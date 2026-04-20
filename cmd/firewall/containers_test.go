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
		PodmanNetworkName: "coolify-mesh",
	}
	inner := &cobra.Command{Use: "containers"}
	rootCmdFor(inner)

	err := emitContainers(context.Background(), inner, parent, fr)
	assert.NoError(t, err)
	// Discovery command was issued.
	assert.Len(t, fr.calls, 1)
	assert.Contains(t, fr.calls[0], "podman ps")
}

func TestEmitContainers_EmptyOutput(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{}}
	parent := &FirewallFlags{
		SSHMeshFlags: common.SSHMeshFlags{
			Servers: []string{"h1"}, SSHUser: "root", SSHPort: 22, Concurrency: 1,
		},
		PodmanNetworkName: "coolify-mesh",
	}
	inner := &cobra.Command{Use: "containers"}
	rootCmdFor(inner)

	err := emitContainers(context.Background(), inner, parent, fr)
	assert.NoError(t, err)
}

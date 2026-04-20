package firewall

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/coollabsio/coolify-cli/cmd/common"
)

func TestEmitList_RunsAndFormatsTable(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{
		"iptables -S COOLIFY-ALLOW": `-A COOLIFY-ALLOW -s 10.0.0.1 -d 10.0.0.2 -p tcp -m tcp --dport 80 -j ACCEPT
`,
	}}
	parent := &FirewallFlags{
		SSHMeshFlags: common.SSHMeshFlags{
			Servers: []string{"h1"}, SSHUser: "root", SSHPort: 22, Concurrency: 1,
		},
	}
	inner := &cobra.Command{Use: "list"}
	rootCmdFor(inner)

	err := emitList(context.Background(), inner, parent, fr)
	assert.NoError(t, err)
	assert.Len(t, fr.calls, 1)
	assert.Contains(t, fr.calls[0], "iptables -S COOLIFY-ALLOW")
}

func TestEmitList_EmptyChain(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{}}
	parent := &FirewallFlags{
		SSHMeshFlags: common.SSHMeshFlags{
			Servers: []string{"h1"}, SSHUser: "root", SSHPort: 22, Concurrency: 1,
		},
	}
	inner := &cobra.Command{Use: "list"}
	rootCmdFor(inner)

	err := emitList(context.Background(), inner, parent, fr)
	assert.NoError(t, err)
}

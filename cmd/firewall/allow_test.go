package firewall

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/coollabsio/coolify-cli/cmd/common"
	"github.com/coollabsio/coolify-cli/internal/ssh"
)

func TestValidateAllowRevokeFlags(t *testing.T) {
	t.Run("missing from", func(t *testing.T) {
		err := validateAllowRevokeFlags(&allowRevokeFlags{To: "x", Port: 80, Proto: "tcp"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--from")
	})
	t.Run("missing to", func(t *testing.T) {
		err := validateAllowRevokeFlags(&allowRevokeFlags{From: "x", Port: 80, Proto: "tcp"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--to")
	})
	t.Run("missing port with proto", func(t *testing.T) {
		err := validateAllowRevokeFlags(&allowRevokeFlags{From: "a", To: "b", Proto: "tcp"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--port")
	})
	t.Run("bad proto", func(t *testing.T) {
		err := validateAllowRevokeFlags(&allowRevokeFlags{From: "a", To: "b", Proto: "icmp", Port: 1})
		assert.Error(t, err)
	})
	t.Run("ok tcp", func(t *testing.T) {
		err := validateAllowRevokeFlags(&allowRevokeFlags{From: "a", To: "b", Proto: "tcp", Port: 80})
		assert.NoError(t, err)
	})
	t.Run("ok no-proto no-port", func(t *testing.T) {
		err := validateAllowRevokeFlags(&allowRevokeFlags{From: "a", To: "b", Proto: "", Port: 0})
		assert.NoError(t, err)
	})
}

// fakeRunner mirrors internal/firewall's fake but lives in this package.
type cmdFakeRunner struct {
	responses map[string]string
	calls     []string
}

func (f *cmdFakeRunner) Run(_ context.Context, _, _ string, _ int, cmd string) (string, string, error) {
	f.calls = append(f.calls, cmd)
	for sub, resp := range f.responses {
		if strings.Contains(cmd, sub) {
			return resp, "", nil
		}
	}
	return "", "", nil
}

var _ ssh.Runner = (*cmdFakeRunner)(nil)

func rootCmdFor(cmd *cobra.Command) *cobra.Command {
	root := &cobra.Command{Use: "coolify"}
	root.PersistentFlags().String("format", "table", "")
	root.AddCommand(cmd)
	return root
}

func TestEmitAllowRevoke_AppliesOneDirection(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{
		"podman ps": "aaa111111111|web|10.210.0.10",
	}}
	// Single host — fake can't route output by host, so collapse.
	parent := &FirewallFlags{
		SSHMeshFlags: common.SSHMeshFlags{
			Servers: []string{"h1"}, SSHUser: "root", SSHPort: 22, Concurrency: 1,
		},
		PodmanNetworkName: "coolify-mesh",
	}
	local := &allowRevokeFlags{
		From: "10.210.1.5", To: "web", Proto: "tcp", Port: 80,
	}
	inner := &cobra.Command{Use: "allow"}
	rootCmdFor(inner)

	err := emitAllowRevoke(context.Background(), inner, parent, local, fr, false)
	assert.NoError(t, err)

	// One apply call issued (excluding discovery).
	var applies []string
	for _, c := range fr.calls {
		if strings.Contains(c, "iptables -A COOLIFY-ALLOW") {
			applies = append(applies, c)
		}
	}
	assert.Len(t, applies, 1)
	assert.Contains(t, applies[0], "-s 10.210.1.5")
	assert.Contains(t, applies[0], "-d 10.210.0.10")
	assert.Contains(t, applies[0], "--dport 80")
}

func TestEmitAllowRevoke_Bidirectional(t *testing.T) {
	// Single host: bidir still issues 2 rules (one per direction), both with
	// Host = the only host. This is a degenerate but valid setup for the test.
	fr := &cmdFakeRunner{responses: map[string]string{
		"podman ps": "aaa111111111|web|10.210.0.10\nbbb222222222|client|10.210.1.5",
	}}
	parent := &FirewallFlags{
		SSHMeshFlags: common.SSHMeshFlags{
			Servers: []string{"h1"}, SSHUser: "root", SSHPort: 22, Concurrency: 1,
		},
		PodmanNetworkName: "coolify-mesh",
	}
	local := &allowRevokeFlags{
		From: "10.210.1.5", To: "10.210.0.10", Proto: "tcp", Port: 80, Bidirectional: true,
	}
	inner := &cobra.Command{Use: "allow"}
	rootCmdFor(inner)

	err := emitAllowRevoke(context.Background(), inner, parent, local, fr, false)
	assert.NoError(t, err)

	var applies int
	for _, c := range fr.calls {
		if strings.Contains(c, "iptables -A COOLIFY-ALLOW") {
			applies++
		}
	}
	assert.Equal(t, 2, applies)
}

func TestEmitAllowRevoke_RevokeIssuesDelete(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{
		"podman ps": "aaa111111111|web|10.210.0.10",
	}}
	parent := &FirewallFlags{
		SSHMeshFlags: common.SSHMeshFlags{
			Servers: []string{"h1"}, SSHUser: "root", SSHPort: 22, Concurrency: 1,
		},
		PodmanNetworkName: "coolify-mesh",
	}
	local := &allowRevokeFlags{
		From: "10.210.1.5", To: "web", Proto: "tcp", Port: 80,
	}
	inner := &cobra.Command{Use: "revoke"}
	rootCmdFor(inner)

	err := emitAllowRevoke(context.Background(), inner, parent, local, fr, true)
	assert.NoError(t, err)

	var del int
	for _, c := range fr.calls {
		if strings.Contains(c, "iptables -D COOLIFY-ALLOW") {
			del++
		}
	}
	assert.Equal(t, 1, del)
}

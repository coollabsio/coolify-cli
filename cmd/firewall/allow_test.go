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

// cmdFakeRunner matches a Runner call against substrings in its response map
// and returns the first hit. Mirrors cmd/init/plan_test.go's pattern.
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

// parentWithToken builds a FirewallFlags pre-wired for the REST path:
// single test host, coold port 8443, non-empty bearer token.
func parentWithToken() *FirewallFlags {
	return &FirewallFlags{
		SSHMeshFlags: common.SSHMeshFlags{
			Servers: []string{"h1"}, SSHUser: "root", SSHPort: 22, Concurrency: 1,
		},
		Namespace:   common.DefaultNamespace,
		CooldToken:  "test-token",
		CooldPort:   8443,
		WGInterface: "wg0",
	}
}

func TestEmitAllowRevoke_PostsOneAllowToCoold(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{
		"podman ps": "aaa111111111|web|10.210.0.10",
	}}
	parent := parentWithToken()
	local := &allowRevokeFlags{
		From: "10.210.1.5", To: "web", Proto: "tcp", Port: 80,
	}
	inner := &cobra.Command{Use: "allow"}
	rootCmdFor(inner)

	err := emitAllowRevoke(context.Background(), inner, parent, local, fr, false)
	assert.NoError(t, err)

	var posts []string
	for _, c := range fr.calls {
		if strings.Contains(c, "-X POST") && strings.Contains(c, "/api/v1/firewall/allow") {
			posts = append(posts, c)
		}
	}
	assert.Len(t, posts, 1)
	// Token carried in Authorization header.
	assert.Contains(t, posts[0], "Authorization: Bearer test-token")
	// JSON body carries namespace + src/dst/port.
	assert.Contains(t, posts[0], `"namespace":"default"`)
	assert.Contains(t, posts[0], `"src":"10.210.1.5"`)
	assert.Contains(t, posts[0], `"dst":"10.210.0.10"`)
	assert.Contains(t, posts[0], `"port":80`)
	// Discovers mgmt IP via wg0 before curl.
	assert.Contains(t, posts[0], "ip -4 -o addr show wg0")
}

// TestEmitAllowRevoke_CarriesNonDefaultNamespace verifies that the user's
// chosen namespace propagates into the JSON body (and therefore into the
// cid hash coold will compute).
func TestEmitAllowRevoke_CarriesNonDefaultNamespace(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{
		"podman ps": "aaa111111111|web|10.220.0.10",
	}}
	parent := parentWithToken()
	parent.Namespace = "alpha"
	local := &allowRevokeFlags{
		From: "10.220.1.5", To: "web", Proto: "tcp", Port: 80,
	}
	inner := &cobra.Command{Use: "allow"}
	rootCmdFor(inner)

	err := emitAllowRevoke(context.Background(), inner, parent, local, fr, false)
	assert.NoError(t, err)
	var post string
	for _, c := range fr.calls {
		if strings.Contains(c, "-X POST") {
			post = c
		}
	}
	assert.NotEmpty(t, post)
	assert.Contains(t, post, `"namespace":"alpha"`)
	// Discovery targets the alpha-namespace bridge, not the default one.
	var psCalls []string
	for _, c := range fr.calls {
		if strings.Contains(c, "podman ps") {
			psCalls = append(psCalls, c)
		}
	}
	assert.NotEmpty(t, psCalls)
	assert.Contains(t, psCalls[0], "coolify-alpha-mesh")
}

func TestEmitAllowRevoke_Bidirectional(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{
		"podman ps": "aaa111111111|web|10.210.0.10\nbbb222222222|client|10.210.1.5",
	}}
	parent := parentWithToken()
	local := &allowRevokeFlags{
		From: "10.210.1.5", To: "10.210.0.10", Proto: "tcp", Port: 80, Bidirectional: true,
	}
	inner := &cobra.Command{Use: "allow"}
	rootCmdFor(inner)

	err := emitAllowRevoke(context.Background(), inner, parent, local, fr, false)
	assert.NoError(t, err)

	var posts int
	for _, c := range fr.calls {
		if strings.Contains(c, "-X POST") && strings.Contains(c, "/api/v1/firewall/allow") {
			posts++
		}
	}
	assert.Equal(t, 2, posts)
}

func TestEmitAllowRevoke_RevokeIssuesDelete(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{
		"podman ps": "aaa111111111|web|10.210.0.10",
	}}
	parent := parentWithToken()
	local := &allowRevokeFlags{
		From: "10.210.1.5", To: "web", Proto: "tcp", Port: 80,
	}
	inner := &cobra.Command{Use: "revoke"}
	rootCmdFor(inner)

	err := emitAllowRevoke(context.Background(), inner, parent, local, fr, true)
	assert.NoError(t, err)

	var deletes []string
	for _, c := range fr.calls {
		if strings.Contains(c, "-X DELETE") && strings.Contains(c, "/api/v1/firewall/allow/") {
			deletes = append(deletes, c)
		}
	}
	assert.Len(t, deletes, 1)
	assert.Contains(t, deletes[0], "Authorization: Bearer test-token")
}

func TestEmitAllowRevoke_FetchesTokenPerHostWhenOverrideAbsent(t *testing.T) {
	// No --coold-token override → CLI SSHes `cat /etc/coolify/api-token`
	// on the destination host and uses the result as the bearer.
	fr := &cmdFakeRunner{responses: map[string]string{
		"podman ps":              "aaa111111111|web|10.210.0.10",
		"/etc/coolify/api-token": "per-host-token\n",
	}}
	parent := parentWithToken()
	parent.CooldToken = ""
	t.Setenv("COOLIFY_COOLD_TOKEN", "")
	local := &allowRevokeFlags{
		From: "10.210.1.5", To: "web", Proto: "tcp", Port: 80,
	}
	inner := &cobra.Command{Use: "allow"}
	rootCmdFor(inner)

	err := emitAllowRevoke(context.Background(), inner, parent, local, fr, false)
	assert.NoError(t, err)
	var post string
	for _, c := range fr.calls {
		if strings.Contains(c, "-X POST") && strings.Contains(c, "/api/v1/firewall/allow") {
			post = c
		}
	}
	assert.NotEmpty(t, post)
	assert.Contains(t, post, "Authorization: Bearer per-host-token")
}

func TestEmitAllowRevoke_FetchFailurePropagates(t *testing.T) {
	// Empty /etc/coolify/api-token on the host → FetchCooldToken errors,
	// and the error surfaces to the caller instead of silently proceeding.
	fr := &cmdFakeRunner{responses: map[string]string{
		"podman ps": "aaa111111111|web|10.210.0.10",
		// No token file → empty stdout → "token is empty" error.
	}}
	parent := parentWithToken()
	parent.CooldToken = ""
	t.Setenv("COOLIFY_COOLD_TOKEN", "")
	local := &allowRevokeFlags{
		From: "10.210.1.5", To: "web", Proto: "tcp", Port: 80,
	}
	inner := &cobra.Command{Use: "allow"}
	rootCmdFor(inner)

	err := emitAllowRevoke(context.Background(), inner, parent, local, fr, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "coold token")
}

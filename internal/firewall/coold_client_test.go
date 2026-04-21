package firewall

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// fakeCooldRunner is a minimal Runner for client-level tests. It captures
// every command and replies based on substring-matched canned responses.
// mu guards calls against concurrent appends from ForEachServer's parallel
// goroutines.
type fakeCooldRunner struct {
	mu        sync.Mutex
	responses map[string]string
	calls     []string
}

func (f *fakeCooldRunner) Run(_ context.Context, _, _ string, _ int, cmd string) (string, string, error) {
	f.mu.Lock()
	f.calls = append(f.calls, cmd)
	f.mu.Unlock()
	for sub, resp := range f.responses {
		if strings.Contains(cmd, sub) {
			return resp, "", nil
		}
	}
	return "", "", nil
}

var _ ssh.Runner = (*fakeCooldRunner)(nil)

func TestShellSingleQuote_Escapes(t *testing.T) {
	assert.Equal(t, `'plain'`, shellSingleQuote("plain"))
	assert.Equal(t, `'it'\''s'`, shellSingleQuote("it's"))
}

func TestBuildCurlAllow_Shape(t *testing.T) {
	cmd := buildCurlAllow("wg0", "tok-xyz", 8443, `{"src":"10.0.0.1","dst":"10.0.0.2"}`)
	assert.Contains(t, cmd, "ip -4 -o addr show wg0")
	assert.Contains(t, cmd, "curl -fsS")
	assert.Contains(t, cmd, "Authorization: Bearer tok-xyz")
	assert.Contains(t, cmd, "Content-Type: application/json")
	assert.Contains(t, cmd, "-X POST")
	assert.Contains(t, cmd, `{"src":"10.0.0.1","dst":"10.0.0.2"}`)
	assert.Contains(t, cmd, `:8443/api/v1/firewall/allow`)
}

func TestBuildCurlRevoke_Shape(t *testing.T) {
	cmd := buildCurlRevoke("wg0", "tok-xyz", 8443, "abc123def456")
	assert.Contains(t, cmd, "curl -fsS")
	assert.Contains(t, cmd, "-X DELETE")
	assert.Contains(t, cmd, "Authorization: Bearer tok-xyz")
	assert.Contains(t, cmd, `:8443/api/v1/firewall/allow/abc123def456`)
}

func TestBuildCurlList_SoftMgmtIP(t *testing.T) {
	cmd := buildCurlList("wg0", "tok-xyz", 8443, "")
	// Missing wg0 yields an empty array and success exit.
	assert.Contains(t, cmd, `echo '[]'; exit 0`)
	assert.Contains(t, cmd, "Authorization: Bearer tok-xyz")
	assert.Contains(t, cmd, `:8443/api/v1/firewall/allow`)
	// Empty namespace → no query string.
	assert.NotContains(t, cmd, "namespace=")
}

// TestBuildCurlList_WithNamespace verifies that a non-empty namespace is
// forwarded as ?namespace=<ns> so coold can filter on its side.
func TestBuildCurlList_WithNamespace(t *testing.T) {
	cmd := buildCurlList("wg0", "tok-xyz", 8443, "alpha")
	assert.Contains(t, cmd, `:8443/api/v1/firewall/allow?namespace=alpha`)
}

func TestCooldApply_SendsJSONPayload(t *testing.T) {
	fr := &fakeCooldRunner{}
	r := AllowRule{
		Src: net.ParseIP("10.0.0.1"), Dst: net.ParseIP("10.0.0.2"),
		Proto: "tcp", Port: 80,
	}
	err := CooldApply(context.Background(), fr, "h1", "root", 22, 8443, "wg0", "t", r)
	assert.NoError(t, err)
	assert.Len(t, fr.calls, 1)
	assert.Contains(t, fr.calls[0], `"src":"10.0.0.1"`)
	assert.Contains(t, fr.calls[0], `"dst":"10.0.0.2"`)
	assert.Contains(t, fr.calls[0], `"proto":"tcp"`)
	assert.Contains(t, fr.calls[0], `"port":80`)
}

func TestCooldApply_OmitsProtoWhenEmpty(t *testing.T) {
	fr := &fakeCooldRunner{}
	r := AllowRule{
		Src: net.ParseIP("10.0.0.1"), Dst: net.ParseIP("10.0.0.2"),
	}
	err := CooldApply(context.Background(), fr, "h1", "root", 22, 8443, "wg0", "t", r)
	assert.NoError(t, err)
	// omitempty drops zero port and empty proto — avoids tripping coold's
	// "port requires proto" validation.
	assert.NotContains(t, fr.calls[0], `"proto"`)
	assert.NotContains(t, fr.calls[0], `"port"`)
}

func TestCooldRevoke_RejectsEmptyID(t *testing.T) {
	fr := &fakeCooldRunner{}
	err := CooldRevoke(context.Background(), fr, "h1", "root", 22, 8443, "wg0", "t", "")
	assert.Error(t, err)
	assert.Empty(t, fr.calls, "no SSH call for empty id")
}

func TestCooldList_ParsesJSON(t *testing.T) {
	fr := &fakeCooldRunner{responses: map[string]string{
		"/api/v1/firewall/allow": `[
			{"src":"10.0.0.1","dst":"10.0.0.2","proto":"tcp","port":80,"id":"abc123def456"},
			{"src":"10.0.0.3","dst":"10.0.0.4"}
		]`,
	}}
	rules, err := CooldList(context.Background(), fr, "h1", "root", 22, 8443, "wg0", "t", "")
	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "h1", rules[0].Host)
	assert.Equal(t, "cid:abc123def456", rules[0].Comment)
	assert.Equal(t, "tcp", rules[0].Proto)
	assert.Equal(t, 80, rules[0].Port)
	// Rule without proto/port/id comes through with zero values, no cid.
	assert.Equal(t, "", rules[1].Proto)
	assert.Equal(t, 0, rules[1].Port)
	assert.Equal(t, "", rules[1].Comment)
}

func TestCooldList_EmptyBody(t *testing.T) {
	fr := &fakeCooldRunner{}
	rules, err := CooldList(context.Background(), fr, "h1", "root", 22, 8443, "wg0", "t", "")
	assert.NoError(t, err)
	assert.Empty(t, rules)
}

func TestCooldListAll_SortsByHost(t *testing.T) {
	// Fake returns the same JSON regardless of host; the sort guarantees the
	// fanout output is stable across runs.
	fr := &fakeCooldRunner{responses: map[string]string{
		"/api/v1/firewall/allow": `[{"src":"10.0.0.1","dst":"10.0.0.2","proto":"tcp","port":80,"id":"aaa111111111"}]`,
	}}
	tokenFor := func(string) (string, error) { return "t", nil }
	rules, results := CooldListAll(context.Background(), fr,
		[]string{"hB", "hA"}, "root", 22, 8443, "wg0", tokenFor, 2, "")
	assert.Len(t, rules, 2)
	assert.Equal(t, "hA", rules[0].Host)
	assert.Equal(t, "hB", rules[1].Host)
	assert.Len(t, results, 2)
}

func TestFetchCooldToken_ReadsFile(t *testing.T) {
	fr := &fakeCooldRunner{responses: map[string]string{
		"/etc/coolify/api-token": "deadbeefcafe\n",
	}}
	tok, err := FetchCooldToken(context.Background(), fr, "h1", "root", 22)
	assert.NoError(t, err)
	assert.Equal(t, "deadbeefcafe", tok)
}

func TestFetchCooldToken_EmptyErrors(t *testing.T) {
	fr := &fakeCooldRunner{}
	_, err := FetchCooldToken(context.Background(), fr, "h1", "root", 22)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is empty")
}

func TestCooldListAll_PropagatesTokenFetchError(t *testing.T) {
	fr := &fakeCooldRunner{responses: map[string]string{
		"/api/v1/firewall/allow": `[]`,
	}}
	tokenFor := func(h string) (string, error) {
		if h == "hBad" {
			return "", assertError("no token")
		}
		return "t", nil
	}
	_, results := CooldListAll(context.Background(), fr,
		[]string{"hOk", "hBad"}, "root", 22, 8443, "wg0", tokenFor, 2, "")
	var okCount, errCount int
	for _, r := range results {
		if r.Err != nil {
			errCount++
		} else {
			okCount++
		}
	}
	assert.Equal(t, 1, okCount)
	assert.Equal(t, 1, errCount)
}

type assertError string

func (e assertError) Error() string { return string(e) }

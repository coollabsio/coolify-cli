package firewall

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestEmitList_CallsCooldGet(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{
		"/api/v1/firewall/allow": `[{"src":"10.0.0.1","dst":"10.0.0.2","proto":"tcp","port":80,"id":"abc123def456"}]`,
	}}
	parent := parentWithToken()
	inner := &cobra.Command{Use: "list"}
	rootCmdFor(inner)

	err := emitList(context.Background(), inner, parent, fr)
	assert.NoError(t, err)
	assert.Len(t, fr.calls, 1)
	assert.Contains(t, fr.calls[0], "curl")
	assert.Contains(t, fr.calls[0], "/api/v1/firewall/allow")
	assert.Contains(t, fr.calls[0], "Authorization: Bearer test-token")
}

func TestEmitList_EmptyCoold(t *testing.T) {
	fr := &cmdFakeRunner{responses: map[string]string{}}
	parent := parentWithToken()
	inner := &cobra.Command{Use: "list"}
	rootCmdFor(inner)

	err := emitList(context.Background(), inner, parent, fr)
	assert.NoError(t, err)
}

func TestEmitList_FetchesPerHostTokenWhenOverrideAbsent(t *testing.T) {
	// Without --coold-token override, each host's token is read via SSH
	// `cat /etc/coolify/api-token` then used as the bearer for GET /allow.
	fr := &cmdFakeRunner{responses: map[string]string{
		"/etc/coolify/api-token": "per-host-token\n",
		"/api/v1/firewall/allow": `[]`,
	}}
	parent := parentWithToken()
	parent.CooldToken = ""
	t.Setenv("COOLIFY_COOLD_TOKEN", "")
	inner := &cobra.Command{Use: "list"}
	rootCmdFor(inner)

	err := emitList(context.Background(), inner, parent, fr)
	assert.NoError(t, err)
	var ranTokenFetch, ranGet bool
	for _, c := range fr.calls {
		if strings.Contains(c, "cat /etc/coolify/api-token") {
			ranTokenFetch = true
		}
		if strings.Contains(c, "curl") && strings.Contains(c, "Authorization: Bearer per-host-token") {
			ranGet = true
		}
	}
	assert.True(t, ranTokenFetch, "CLI should SSH-fetch the token")
	assert.True(t, ranGet, "bearer should be the fetched token")
}

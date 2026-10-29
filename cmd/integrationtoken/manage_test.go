package integrationtoken

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCommand_CloudflareOptions(t *testing.T) {
	fake := newFakeAPI(t, createdHandler)

	_, _, err := runCmd(t, "cf-secret\n", "integration-token", "create", "--provider", "cloudflare", "--name", "cf",
		"--capabilities", "dns,tunnel", "--automatic-dns=false", "--provider-token-stdin")
	require.NoError(t, err)

	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.JSONEq(t, `{"provider":"cloudflare","name":"cf","token":"cf-secret","capabilities":["dns","tunnel"],"automatic_dns":false}`, reqs[0].Body)
}

func TestUpdateCommand_SendsOnlyChangedFields(t *testing.T) {
	fake := newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"uuid":"tok-uuid","name":"cf","provider":"cloudflare","capabilities":["dns"],"automatic_dns":false}`))
	})

	stdout, stderr, err := runCmd(t, "rotated-secret\n", "integration-token", "update", "tok-uuid", "--proxied-dns", "--rotate", "--provider-token-stdin")
	require.NoError(t, err)

	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, http.MethodPatch, reqs[0].Method)
	assert.Equal(t, "/api/v1/security/integration-tokens/tok-uuid", reqs[0].RequestURI)
	assert.JSONEq(t, `{"token":"rotated-secret","proxied_dns":true}`, reqs[0].Body)
	assert.NotContains(t, stdout+stderr, "rotated-secret")
}

func TestUpdateCommand_NeedsAChange(t *testing.T) {
	fake := newFakeAPI(t, func(http.ResponseWriter, *http.Request) {})

	_, _, err := runCmd(t, "", "integration-token", "update", "tok-uuid")

	require.ErrorContains(t, err, "nothing to update")
	assert.Empty(t, fake.recorded())
}

func TestDeleteCommand_AsksForConfirmation(t *testing.T) {
	fake := newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	_, _, err := runCmd(t, "no\n", "integration-token", "delete", "tok-uuid")
	require.NoError(t, err)
	assert.Empty(t, fake.recorded())

	_, _, err = runCmd(t, "yes\n", "integration-token", "delete", "tok-uuid")
	require.NoError(t, err)
	require.Len(t, fake.recorded(), 1)
	assert.Equal(t, http.MethodDelete, fake.recorded()[0].Method)
}

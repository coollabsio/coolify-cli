package integrationtoken

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createdHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`{"uuid":"tok-uuid"}`))
}

func TestCreateCommand_TokenFromStdin(t *testing.T) {
	fake := newFakeAPI(t, createdHandler)

	stdout, stderr, err := runCmd(t, "dp.st.supersecret\n",
		"integration-token", "create", "--provider", "doppler", "--name", "prod", "--provider-token-stdin")
	require.NoError(t, err)

	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, http.MethodPost, reqs[0].Method)
	assert.Equal(t, "/api/v1/security/integration-tokens", reqs[0].RequestURI)
	assert.JSONEq(t, `{"provider":"doppler","name":"prod","token":"dp.st.supersecret"}`, reqs[0].Body)

	assert.Contains(t, stdout, "tok-uuid")
	assert.NotContains(t, stdout, "supersecret")
	assert.NotContains(t, stderr, "supersecret")
}

func TestCreateCommand_TokenFromEnvWithMetadata(t *testing.T) {
	fake := newFakeAPI(t, createdHandler)
	t.Setenv(TokenEnvVar, "hvs.envsecret")

	stdout, _, err := runCmd(t, "", "integration-tokens", "create", "--provider", "vault", "--name", "vault",
		"--metadata", `{"base_url":"https://vault.example.com","namespace":"admin"}`, "--format", "json")
	require.NoError(t, err)

	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(reqs[0].Body), &body))
	assert.Equal(t, "hvs.envsecret", body["token"])
	assert.Equal(t, map[string]any{"base_url": "https://vault.example.com", "namespace": "admin"}, body["metadata"])

	assert.JSONEq(t, `{"uuid":"tok-uuid"}`, stdout)
}

func TestCreateCommand_HiddenPrompt(t *testing.T) {
	fake := newFakeAPI(t, createdHandler)
	t.Setenv(TokenEnvVar, "")

	origFD, origRead := terminalFD, readHiddenToken
	t.Cleanup(func() { terminalFD, readHiddenToken = origFD, origRead })
	terminalFD = func(io.Reader) (int, bool) { return 0, true }
	readHiddenToken = func(int) ([]byte, error) { return []byte("dp.sa.prompted"), nil }

	_, stderr, err := runCmd(t, "", "integration-token", "create", "--provider", "doppler", "--name", "prod")
	require.NoError(t, err)
	assert.Contains(t, stderr, "Provider token:")
	assert.NotContains(t, stderr, "prompted")
	require.Len(t, fake.recorded(), 1)
	assert.Contains(t, fake.recorded()[0].Body, `"token":"dp.sa.prompted"`)

	readHiddenToken = func(int) ([]byte, error) { return nil, errors.New("boom") }
	_, _, err = runCmd(t, "", "integration-token", "create", "--provider", "doppler", "--name", "prod")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read provider token")
}

func TestCreateCommand_Errors(t *testing.T) {
	tests := []struct {
		name    string
		stdin   string
		args    []string
		wantErr string
	}{
		{name: "no token source", args: []string{"--provider", "doppler", "--name", "n"}, wantErr: "no provider token given"},
		{name: "empty stdin", args: []string{"--provider", "doppler", "--name", "n", "--provider-token-stdin"}, stdin: "\n", wantErr: "must not be empty"},
		{name: "invalid provider", args: []string{"--provider", "aws", "--name", "n", "--provider-token-stdin"}, stdin: "x", wantErr: "--provider must be one of: doppler, infisical, vault"},
		{name: "missing name", args: []string{"--provider", "doppler", "--provider-token-stdin"}, stdin: "dp.st.x", wantErr: "--name is required"},
		{name: "bad metadata json", args: []string{"--provider", "vault", "--name", "n", "--metadata", "{nope", "--provider-token-stdin"}, stdin: "x", wantErr: "--metadata must be a JSON object"},
		{name: "infisical needs metadata", args: []string{"--provider", "infisical", "--name", "n", "--provider-token-stdin"}, stdin: "x", wantErr: "client_id"},
		{name: "doppler prefix", args: []string{"--provider", "doppler", "--name", "n", "--provider-token-stdin"}, stdin: "nope", wantErr: "dp.st."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(TokenEnvVar, "")
			fake := newFakeAPI(t, createdHandler)
			_, _, err := runCmd(t, tt.stdin, append([]string{"integration-token", "create"}, tt.args...)...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Empty(t, fake.recorded())
		})
	}
}

func TestCreateCommand_TokenFlagIsNotAccepted(t *testing.T) {
	cmd := NewCreateCommand()
	assert.Nil(t, cmd.Flags().Lookup("provider-token"))
	assert.Nil(t, cmd.Flags().Lookup("token"))
}

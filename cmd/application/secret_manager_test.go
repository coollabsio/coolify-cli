package application

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const secretManagerResponse = `{"integration_token_uuid":"tok-uuid","provider":"vault","settings":{"mount":"secret","path":"apps/web"}}`

func TestSecretManagerSet_SendsRequest(t *testing.T) {
	fake := newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(secretManagerResponse))
	})

	stdout, _, err := runCmd(t, "", "app", "secret-manager", "set", "app-uuid",
		"--integration-token", "tok-uuid", "--settings", `{"mount":"secret","path":"apps/web"}`)
	require.NoError(t, err)

	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, http.MethodPatch, reqs[0].Method)
	assert.Equal(t, "/api/v1/applications/app-uuid/secret-manager", reqs[0].RequestURI)
	assert.JSONEq(t, `{"integration_token_uuid":"tok-uuid","settings":{"mount":"secret","path":"apps/web"}}`, reqs[0].Body)

	assert.Contains(t, stdout, "tok-uuid")
	assert.Contains(t, stdout, "vault")
	assert.Contains(t, stdout, `{"mount":"secret","path":"apps/web"}`)
}

func TestSecretManagerSet_WithoutSettingsAndJSON(t *testing.T) {
	fake := newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"integration_token_uuid":"tok-uuid","provider":"doppler","settings":null}`))
	})

	stdout, _, err := runCmd(t, "", "app", "secret-manager", "set", "app-uuid", "--integration-token", "tok-uuid", "--format", "json")
	require.NoError(t, err)
	require.Len(t, fake.recorded(), 1)
	assert.JSONEq(t, `{"integration_token_uuid":"tok-uuid"}`, fake.recorded()[0].Body)

	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "doppler", out["provider"])
}

func TestSecretManagerSet_Validation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "missing token", args: []string{"app-uuid"}, wantErr: "--integration-token is required"},
		{name: "bad settings", args: []string{"app-uuid", "--integration-token", "t", "--settings", "[1]"}, wantErr: "--settings must be a JSON object"},
		{name: "missing app uuid", args: []string{"--integration-token", "t"}, wantErr: "app_uuid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t, func(http.ResponseWriter, *http.Request) {})
			_, _, err := runCmd(t, "", append([]string{"app", "secret-manager", "set"}, tt.args...)...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Empty(t, fake.recorded())
		})
	}
}

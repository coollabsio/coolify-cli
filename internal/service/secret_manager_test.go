package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestSecretManagerService_CreateIntegrationToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/security/integration-tokens", r.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, map[string]any{
			"provider": "vault",
			"name":     "prod vault",
			"token":    "hvs.secret",
			"metadata": map[string]any{"base_url": "https://vault.example.com", "namespace": "admin"},
		}, body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"uuid":"tok-uuid"}`))
	}))
	defer server.Close()

	svc := NewSecretManagerService(api.NewClient(server.URL, "test-token"))
	resp, err := svc.CreateIntegrationToken(context.Background(), models.IntegrationTokenCreateRequest{
		Provider: "vault",
		Name:     "prod vault",
		Token:    "hvs.secret",
		Metadata: map[string]any{"base_url": "https://vault.example.com", "namespace": "admin"},
	})
	require.NoError(t, err)
	assert.Equal(t, "tok-uuid", resp.UUID)
}

func TestSecretManagerService_CreateIntegrationToken_OmitsEmptyMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		_, hasMetadata := body["metadata"]
		assert.False(t, hasMetadata)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"uuid":"tok-uuid"}`))
	}))
	defer server.Close()

	svc := NewSecretManagerService(api.NewClient(server.URL, "test-token"))
	_, err := svc.CreateIntegrationToken(context.Background(), models.IntegrationTokenCreateRequest{
		Provider: "doppler", Name: "doppler", Token: "dp.st.abc",
	})
	require.NoError(t, err)
}

func TestSecretManagerService_CreateIntegrationToken_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"The Doppler token could not be validated."}`))
	}))
	defer server.Close()

	svc := NewSecretManagerService(api.NewClient(server.URL, "test-token"))
	_, err := svc.CreateIntegrationToken(context.Background(), models.IntegrationTokenCreateRequest{
		Provider: "doppler", Name: "doppler", Token: "dp.st.abc",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not be validated")
	assert.NotContains(t, err.Error(), "dp.st.abc")
}

func TestValidateIntegrationTokenCreateRequest(t *testing.T) {
	infisicalMeta := map[string]any{"base_url": "https://app.infisical.com", "client_id": "cid"}
	tests := []struct {
		name    string
		req     models.IntegrationTokenCreateRequest
		wantErr string
	}{
		{name: "doppler service token", req: models.IntegrationTokenCreateRequest{Provider: "doppler", Name: "n", Token: "dp.st.x"}},
		{name: "doppler service account", req: models.IntegrationTokenCreateRequest{Provider: "doppler", Name: "n", Token: "dp.sa.x"}},
		{name: "infisical", req: models.IntegrationTokenCreateRequest{Provider: "infisical", Name: "n", Token: "secret", Metadata: infisicalMeta}},
		{name: "vault", req: models.IntegrationTokenCreateRequest{Provider: "vault", Name: "n", Token: "hvs", Metadata: map[string]any{"base_url": "https://v"}}},
		{name: "unknown provider", req: models.IntegrationTokenCreateRequest{Provider: "aws", Name: "n", Token: "t"}, wantErr: "--provider must be one of: doppler, infisical, vault"},
		{name: "empty provider", req: models.IntegrationTokenCreateRequest{Name: "n", Token: "t"}, wantErr: "--provider"},
		{name: "missing name", req: models.IntegrationTokenCreateRequest{Provider: "doppler", Token: "dp.st.x"}, wantErr: "--name is required"},
		{name: "missing token", req: models.IntegrationTokenCreateRequest{Provider: "doppler", Name: "n"}, wantErr: "must not be empty"},
		{name: "bad doppler prefix", req: models.IntegrationTokenCreateRequest{Provider: "doppler", Name: "n", Token: "abc"}, wantErr: "dp.st."},
		{name: "infisical missing client id", req: models.IntegrationTokenCreateRequest{Provider: "infisical", Name: "n", Token: "t", Metadata: map[string]any{"base_url": "https://x"}}, wantErr: "client_id"},
		{name: "vault missing base url", req: models.IntegrationTokenCreateRequest{Provider: "vault", Name: "n", Token: "t"}, wantErr: "base_url"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntegrationTokenCreateRequest(tt.req)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestSecretManagerService_CreateIntegrationToken_InvalidSkipsRequest(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { calls.Add(1) }))
	defer server.Close()

	svc := NewSecretManagerService(api.NewClient(server.URL, "test-token"))
	_, err := svc.CreateIntegrationToken(context.Background(), models.IntegrationTokenCreateRequest{Provider: "nope"})
	require.Error(t, err)
	assert.Zero(t, calls.Load())
}

func TestSecretManagerService_ConfigureApplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/api/v1/applications/app-uuid/secret-manager", r.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, map[string]any{
			"integration_token_uuid": "tok-uuid",
			"settings":               map[string]any{"mount": "secret", "path": "apps/web"},
		}, body)
		_, _ = w.Write([]byte(`{"integration_token_uuid":"tok-uuid","provider":"vault","settings":{"mount":"secret","path":"apps/web"}}`))
	}))
	defer server.Close()

	svc := NewSecretManagerService(api.NewClient(server.URL, "test-token"))
	resp, err := svc.ConfigureApplication(context.Background(), "app-uuid", models.ApplicationSecretManagerRequest{
		IntegrationTokenUUID: "tok-uuid",
		Settings:             map[string]any{"mount": "secret", "path": "apps/web"},
	})
	require.NoError(t, err)
	assert.Equal(t, "tok-uuid", resp.IntegrationTokenUUID)
	assert.Equal(t, "vault", resp.Provider)
	assert.Equal(t, "apps/web", resp.Settings["path"])
}

func TestSecretManagerService_ConfigureApplication_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Secret manager integration token not found."}`))
	}))
	defer server.Close()

	svc := NewSecretManagerService(api.NewClient(server.URL, "test-token"))
	_, err := svc.ConfigureApplication(context.Background(), "app-uuid", models.ApplicationSecretManagerRequest{IntegrationTokenUUID: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "integration token not found")

	_, err = svc.ConfigureApplication(context.Background(), "app-uuid", models.ApplicationSecretManagerRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--integration-token is required")
}

func TestParseJSONObject(t *testing.T) {
	got, err := ParseJSONObject("settings", "")
	require.NoError(t, err)
	assert.Nil(t, got)

	got, err = ParseJSONObject("settings", `{"project":"web","config":"prd"}`)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"project": "web", "config": "prd"}, got)

	for _, bad := range []string{`[1]`, `"x"`, `{bad`, `null`} {
		_, err = ParseJSONObject("settings", bad)
		require.Error(t, err, bad)
		assert.Contains(t, err.Error(), "--settings must be a JSON object")
	}
}

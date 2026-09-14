package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestSecretManagerService_CreatesTokenAndConfiguresApplication(t *testing.T) {
	type observed struct {
		method string
		path   string
		body   map[string]any
	}
	requests := make(chan observed, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		item := observed{method: r.Method, path: r.URL.EscapedPath()}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&item.body))
		requests <- item
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"uuid":"token-1"}`))
			return
		}
		_, _ = w.Write([]byte(`{"integration_token_uuid":"token-1","provider":"doppler","settings":{"project":"web","config":"production"}}`))
	}))
	defer server.Close()

	svc := NewSecretManagerService(api.NewClient(server.URL, "token", api.WithRetries(0)))
	created, err := svc.CreateToken(context.Background(), models.SecretManagerTokenCreateRequest{
		Provider: "doppler", Name: "Production", Token: "dp.sa.secret",
	})
	require.NoError(t, err)
	assert.Equal(t, "token-1", created.UUID)

	configured, err := svc.ConfigureApplication(context.Background(), "app/id", models.ApplicationSecretManagerRequest{
		IntegrationTokenUUID: "token-1",
		Settings:             map[string]string{"project": "web", "config": "production"},
	})
	require.NoError(t, err)
	assert.Equal(t, "doppler", configured.Provider)

	createRequest := <-requests
	assert.Equal(t, http.MethodPost, createRequest.method)
	assert.Equal(t, "/api/v1/security/integration-tokens", createRequest.path)
	assert.Equal(t, "dp.sa.secret", createRequest.body["token"])

	configureRequest := <-requests
	assert.Equal(t, http.MethodPatch, configureRequest.method)
	assert.Equal(t, "/api/v1/applications/app%2Fid/secret-manager", configureRequest.path)
	assert.Equal(t, "token-1", configureRequest.body["integration_token_uuid"])
}

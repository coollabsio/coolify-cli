package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrivateKeyService_List(t *testing.T) {
	keys := []models.PrivateKey{
		{
			UUID: "key-1",
			Name: "Test Key 1",
		},
		{
			UUID: "key-2",
			Name: "Test Key 2",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/security/keys", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(keys)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewPrivateKeyService(client)

	result, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "key-1", result[0].UUID)
	assert.Equal(t, "Test Key 1", result[0].Name)
}

func TestPrivateKeyService_Create(t *testing.T) {
	req := models.PrivateKeyCreateRequest{
		Name:       "New Key",
		PrivateKey: "ssh-rsa AAAAB3...",
	}

	key := models.PrivateKey{
		UUID: "key-123",
		Name: req.Name,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/security/keys", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var receivedReq models.PrivateKeyCreateRequest
		json.NewDecoder(r.Body).Decode(&receivedReq)
		assert.Equal(t, req.Name, receivedReq.Name)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(key)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewPrivateKeyService(client)

	result, err := svc.Create(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "key-123", result.UUID)
	assert.Equal(t, "New Key", result.Name)
}

func TestPrivateKeyService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/security/keys/key-123", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewPrivateKeyService(client)

	err := svc.Delete(context.Background(), "key-123")
	require.NoError(t, err)
}

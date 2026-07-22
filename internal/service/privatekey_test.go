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
		_ = json.NewEncoder(w).Encode(keys)
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

func TestPrivateKeyService_Get(t *testing.T) {
	desc := "My deploy key"
	key := models.PrivateKey{
		UUID:        "key-123",
		Name:        "Test Key",
		Description: &desc,
		PublicKey:   "ssh-rsa AAAAB3...",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/security/keys/key-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(key)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewPrivateKeyService(client)

	result, err := svc.Get(context.Background(), "key-123")
	require.NoError(t, err)
	assert.Equal(t, "key-123", result.UUID)
	assert.Equal(t, "Test Key", result.Name)
	require.NotNil(t, result.Description)
	assert.Equal(t, "My deploy key", *result.Description)
	assert.Equal(t, "ssh-rsa AAAAB3...", result.PublicKey)
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
		_ = json.NewDecoder(r.Body).Decode(&receivedReq)
		assert.Equal(t, req.Name, receivedReq.Name)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(key)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewPrivateKeyService(client)

	result, err := svc.Create(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "key-123", result.UUID)
	assert.Equal(t, "New Key", result.Name)
}

func TestPrivateKeyService_Update(t *testing.T) {
	name := "Updated Key"
	desc := "Updated description"
	req := models.PrivateKeyUpdateRequest{
		Name:        &name,
		Description: &desc,
		PrivateKey:  "-----BEGIN OPENSSH PRIVATE KEY-----\ntest\n-----END OPENSSH PRIVATE KEY-----",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/security/keys/key-123", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var receivedReq models.PrivateKeyUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&receivedReq)
		require.NotNil(t, receivedReq.Name)
		assert.Equal(t, "Updated Key", *receivedReq.Name)
		require.NotNil(t, receivedReq.Description)
		assert.Equal(t, "Updated description", *receivedReq.Description)
		assert.Equal(t, req.PrivateKey, receivedReq.PrivateKey)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.PrivateKey{UUID: "key-123"})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewPrivateKeyService(client)

	result, err := svc.Update(context.Background(), "key-123", req)
	require.NoError(t, err)
	assert.Equal(t, "key-123", result.UUID)
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

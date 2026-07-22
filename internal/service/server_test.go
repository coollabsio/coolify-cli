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

func TestServerService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		servers := []models.Server{
			{UUID: "uuid-1", Name: "server-1"},
			{UUID: "uuid-2", Name: "server-2"},
		}
		_ = json.NewEncoder(w).Encode(servers)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServerService(client)

	servers, err := svc.List(context.Background())

	require.NoError(t, err)
	assert.Len(t, servers, 2)
	assert.Equal(t, "uuid-1", servers[0].UUID)
	assert.Equal(t, "server-1", servers[0].Name)
}

func TestServerService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/test-uuid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		server := models.Server{
			UUID: "test-uuid",
			Name: "test-server",
			IP:   "192.168.1.100",
		}
		_ = json.NewEncoder(w).Encode(server)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServerService(client)

	result, err := svc.Get(context.Background(), "test-uuid")

	require.NoError(t, err)
	assert.Equal(t, "test-uuid", result.UUID)
	assert.Equal(t, "test-server", result.Name)
	assert.Equal(t, "192.168.1.100", result.IP)
}

func TestServerService_GetResources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/test-uuid", r.URL.Path)
		assert.Equal(t, "resources=true", r.URL.RawQuery)
		assert.Equal(t, "GET", r.Method)

		resources := models.Resources{
			Resources: []models.Resource{
				{UUID: "res-1", Name: "resource-1", Type: "application"},
			},
		}
		_ = json.NewEncoder(w).Encode(resources)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServerService(client)

	result, err := svc.GetResources(context.Background(), "test-uuid")

	require.NoError(t, err)
	assert.Len(t, result.Resources, 1)
	assert.Equal(t, "res-1", result.Resources[0].UUID)
}

func TestServerService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.ServerCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "new-server", req.Name)
		assert.Equal(t, "192.168.1.200", req.IP)

		response := models.Response{Message: "Server created"}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServerService(client)

	req := models.ServerCreateRequest{
		Name:           "new-server",
		IP:             "192.168.1.200",
		Port:           22,
		User:           "root",
		PrivateKeyUUID: "key-uuid",
	}

	result, err := svc.Create(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "Server created", result.Message)
}

func TestServerService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/test-uuid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req models.ServerUpdateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.NotNil(t, req.Name)
		assert.Equal(t, "renamed", *req.Name)
		require.NotNil(t, req.IsTerminalEnabled)
		assert.True(t, *req.IsTerminalEnabled)
		assert.Nil(t, req.IP)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.Response{UUID: "test-uuid"})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServerService(client)

	name := "renamed"
	enabled := true
	result, err := svc.Update(context.Background(), "test-uuid", models.ServerUpdateRequest{
		Name:              &name,
		IsTerminalEnabled: &enabled,
	})

	require.NoError(t, err)
	assert.Equal(t, "test-uuid", result.UUID)
}

func TestServerService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/test-uuid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(models.Response{Message: "Server deleted"})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServerService(client)

	err := svc.Delete(context.Background(), "test-uuid")

	require.NoError(t, err)
}

func TestServerService_Validate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/test-uuid/validate", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var request models.ServerValidationRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		assert.False(t, request.Install)

		response := models.Response{Message: "Server is valid"}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServerService(client)

	result, err := svc.Validate(context.Background(), "test-uuid", false)

	require.NoError(t, err)
	assert.Equal(t, "Server is valid", result.Message)
}

func TestServerService_ValidateAndInstall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/servers/test-uuid/validate", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var request models.ServerValidationRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		assert.True(t, request.Install)

		_ = json.NewEncoder(w).Encode(models.Response{Message: "Validation and installation started."})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServerService(client)

	result, err := svc.Validate(context.Background(), "test-uuid", true)

	require.NoError(t, err)
	assert.Equal(t, "Validation and installation started.", result.Message)
}

package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"id": 1,
				"uuid": "service-uuid-1",
				"name": "PostgreSQL",
				"status": "running",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			},
			{
				"id": 2,
				"uuid": "service-uuid-2",
				"name": "Redis",
				"status": "stopped",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}
		]`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	services, err := svc.List(context.Background())

	require.NoError(t, err)
	assert.Len(t, services, 2)
	assert.Equal(t, "service-uuid-1", services[0].UUID)
	assert.Equal(t, "PostgreSQL", services[0].Name)
	assert.Equal(t, "running", services[0].Status)
	assert.Equal(t, "service-uuid-2", services[1].UUID)
	assert.Equal(t, "Redis", services[1].Name)
	assert.Equal(t, "stopped", services[1].Status)
}

func TestServiceService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	services, err := svc.List(context.Background())

	require.NoError(t, err)
	assert.Len(t, services, 0)
}

func TestServiceService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	_, err := svc.List(context.Background())

	require.Error(t, err)
}

func TestServiceService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": 1,
			"uuid": "service-uuid-123",
			"name": "PostgreSQL 16",
			"description": "Production database",
			"status": "running",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z",
			"databases": [
				{
					"id": 10,
					"uuid": "db-uuid-1",
					"name": "main_db",
					"type": "postgresql"
				}
			]
		}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	service, err := svc.Get(context.Background(), "service-uuid-123")

	require.NoError(t, err)
	assert.Equal(t, "service-uuid-123", service.UUID)
	assert.Equal(t, "PostgreSQL 16", service.Name)
	assert.Equal(t, "running", service.Status)
	assert.NotNil(t, service.Description)
	assert.Equal(t, "Production database", *service.Description)
	assert.Len(t, service.Databases, 1)
	assert.Equal(t, "db-uuid-1", service.Databases[0].UUID)
}

func TestServiceService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "service not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	_, err := svc.Get(context.Background(), "nonexistent")

	require.Error(t, err)
}

func TestServiceService_Start(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/start", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Service starting request queued."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	resp, err := svc.Start(context.Background(), "service-uuid-123")

	require.NoError(t, err)
	assert.Equal(t, "Service starting request queued.", resp.Message)
}

func TestServiceService_Start_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "service already running"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	_, err := svc.Start(context.Background(), "service-uuid-123")

	require.Error(t, err)
}

func TestServiceService_Stop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/stop", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Service stopping request queued."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	resp, err := svc.Stop(context.Background(), "service-uuid-123")

	require.NoError(t, err)
	assert.Equal(t, "Service stopping request queued.", resp.Message)
}

func TestServiceService_Stop_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "service already stopped"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	_, err := svc.Stop(context.Background(), "service-uuid-123")

	require.Error(t, err)
}

func TestServiceService_Restart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/restart", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Service restarting request queued."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	resp, err := svc.Restart(context.Background(), "service-uuid-123")

	require.NoError(t, err)
	assert.Equal(t, "Service restarting request queued.", resp.Message)
}

func TestServiceService_Restart_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "service not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	_, err := svc.Restart(context.Background(), "service-uuid-123")

	require.Error(t, err)
}

func TestServiceService_ListEnvs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"uuid": "env-1",
				"key": "DATABASE_URL",
				"value": "postgres://localhost",
				"is_build_time": false,
				"is_preview": false
			},
			{
				"uuid": "env-2",
				"key": "API_KEY",
				"value": "secret",
				"is_build_time": true,
				"is_preview": false
			}
		]`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	envs, err := svc.ListEnvs(context.Background(), "service-uuid-123")

	require.NoError(t, err)
	assert.Len(t, envs, 2)
	assert.Equal(t, "DATABASE_URL", envs[0].Key)
	assert.Equal(t, "API_KEY", envs[1].Key)
}

func TestServiceService_CreateEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"uuid": "env-new",
			"key": "NEW_VAR",
			"value": "new_value",
			"is_build_time": false,
			"is_preview": false
		}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	env, err := svc.CreateEnv(context.Background(), "service-uuid-123", &models.EnvironmentVariableCreateRequest{
		Key:   "NEW_VAR",
		Value: "new_value",
	})

	require.NoError(t, err)
	assert.Equal(t, "NEW_VAR", env.Key)
	assert.Equal(t, "new_value", env.Value)
}

func TestServiceService_UpdateEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"uuid": "env-123",
			"key": "UPDATED_VAR",
			"value": "updated_value",
			"is_build_time": true,
			"is_preview": false
		}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	newKey := "UPDATED_VAR"
	env, err := svc.UpdateEnv(context.Background(), "service-uuid-123", &models.EnvironmentVariableUpdateRequest{
		UUID: "env-123",
		Key:  &newKey,
	})

	require.NoError(t, err)
	assert.Equal(t, "UPDATED_VAR", env.Key)
}

func TestServiceService_DeleteEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/envs/env-456", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewServiceService(client)

	err := svc.DeleteEnv(context.Background(), "service-uuid-123", "env-456")

	require.NoError(t, err)
}

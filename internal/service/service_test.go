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

func TestService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
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
	svc := NewService(client)

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

func TestService_Logs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-1/logs", r.URL.Path)
		assert.Equal(t, "app", r.URL.Query().Get("sub_service_name"))
		assert.Equal(t, "75", r.URL.Query().Get("lines"))
		assert.Equal(t, "true", r.URL.Query().Get("show_timestamps"))
		_, _ = w.Write([]byte(`{"logs":"service output"}`))
	}))
	defer server.Close()

	svc := NewService(api.NewClient(server.URL, "test-token"))
	response, err := svc.Logs(context.Background(), "service-1", "app", 75, true)

	require.NoError(t, err)
	assert.Equal(t, "service output", response.Logs)
}

func TestService_Move(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-1/move", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		var request models.MoveResourceRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		assert.Equal(t, "env-2", request.EnvironmentUUID)
		_, _ = w.Write([]byte(`{"message":"Service moved successfully.","uuid":"service-1","project_uuid":"project-1","environment_uuid":"env-2"}`))
	}))
	defer server.Close()

	svc := NewService(api.NewClient(server.URL, "test-token"))
	response, err := svc.Move(context.Background(), "service-1", "env-2")

	require.NoError(t, err)
	assert.Equal(t, "service-1", response.UUID)
}

func TestService_ServiceApplicationOperations(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		call   func(*Service) error
	}{
		{name: "list", method: http.MethodGet, path: "/api/v1/services/service-1/applications", call: func(s *Service) error { _, err := s.ListApplications(context.Background(), "service-1"); return err }},
		{name: "get", method: http.MethodGet, path: "/api/v1/services/service-1/applications/app-1", call: func(s *Service) error {
			_, err := s.GetApplication(context.Background(), "service-1", "app-1")
			return err
		}},
		{name: "update", method: http.MethodPatch, path: "/api/v1/services/service-1/applications/app-1?force_domain_override=true", call: func(s *Service) error {
			_, err := s.UpdateApplication(context.Background(), "service-1", "app-1", true, &models.ServiceApplicationUpdateRequest{HumanName: stringPointer("Frontend")})
			return err
		}},
		{name: "logs", method: http.MethodGet, path: "/api/v1/services/service-1/applications/app-1/logs?lines=42", call: func(s *Service) error {
			_, err := s.ApplicationLogs(context.Background(), "service-1", "app-1", 42)
			return err
		}},
		{name: "start", method: http.MethodPost, path: "/api/v1/services/service-1/applications/app-1/start?force=true&latest=true", call: func(s *Service) error {
			_, err := s.StartApplication(context.Background(), "service-1", "app-1", true, true)
			return err
		}},
		{name: "restart", method: http.MethodPost, path: "/api/v1/services/service-1/applications/app-1/restart", call: func(s *Service) error {
			_, err := s.RestartApplication(context.Background(), "service-1", "app-1")
			return err
		}},
		{name: "stop", method: http.MethodPost, path: "/api/v1/services/service-1/applications/app-1/stop", call: func(s *Service) error {
			_, err := s.StopApplication(context.Background(), "service-1", "app-1")
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.path, r.URL.RequestURI())
				if tt.name == "list" {
					_, _ = w.Write([]byte(`[]`))
					return
				}
				if tt.name == "logs" {
					_, _ = w.Write([]byte(`{"logs":"output"}`))
					return
				}
				if tt.name == "start" || tt.name == "restart" || tt.name == "stop" {
					_, _ = w.Write([]byte(`{"message":"queued"}`))
					return
				}
				_, _ = w.Write([]byte(`{"uuid":"app-1","name":"app"}`))
			}))
			defer server.Close()

			err := tt.call(NewService(api.NewClient(server.URL, "test-token")))
			require.NoError(t, err)
		})
	}
}

func TestService_ServiceDatabaseOperations(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		call   func(*Service) error
	}{
		{name: "list", method: http.MethodGet, path: "/api/v1/services/service-1/databases", call: func(s *Service) error { _, err := s.ListDatabases(context.Background(), "service-1"); return err }},
		{name: "get", method: http.MethodGet, path: "/api/v1/services/service-1/databases/database-1", call: func(s *Service) error {
			_, err := s.GetDatabase(context.Background(), "service-1", "database-1")
			return err
		}},
		{name: "update", method: http.MethodPatch, path: "/api/v1/services/service-1/databases/database-1", call: func(s *Service) error {
			_, err := s.UpdateDatabase(context.Background(), "service-1", "database-1", &models.ServiceDatabaseUpdateRequest{HumanName: stringPointer("Primary")})
			return err
		}},
		{name: "logs", method: http.MethodGet, path: "/api/v1/services/service-1/databases/database-1/logs?lines=42", call: func(s *Service) error {
			_, err := s.DatabaseLogs(context.Background(), "service-1", "database-1", 42)
			return err
		}},
		{name: "start", method: http.MethodPost, path: "/api/v1/services/service-1/databases/database-1/start?force=true&latest=true", call: func(s *Service) error {
			_, err := s.StartDatabase(context.Background(), "service-1", "database-1", true, true)
			return err
		}},
		{name: "restart", method: http.MethodPost, path: "/api/v1/services/service-1/databases/database-1/restart", call: func(s *Service) error {
			_, err := s.RestartDatabase(context.Background(), "service-1", "database-1")
			return err
		}},
		{name: "stop", method: http.MethodPost, path: "/api/v1/services/service-1/databases/database-1/stop", call: func(s *Service) error {
			_, err := s.StopDatabase(context.Background(), "service-1", "database-1")
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.path, r.URL.RequestURI())
				switch tt.name {
				case "list":
					_, _ = w.Write([]byte(`[]`))
				case "logs":
					_, _ = w.Write([]byte(`{"logs":"output"}`))
				case "start", "restart", "stop":
					_, _ = w.Write([]byte(`{"message":"queued"}`))
				default:
					_, _ = w.Write([]byte(`{"uuid":"database-1","name":"postgres"}`))
				}
			}))
			defer server.Close()

			err := tt.call(NewService(api.NewClient(server.URL, "test-token")))
			require.NoError(t, err)
		})
	}
}

func stringPointer(value string) *string { return &value }

func TestService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	services, err := svc.List(context.Background())

	require.NoError(t, err)
	assert.Empty(t, services)
}

func TestService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	_, err := svc.List(context.Background())

	require.Error(t, err)
}

func TestService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
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
	svc := NewService(client)

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

func TestService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "service not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	_, err := svc.Get(context.Background(), "nonexistent")

	require.Error(t, err)
}

func TestService_Start(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/start", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Service starting request queued."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	resp, err := svc.Start(context.Background(), "service-uuid-123")

	require.NoError(t, err)
	assert.Equal(t, "Service starting request queued.", resp.Message)
}

func TestService_Start_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "service already running"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	_, err := svc.Start(context.Background(), "service-uuid-123")

	require.Error(t, err)
}

func TestService_Stop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/stop", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Service stopping request queued."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	resp, err := svc.Stop(context.Background(), "service-uuid-123")

	require.NoError(t, err)
	assert.Equal(t, "Service stopping request queued.", resp.Message)
}

func TestService_Stop_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "service already stopped"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	_, err := svc.Stop(context.Background(), "service-uuid-123")

	require.Error(t, err)
}

func TestService_Restart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/restart", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Service restarting request queued."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	resp, err := svc.Restart(context.Background(), "service-uuid-123")

	require.NoError(t, err)
	assert.Equal(t, "Service restarting request queued.", resp.Message)
}

func TestService_Restart_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "service not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	_, err := svc.Restart(context.Background(), "service-uuid-123")

	require.Error(t, err)
}

func TestService_ListEnvs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"uuid": "env-1",
				"key": "DATABASE_URL",
				"value": "postgres://localhost",
				"is_buildtime": false,
				"is_preview": false
			},
			{
				"uuid": "env-2",
				"key": "API_KEY",
				"value": "secret",
				"is_buildtime": true,
				"is_preview": false
			}
		]`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	envs, err := svc.ListEnvs(context.Background(), "service-uuid-123")

	require.NoError(t, err)
	assert.Len(t, envs, 2)
	assert.Equal(t, "DATABASE_URL", envs[0].Key)
	assert.Equal(t, "API_KEY", envs[1].Key)
}

func TestService_CreateEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"uuid": "env-new",
			"key": "NEW_VAR",
			"value": "new_value",
			"is_buildtime": false,
			"is_preview": false
		}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	env, err := svc.CreateEnv(context.Background(), "service-uuid-123", &models.ServiceEnvironmentVariableCreateRequest{
		Key:   "NEW_VAR",
		Value: "new_value",
	})

	require.NoError(t, err)
	assert.Equal(t, "NEW_VAR", env.Key)
	assert.Equal(t, "new_value", env.Value)
}

func TestService_UpdateEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"uuid": "env-123",
			"key": "UPDATED_VAR",
			"value": "updated_value",
			"is_buildtime": true,
			"is_preview": false,
			"comment": "updated comment"
		}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	newKey := "UPDATED_VAR"
	newComment := "updated comment"
	env, err := svc.UpdateEnv(context.Background(), "service-uuid-123", &models.ServiceEnvironmentVariableUpdateRequest{
		Key:     &newKey,
		Comment: &newComment,
	})

	require.NoError(t, err)
	assert.Equal(t, "UPDATED_VAR", env.Key)
	assert.Equal(t, "updated comment", *env.Comment)
}

func TestService_DeleteEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/service-uuid-123/envs/env-456", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	err := svc.DeleteEnv(context.Background(), "service-uuid-123", "env-456")

	require.NoError(t, err)
}

func TestService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"uuid": "service-new-uuid",
			"name": "WordPress",
			"status": "starting"
		}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	name := "My WordPress"
	service, err := svc.Create(context.Background(), &models.ServiceCreateRequest{
		Type:            "wordpress-with-mysql",
		ServerUUID:      "server-uuid",
		ProjectUUID:     "project-uuid",
		EnvironmentName: "production",
		Name:            &name,
	})

	require.NoError(t, err)
	assert.Equal(t, "service-new-uuid", service.UUID)
	assert.Equal(t, "WordPress", service.Name)
}

func TestService_Create_WithInstantDeploy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"uuid": "service-instant-uuid",
			"name": "Ghost Blog",
			"status": "running"
		}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	instantDeploy := true
	service, err := svc.Create(context.Background(), &models.ServiceCreateRequest{
		Type:            "ghost",
		ServerUUID:      "server-uuid",
		ProjectUUID:     "project-uuid",
		EnvironmentName: "production",
		InstantDeploy:   &instantDeploy,
	})

	require.NoError(t, err)
	assert.Equal(t, "service-instant-uuid", service.UUID)
	assert.Equal(t, "Ghost Blog", service.Name)
}

func TestService_Create_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "invalid service type"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	_, err := svc.Create(context.Background(), &models.ServiceCreateRequest{
		Type:            "invalid-type",
		ServerUUID:      "server-uuid",
		ProjectUUID:     "project-uuid",
		EnvironmentName: "production",
	})

	require.Error(t, err)
}

func TestService_ListStorages(t *testing.T) {
	hostPath := "/var/data"
	resp := models.StoragesResponse{
		PersistentStorages: []models.PersistentStorage{
			{
				ID:                     1,
				UUID:                   "ps-uuid-1",
				Name:                   "data-volume",
				MountPath:              "/data",
				HostPath:               &hostPath,
				IsPreviewSuffixEnabled: true,
			},
		},
		FileStorages: []models.FileStorage{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/svc-uuid-123/storages", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	result, err := svc.ListStorages(context.Background(), "svc-uuid-123")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "persistent", result[0].Type)
	assert.Equal(t, "data-volume", result[0].Name)
	assert.True(t, result[0].IsPreviewSuffixEnabled)
}

func TestService_CreateStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/svc-uuid-123/storages", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.ServiceStorageCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "persistent", req.Type)
		assert.Equal(t, "/data", req.MountPath)
		assert.Equal(t, "sub-resource-uuid", req.ResourceUUID)

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	name := "my-volume"
	req := &models.ServiceStorageCreateRequest{
		Type:         "persistent",
		MountPath:    "/data",
		ResourceUUID: "sub-resource-uuid",
		Name:         &name,
	}

	err := svc.CreateStorage(context.Background(), "svc-uuid-123", req)
	require.NoError(t, err)
}

func TestService_UpdateStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/svc-uuid-123/storages", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req models.StorageUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "persistent", req.Type)
		assert.NotNil(t, req.UUID)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	storageUUID := "storage-uuid-1"
	name := "new-name"
	req := &models.StorageUpdateRequest{
		UUID: &storageUUID,
		Type: "persistent",
		Name: &name,
	}

	err := svc.UpdateStorage(context.Background(), "svc-uuid-123", req)
	require.NoError(t, err)
}

func TestService_DeleteStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/svc-uuid-123/storages/storage-uuid-1", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"Storage deleted."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewService(client)

	err := svc.DeleteStorage(context.Background(), "svc-uuid-123", "storage-uuid-1")
	require.NoError(t, err)
}

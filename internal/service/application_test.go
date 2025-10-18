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

func TestApplicationService_List(t *testing.T) {
	desc1 := "App Description 1"
	desc2 := "App Description 2"
	branch1 := "main"
	branch2 := "develop"
	fqdn1 := "app1.example.com"
	fqdn2 := "app2.example.com"

	applications := []models.Application{
		{
			ID:          1,
			UUID:        "app-uuid-1",
			Name:        "Test App 1",
			Description: &desc1,
			Status:      "running",
			GitBranch:   &branch1,
			FQDN:        &fqdn1,
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-02T00:00:00Z",
		},
		{
			ID:          2,
			UUID:        "app-uuid-2",
			Name:        "Test App 2",
			Description: &desc2,
			Status:      "stopped",
			GitBranch:   &branch2,
			FQDN:        &fqdn2,
			CreatedAt:   "2024-01-03T00:00:00Z",
			UpdatedAt:   "2024-01-04T00:00:00Z",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(applications)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "app-uuid-1", result[0].UUID)
	assert.Equal(t, "Test App 1", result[0].Name)
	assert.Equal(t, "running", result[0].Status)
	assert.Equal(t, "main", *result[0].GitBranch)
	assert.Equal(t, "app-uuid-2", result[1].UUID)
	assert.Equal(t, "Test App 2", result[1].Name)
	assert.Equal(t, "stopped", result[1].Status)
}

func TestApplicationService_List_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]models.Application{})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestApplicationService_List_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal server error"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.List(context.Background())
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list applications")
}

func TestApplicationService_Get(t *testing.T) {
	desc := "Test Application"
	branch := "main"
	fqdn := "test.example.com"

	application := models.Application{
		ID:          1,
		UUID:        "app-uuid-123",
		Name:        "Test App",
		Description: &desc,
		Status:      "running",
		GitBranch:   &branch,
		FQDN:        &fqdn,
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-02T00:00:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(application)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Get(context.Background(), "app-uuid-123")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "app-uuid-123", result.UUID)
	assert.Equal(t, "Test App", result.Name)
	assert.Equal(t, "running", result.Status)
	assert.Equal(t, "main", *result.GitBranch)
	assert.Equal(t, "test.example.com", *result.FQDN)
}

func TestApplicationService_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"application not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Get(context.Background(), "non-existent-uuid")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get application")
}

func TestApplicationService_Get_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal server error"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Get(context.Background(), "app-uuid-123")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get application")
}

func TestApplicationService_Update(t *testing.T) {
	newName := "Updated App Name"
	newBranch := "develop"
	newDesc := "Updated description"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body
		var req models.ApplicationUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		assert.NotNil(t, req.Name)
		assert.Equal(t, newName, *req.Name)
		assert.NotNil(t, req.GitBranch)
		assert.Equal(t, newBranch, *req.GitBranch)

		// Return updated application
		updatedApp := models.Application{
			ID:          1,
			UUID:        "app-uuid-123",
			Name:        newName,
			Description: &newDesc,
			Status:      "running",
			GitBranch:   &newBranch,
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-05T00:00:00Z",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updatedApp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	req := models.ApplicationUpdateRequest{
		Name:      &newName,
		GitBranch: &newBranch,
	}

	result, err := svc.Update(context.Background(), "app-uuid-123", req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "app-uuid-123", result.UUID)
	assert.Equal(t, newName, result.Name)
	assert.Equal(t, newBranch, *result.GitBranch)
}

func TestApplicationService_Update_PartialUpdate(t *testing.T) {
	newDomains := "app.example.com,www.example.com"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-456", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		// Verify only domains field is in request
		var req models.ApplicationUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		assert.Nil(t, req.Name)
		assert.Nil(t, req.GitBranch)
		assert.NotNil(t, req.Domains)
		assert.Equal(t, newDomains, *req.Domains)

		// Return updated application
		fqdn := "app.example.com"
		updatedApp := models.Application{
			ID:        2,
			UUID:      "app-uuid-456",
			Name:      "Existing App",
			Status:    "running",
			FQDN:      &fqdn,
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-05T00:00:00Z",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updatedApp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	req := models.ApplicationUpdateRequest{
		Domains: &newDomains,
	}

	result, err := svc.Update(context.Background(), "app-uuid-456", req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "app-uuid-456", result.UUID)
	assert.Equal(t, "app.example.com", *result.FQDN)
}

func TestApplicationService_Update_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"application not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	newName := "Updated Name"
	req := models.ApplicationUpdateRequest{
		Name: &newName,
	}

	result, err := svc.Update(context.Background(), "non-existent-uuid", req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update application")
}

func TestApplicationService_Update_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal server error"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	newName := "Updated Name"
	req := models.ApplicationUpdateRequest{
		Name: &newName,
	}

	result, err := svc.Update(context.Background(), "app-uuid-123", req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update application")
}

func TestApplicationService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	err := svc.Delete(context.Background(), "app-uuid-123")
	require.NoError(t, err)
}

func TestApplicationService_Delete_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"application not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	err := svc.Delete(context.Background(), "non-existent-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete application")
}

func TestApplicationService_Delete_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal server error"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	err := svc.Delete(context.Background(), "app-uuid-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete application")
}

func TestApplicationService_Start(t *testing.T) {
	deploymentUUID := "deploy-uuid-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/start", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		resp := models.ApplicationLifecycleResponse{
			Message:        "Deployment request queued.",
			DeploymentUUID: &deploymentUUID,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Start(context.Background(), "app-uuid-123", false, false)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Deployment request queued.", result.Message)
	assert.NotNil(t, result.DeploymentUUID)
	assert.Equal(t, "deploy-uuid-123", *result.DeploymentUUID)
}

func TestApplicationService_Start_WithForce(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/start", r.URL.Path)
		assert.Equal(t, "force=true", r.URL.RawQuery)

		resp := models.ApplicationLifecycleResponse{
			Message: "Deployment request queued.",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Start(context.Background(), "app-uuid-123", true, false)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestApplicationService_Start_WithInstantDeploy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/start", r.URL.Path)
		assert.Equal(t, "instant_deploy=true", r.URL.RawQuery)

		resp := models.ApplicationLifecycleResponse{
			Message: "Deployment request queued.",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Start(context.Background(), "app-uuid-123", false, true)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestApplicationService_Start_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"failed to start application"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Start(context.Background(), "app-uuid-123", false, false)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to start application")
}

func TestApplicationService_Stop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/stop", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		resp := models.ApplicationLifecycleResponse{
			Message: "Application stopped successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Stop(context.Background(), "app-uuid-123")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Application stopped successfully", result.Message)
}

func TestApplicationService_Stop_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"failed to stop application"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Stop(context.Background(), "app-uuid-123")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to stop application")
}

func TestApplicationService_Restart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/restart", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		resp := models.ApplicationLifecycleResponse{
			Message: "Application restarted successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Restart(context.Background(), "app-uuid-123")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Application restarted successfully", result.Message)
}

func TestApplicationService_Restart_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"failed to restart application"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Restart(context.Background(), "app-uuid-123")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to restart application")
}

func TestApplicationService_Logs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/logs", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		resp := models.ApplicationLogsResponse{
			Logs: "[2025-10-15 12:00:00] Application started\n[2025-10-15 12:00:01] Server listening on port 3000\n",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Logs(context.Background(), "app-uuid-123", 0)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result.Logs, "Application started")
}

func TestApplicationService_Logs_WithLines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/logs", r.URL.Path)
		assert.Equal(t, "lines=50", r.URL.RawQuery)
		assert.Equal(t, "GET", r.Method)

		resp := models.ApplicationLogsResponse{
			Logs: "[2025-10-15 12:00:00] Log line\n",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Logs(context.Background(), "app-uuid-123", 50)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestApplicationService_Logs_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"application not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Logs(context.Background(), "app-uuid-123", 0)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get logs for application")
}

func TestApplicationService_ListEnvs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		envs := []models.EnvironmentVariable{
			{
				ID:          1,
				UUID:        "env-uuid-1",
				Key:         "DATABASE_URL",
				Value:       "********",
				IsBuildTime: false,
				IsPreview:   false,
			},
			{
				ID:          2,
				UUID:        "env-uuid-2",
				Key:         "API_KEY",
				Value:       "********",
				IsBuildTime: true,
				IsPreview:   false,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(envs)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.ListEnvs(context.Background(), "app-uuid-123")
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "DATABASE_URL", result[0].Key)
	assert.Equal(t, "API_KEY", result[1].Key)
}

func TestApplicationService_ListEnvs_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/envs", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]models.EnvironmentVariable{})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.ListEnvs(context.Background(), "app-uuid-123")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestApplicationService_ListEnvs_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"application not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.ListEnvs(context.Background(), "app-uuid-123")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list environment variables")
}

func TestApplicationService_CreateEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		env := models.EnvironmentVariable{
			ID:          1,
			UUID:        "env-uuid-1",
			Key:         "API_KEY",
			Value:       "secret123",
			IsBuildTime: false,
			IsPreview:   false,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(env)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	isBuildTime := false
	req := &models.EnvironmentVariableCreateRequest{
		Key:         "API_KEY",
		Value:       "secret123",
		IsBuildTime: &isBuildTime,
	}

	result, err := svc.CreateEnv(context.Background(), "app-uuid-123", req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "API_KEY", result.Key)
	assert.Equal(t, "env-uuid-1", result.UUID)
}

func TestApplicationService_CreateEnv_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"key already exists"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	req := &models.EnvironmentVariableCreateRequest{
		Key:   "API_KEY",
		Value: "secret123",
	}

	result, err := svc.CreateEnv(context.Background(), "app-uuid-123", req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create environment variable")
}

func TestApplicationService_UpdateEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		env := models.EnvironmentVariable{
			UUID:  "env-uuid-1",
			Key:   "API_KEY",
			Value: "newsecret456",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(env)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	newValue := "newsecret456"
	req := &models.EnvironmentVariableUpdateRequest{
		UUID:  "env-uuid-1",
		Value: &newValue,
	}

	result, err := svc.UpdateEnv(context.Background(), "app-uuid-123", req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "API_KEY", result.Key)
	assert.Equal(t, "newsecret456", result.Value)
}

func TestApplicationService_UpdateEnv_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"environment variable not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	newValue := "newsecret456"
	req := &models.EnvironmentVariableUpdateRequest{
		UUID:  "env-uuid-1",
		Value: &newValue,
	}

	result, err := svc.UpdateEnv(context.Background(), "app-uuid-123", req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update environment variable")
}

func TestApplicationService_DeleteEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/envs/env-uuid-1", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	err := svc.DeleteEnv(context.Background(), "app-uuid-123", "env-uuid-1")
	require.NoError(t, err)
}

func TestApplicationService_DeleteEnv_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"environment variable not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	err := svc.DeleteEnv(context.Background(), "app-uuid-123", "env-uuid-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete environment variable")
}

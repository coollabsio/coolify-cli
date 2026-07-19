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
	repo := "https://github.com/example/app"
	buildPack := "nixpacks"
	ports := "3000"
	installCmd := "npm install"
	buildCmd := "npm run build"
	startCmd := "npm start"
	healthPath := "/health"
	limitsCPUs := "2"
	limitsMemory := "512M"
	preDeployCmd := "npm run migrate"
	isAutoDeployEnabled := true

	application := models.Application{
		ID:                   1,
		UUID:                 "app-uuid-123",
		Name:                 "Test App",
		Description:          &desc,
		Status:               "running",
		GitBranch:            &branch,
		FQDN:                 &fqdn,
		GitRepository:        &repo,
		BuildPack:            &buildPack,
		PortsExposes:         &ports,
		InstallCommand:       &installCmd,
		BuildCommand:         &buildCmd,
		StartCommand:         &startCmd,
		HealthCheckPath:      &healthPath,
		LimitsCPUs:           &limitsCPUs,
		LimitsMemory:         &limitsMemory,
		PreDeploymentCommand: &preDeployCmd,
		Settings: &models.ApplicationSettings{
			IsAutoDeployEnabled: &isAutoDeployEnabled,
		},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-02T00:00:00Z",
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
	assert.Equal(t, "https://github.com/example/app", *result.GitRepository)
	assert.Equal(t, "nixpacks", *result.BuildPack)
	assert.Equal(t, "3000", *result.PortsExposes)
	assert.Equal(t, "npm install", *result.InstallCommand)
	assert.Equal(t, "npm run build", *result.BuildCommand)
	assert.Equal(t, "npm start", *result.StartCommand)
	assert.Equal(t, "/health", *result.HealthCheckPath)
	assert.Equal(t, "2", *result.LimitsCPUs)
	assert.Equal(t, "512M", *result.LimitsMemory)
	assert.Equal(t, "npm run migrate", *result.PreDeploymentCommand)
	require.NotNil(t, result.Settings)
	assert.True(t, *result.Settings.IsAutoDeployEnabled)
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

func TestApplicationService_DeletePreview_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/previews/42", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"Preview deletion request queued."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	err := svc.DeletePreview(context.Background(), "app-uuid-123", "42")
	require.NoError(t, err)
}

func TestApplicationService_DeletePreview_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Preview not found."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	err := svc.DeletePreview(context.Background(), "app-uuid-123", "999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete preview")
}

func TestApplicationService_DeletePreview_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal server error"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	err := svc.DeletePreview(context.Background(), "app-uuid-123", "42")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete preview")
}

func TestApplicationService_Start(t *testing.T) {
	deploymentUUID := "deploy-uuid-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/start", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
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
		assert.Equal(t, http.MethodPost, r.Method)
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
		assert.Equal(t, http.MethodPost, r.Method)
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

	result, err := svc.Logs(context.Background(), "app-uuid-123", 0, false)
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

	result, err := svc.Logs(context.Background(), "app-uuid-123", 50, false)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestApplicationService_Logs_WithLinesAndTimestamps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/logs", r.URL.Path)
		assert.Equal(t, "25", r.URL.Query().Get("lines"))
		assert.Equal(t, "true", r.URL.Query().Get("show_timestamps"))
		_ = json.NewEncoder(w).Encode(models.ApplicationLogsResponse{Logs: "timestamped"})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	result, err := NewApplicationService(client).Logs(context.Background(), "app-uuid-123", 25, true)

	require.NoError(t, err)
	assert.Equal(t, "timestamped", result.Logs)
}

func TestApplicationService_Move(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/move", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		var req models.ApplicationMoveRequest
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "environment-uuid-456", req.EnvironmentUUID)
		_ = json.NewEncoder(w).Encode(models.ApplicationMoveResponse{
			Message:         "Application moved successfully.",
			UUID:            "app-uuid-123",
			ProjectUUID:     "project-uuid-789",
			EnvironmentUUID: "environment-uuid-456",
		})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	result, err := NewApplicationService(client).Move(context.Background(), "app-uuid-123", models.ApplicationMoveRequest{
		EnvironmentUUID: "environment-uuid-456",
	})

	require.NoError(t, err)
	assert.Equal(t, "project-uuid-789", result.ProjectUUID)
	assert.Equal(t, "environment-uuid-456", result.EnvironmentUUID)
}

func TestApplicationService_Logs_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"application not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.Logs(context.Background(), "app-uuid-123", 0, false)
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

		comment := "API key for external service"
		env := models.EnvironmentVariable{
			UUID:    "env-uuid-1",
			Key:     "API_KEY",
			Value:   "newsecret456",
			Comment: &comment,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(env)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	newKey := "API_KEY"
	newValue := "newsecret456"
	newComment := "API key for external service"
	req := &models.EnvironmentVariableUpdateRequest{
		Key:     &newKey,
		Value:   &newValue,
		Comment: &newComment,
	}

	result, err := svc.UpdateEnv(context.Background(), "app-uuid-123", req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "API_KEY", result.Key)
	assert.Equal(t, "newsecret456", result.Value)
	assert.Equal(t, "API key for external service", *result.Comment)
}

func TestApplicationService_UpdateEnv_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"environment variable not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	newKey := "API_KEY"
	newValue := "newsecret456"
	req := &models.EnvironmentVariableUpdateRequest{
		Key:   &newKey,
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

func TestApplicationService_ListEnvs_AllFields(t *testing.T) {
	// Test that all fields including is_buildtime (without underscore), is_runtime, and is_shared
	// are correctly parsed from API response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Mock API response with all fields
		envs := []models.EnvironmentVariable{
			{
				UUID:           "env-1",
				Key:            "DATABASE_URL",
				Value:          "postgres://localhost",
				IsBuildTime:    true,
				IsPreview:      false,
				IsLiteralValue: true,
				IsShownOnce:    false,
				IsRuntime:      true,
				IsShared:       false,
			},
			{
				UUID:           "env-2",
				Key:            "API_KEY",
				Value:          "secret",
				IsBuildTime:    false,
				IsPreview:      true,
				IsLiteralValue: false,
				IsShownOnce:    false,
				IsRuntime:      false,
				IsShared:       true,
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

	// Verify first env var
	assert.Equal(t, "DATABASE_URL", result[0].Key)
	assert.True(t, result[0].IsBuildTime, "IsBuildTime should be true for DATABASE_URL")
	assert.True(t, result[0].IsRuntime, "IsRuntime should be true for DATABASE_URL")
	assert.False(t, result[0].IsShared, "IsShared should be false for DATABASE_URL")
	assert.True(t, result[0].IsLiteralValue)
	assert.False(t, result[0].IsPreview)

	// Verify second env var
	assert.Equal(t, "API_KEY", result[1].Key)
	assert.False(t, result[1].IsBuildTime, "IsBuildTime should be false for API_KEY")
	assert.False(t, result[1].IsRuntime, "IsRuntime should be false for API_KEY")
	assert.True(t, result[1].IsShared, "IsShared should be true for API_KEY")
	assert.False(t, result[1].IsLiteralValue)
	assert.True(t, result[1].IsPreview)
}

func TestApplicationService_EnvBuildtimeFlag(t *testing.T) {
	// Test specifically that is_buildtime (without underscore) unmarshals correctly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/envs", r.URL.Path)

		// Directly write JSON with is_buildtime (no underscore) to mimic actual API
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"uuid": "env-test-1",
				"key": "BUILD_VAR",
				"value": "build_value",
				"is_buildtime": true,
				"is_preview": false,
				"is_literal": false,
				"is_shown_once": false,
				"is_runtime": true,
				"is_shared": false
			}
		]`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.ListEnvs(context.Background(), "app-uuid-123")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "BUILD_VAR", result[0].Key)
	assert.True(t, result[0].IsBuildTime, "is_buildtime field should unmarshal correctly to true")
	assert.True(t, result[0].IsRuntime, "is_runtime field should unmarshal correctly to true")
}

func TestApplicationService_EnvRuntimeAndShared(t *testing.T) {
	// Test is_runtime and is_shared fields specifically
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/envs", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"uuid": "env-runtime",
				"key": "RUNTIME_VAR",
				"value": "runtime_value",
				"is_buildtime": false,
				"is_preview": false,
				"is_literal": false,
				"is_shown_once": false,
				"is_runtime": true,
				"is_shared": false
			},
			{
				"uuid": "env-shared",
				"key": "SHARED_VAR",
				"value": "shared_value",
				"is_buildtime": false,
				"is_preview": false,
				"is_literal": false,
				"is_shown_once": false,
				"is_runtime": false,
				"is_shared": true
			}
		]`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.ListEnvs(context.Background(), "app-uuid-123")
	require.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify runtime var
	assert.Equal(t, "RUNTIME_VAR", result[0].Key)
	assert.True(t, result[0].IsRuntime, "IsRuntime should be true")
	assert.False(t, result[0].IsShared, "IsShared should be false")

	// Verify shared var
	assert.Equal(t, "SHARED_VAR", result[1].Key)
	assert.False(t, result[1].IsRuntime, "IsRuntime should be false")
	assert.True(t, result[1].IsShared, "IsShared should be true")
}

func TestApplicationService_CreatePublic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/public", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var req models.ApplicationCreatePublicRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "proj-uuid", req.ProjectUUID)
		assert.Equal(t, "server-uuid", req.ServerUUID)
		assert.Equal(t, "https://github.com/user/repo", req.GitRepository)
		assert.Equal(t, "main", req.GitBranch)
		assert.Equal(t, "nixpacks", req.BuildPack)
		assert.Equal(t, "3000", req.PortsExposes)

		branch := "main"
		fqdn := "app.example.com"
		app := models.Application{
			UUID:      "new-app-uuid",
			Name:      "My App",
			Status:    "starting",
			GitBranch: &branch,
			FQDN:      &fqdn,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	envName := "production"
	req := &models.ApplicationCreatePublicRequest{
		ProjectUUID:     "proj-uuid",
		ServerUUID:      "server-uuid",
		GitRepository:   "https://github.com/user/repo",
		GitBranch:       "main",
		BuildPack:       "nixpacks",
		PortsExposes:    "3000",
		EnvironmentName: &envName,
	}

	result, err := svc.CreatePublic(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new-app-uuid", result.UUID)
	assert.Equal(t, "My App", result.Name)
}

func TestApplicationService_CreatePublic_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"invalid repository URL"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	envName := "production"
	req := &models.ApplicationCreatePublicRequest{
		ProjectUUID:     "proj-uuid",
		ServerUUID:      "server-uuid",
		GitRepository:   "invalid-repo",
		GitBranch:       "main",
		BuildPack:       "nixpacks",
		PortsExposes:    "3000",
		EnvironmentName: &envName,
	}

	result, err := svc.CreatePublic(context.Background(), req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create application from public repository")
}

func TestApplicationService_CreateGitHubApp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/private-github-app", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.ApplicationCreateGitHubAppRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "github-app-uuid", req.GitHubAppUUID)
		assert.Equal(t, "owner/repo", req.GitRepository)

		branch := "main"
		app := models.Application{
			UUID:      "new-app-uuid",
			Name:      "Private App",
			Status:    "starting",
			GitBranch: &branch,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	envName := "production"
	req := &models.ApplicationCreateGitHubAppRequest{
		ProjectUUID:     "proj-uuid",
		ServerUUID:      "server-uuid",
		GitHubAppUUID:   "github-app-uuid",
		GitRepository:   "owner/repo",
		GitBranch:       "main",
		BuildPack:       "nixpacks",
		PortsExposes:    "3000",
		EnvironmentName: &envName,
	}

	result, err := svc.CreateGitHubApp(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new-app-uuid", result.UUID)
}

func TestApplicationService_CreateDeployKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/private-deploy-key", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.ApplicationCreateDeployKeyRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "key-uuid", req.PrivateKeyUUID)
		assert.Equal(t, "git@github.com:owner/repo.git", req.GitRepository)

		branch := "main"
		app := models.Application{
			UUID:      "new-app-uuid",
			Name:      "Deploy Key App",
			Status:    "starting",
			GitBranch: &branch,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	envName := "production"
	req := &models.ApplicationCreateDeployKeyRequest{
		ProjectUUID:     "proj-uuid",
		ServerUUID:      "server-uuid",
		PrivateKeyUUID:  "key-uuid",
		GitRepository:   "git@github.com:owner/repo.git",
		GitBranch:       "main",
		BuildPack:       "nixpacks",
		PortsExposes:    "3000",
		EnvironmentName: &envName,
	}

	result, err := svc.CreateDeployKey(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new-app-uuid", result.UUID)
}

func TestApplicationService_CreateDockerfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/dockerfile", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.ApplicationCreateDockerfileRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Contains(t, req.Dockerfile, "FROM node:18")

		app := models.Application{
			UUID:   "new-app-uuid",
			Name:   "Dockerfile App",
			Status: "starting",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	envName := "production"
	req := &models.ApplicationCreateDockerfileRequest{
		ProjectUUID:     "proj-uuid",
		ServerUUID:      "server-uuid",
		Dockerfile:      "FROM node:18\nCOPY . .\nCMD [\"node\", \"app.js\"]",
		EnvironmentName: &envName,
	}

	result, err := svc.CreateDockerfile(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new-app-uuid", result.UUID)
}

func TestApplicationService_CreateDockerImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/dockerimage", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.ApplicationCreateDockerImageRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "nginx:latest", req.DockerRegistryImageName)
		assert.Equal(t, "80", req.PortsExposes)

		app := models.Application{
			UUID:   "new-app-uuid",
			Name:   "Docker Image App",
			Status: "starting",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(app)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	envName := "production"
	req := &models.ApplicationCreateDockerImageRequest{
		ProjectUUID:             "proj-uuid",
		ServerUUID:              "server-uuid",
		DockerRegistryImageName: "nginx:latest",
		PortsExposes:            "80",
		EnvironmentName:         &envName,
	}

	result, err := svc.CreateDockerImage(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new-app-uuid", result.UUID)
}

func TestApplicationService_BulkUpdateEnvs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/envs/bulk", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Verify request body was correctly serialized
		var reqBody BulkUpdateEnvsRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Len(t, reqBody.Data, 1)
		assert.Equal(t, "KEY1", reqBody.Data[0].Key)
		assert.Equal(t, "VAL1", reqBody.Data[0].Value)

		// Return array response as per API
		w.Header().Set("Content-Type", "application/json")
		resp := []models.EnvironmentVariable{
			{
				Key:   "KEY1",
				Value: "VAL1",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	req := &BulkUpdateEnvsRequest{
		Data: []models.EnvironmentVariableCreateRequest{
			{Key: "KEY1", Value: "VAL1"},
		},
	}

	result, err := svc.BulkUpdateEnvs(context.Background(), "app-uuid-123", req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, *result, 1)
	assert.Equal(t, "KEY1", (*result)[0].Key)
}

func TestApplicationService_BulkUpdateEnvs_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal server error"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	req := &BulkUpdateEnvsRequest{
		Data: []models.EnvironmentVariableCreateRequest{
			{Key: "KEY1", Value: "VAL1"},
		},
	}

	result, err := svc.BulkUpdateEnvs(context.Background(), "app-uuid-123", req)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to bulk update environment variables")
}

func TestApplicationService_ListStorages(t *testing.T) {
	hostPath := "/var/data"
	content := "key: value"
	resp := models.StoragesResponse{
		PersistentStorages: []models.PersistentStorage{
			{
				ID:                     1,
				UUID:                   "ps-uuid-1",
				Name:                   "data-volume",
				MountPath:              "/data",
				HostPath:               &hostPath,
				IsPreviewSuffixEnabled: false,
				IsReadOnly:             false,
				ResourceType:           "App\\Models\\Application",
				ResourceID:             10,
			},
		},
		FileStorages: []models.FileStorage{
			{
				ID:                     2,
				UUID:                   "fs-uuid-1",
				FsPath:                 "/app/config.yml",
				MountPath:              "/app/config.yml",
				Content:                &content,
				IsDirectory:            false,
				IsBasedOnGit:           false,
				IsPreviewSuffixEnabled: true,
				ResourceType:           "App\\Models\\Application",
				ResourceID:             10,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/storages", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.ListStorages(context.Background(), "app-uuid-123")
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "ps-uuid-1", result[0].UUID)
	assert.Equal(t, "persistent", result[0].Type)
	assert.Equal(t, "data-volume", result[0].Name)
	assert.Equal(t, "/data", result[0].MountPath)
	assert.Equal(t, "/var/data", result[0].HostPath)
	assert.False(t, result[0].IsPreviewSuffixEnabled)
	assert.Equal(t, "fs-uuid-1", result[1].UUID)
	assert.Equal(t, "file", result[1].Type)
	assert.Equal(t, "/app/config.yml", result[1].Name)
	assert.Equal(t, "key: value", result[1].Content)
	assert.True(t, result[1].IsPreviewSuffixEnabled)
}

func TestApplicationService_ListStorages_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/storages", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"persistent_storages":[],"file_storages":[]}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.ListStorages(context.Background(), "app-uuid-123")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestApplicationService_ListStorages_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"application not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	result, err := svc.ListStorages(context.Background(), "app-uuid-123")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list storages")
}

func TestApplicationService_UpdateStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/storages", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var req models.StorageUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.NotNil(t, req.UUID)
		assert.Equal(t, "storage-uuid-1", *req.UUID)
		assert.Equal(t, "persistent", req.Type)
		assert.NotNil(t, req.Name)
		assert.Equal(t, "new-name", *req.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"Storage updated."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	name := "new-name"
	storageUUID := "storage-uuid-1"
	req := &models.StorageUpdateRequest{
		UUID: &storageUUID,
		Type: "persistent",
		Name: &name,
	}

	err := svc.UpdateStorage(context.Background(), "app-uuid-123", req)
	require.NoError(t, err)
}

func TestApplicationService_UpdateStorage_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"application not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	name := "new-name"
	storageUUID := "storage-uuid-1"
	req := &models.StorageUpdateRequest{
		UUID: &storageUUID,
		Type: "persistent",
		Name: &name,
	}

	err := svc.UpdateStorage(context.Background(), "non-existent-uuid", req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update storage")
}

func TestApplicationService_UpdateStorage_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal server error"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	name := "new-name"
	storageUUID := "storage-uuid-1"
	req := &models.StorageUpdateRequest{
		UUID: &storageUUID,
		Type: "persistent",
		Name: &name,
	}

	err := svc.UpdateStorage(context.Background(), "app-uuid-123", req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update storage")
}

func TestApplicationService_CreateStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/storages", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		var req models.StorageCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "persistent", req.Type)
		assert.Equal(t, "/data", req.MountPath)
		assert.NotNil(t, req.Name)
		assert.Equal(t, "my-volume", *req.Name)

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	name := "my-volume"
	req := &models.StorageCreateRequest{
		Type:      "persistent",
		MountPath: "/data",
		Name:      &name,
	}

	err := svc.CreateStorage(context.Background(), "app-uuid-123", req)
	require.NoError(t, err)
}

func TestApplicationService_CreateStorage_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"invalid request"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	req := &models.StorageCreateRequest{
		Type:      "persistent",
		MountPath: "/data",
	}

	err := svc.CreateStorage(context.Background(), "app-uuid-123", req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create storage")
}

func TestApplicationService_DeleteStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid-123/storages/storage-uuid-1", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"Storage deleted."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	err := svc.DeleteStorage(context.Background(), "app-uuid-123", "storage-uuid-1")
	require.NoError(t, err)
}

func TestApplicationService_DeleteStorage_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"storage not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewApplicationService(client)

	err := svc.DeleteStorage(context.Background(), "app-uuid-123", "nonexistent-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete storage")
}

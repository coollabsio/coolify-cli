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

func TestProjectService_List(t *testing.T) {
	desc1 := "Description 1"
	desc2 := "Description 2"
	projects := []models.Project{
		{
			UUID:        "proj-1",
			Name:        "Test Project 1",
			Description: &desc1,
		},
		{
			UUID:        "proj-2",
			Name:        "Test Project 2",
			Description: &desc2,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(projects)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	result, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "proj-1", result[0].UUID)
	assert.Equal(t, "Test Project 1", result[0].Name)
}

func TestProjectService_Get(t *testing.T) {
	desc := "Test Description"
	project := models.Project{
		UUID:        "proj-1",
		Name:        "Test Project",
		Description: &desc,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects/proj-1", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(project)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	result, err := svc.Get(context.Background(), "proj-1")
	require.NoError(t, err)
	assert.Equal(t, "proj-1", result.UUID)
	assert.Equal(t, "Test Project", result.Name)
}

func TestProjectService_Create(t *testing.T) {
	desc := "New Project Description"
	project := models.Project{
		UUID:        "proj-new",
		Name:        "New Project",
		Description: &desc,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.ProjectCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "New Project", req.Name)
		assert.NotNil(t, req.Description)
		assert.Equal(t, "New Project Description", *req.Description)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(project)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	result, err := svc.Create(context.Background(), &models.ProjectCreateRequest{
		Name:        "New Project",
		Description: &desc,
	})
	require.NoError(t, err)
	assert.Equal(t, "proj-new", result.UUID)
	assert.Equal(t, "New Project", result.Name)
}

func TestProjectService_Create_NameOnly(t *testing.T) {
	project := models.Project{
		UUID: "proj-minimal",
		Name: "Minimal Project",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.ProjectCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "Minimal Project", req.Name)
		assert.Nil(t, req.Description)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(project)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	result, err := svc.Create(context.Background(), &models.ProjectCreateRequest{
		Name: "Minimal Project",
	})
	require.NoError(t, err)
	assert.Equal(t, "proj-minimal", result.UUID)
	assert.Equal(t, "Minimal Project", result.Name)
}

func TestProjectService_Update(t *testing.T) {
	desc := "Updated description"
	project := models.Project{
		UUID:        "proj-1",
		Name:        "Updated Name",
		Description: &desc,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects/proj-1", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req models.ProjectUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.NotNil(t, req.Name)
		assert.Equal(t, "Updated Name", *req.Name)
		assert.NotNil(t, req.Description)
		assert.Equal(t, "Updated description", *req.Description)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(project)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	name := "Updated Name"
	result, err := svc.Update(context.Background(), "proj-1", models.ProjectUpdateRequest{
		Name:        &name,
		Description: &desc,
	})
	require.NoError(t, err)
	assert.Equal(t, "proj-1", result.UUID)
	assert.Equal(t, "Updated Name", result.Name)
	assert.Equal(t, "Updated description", *result.Description)
}

func TestProjectService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects/proj-1", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Project deleted."})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	err := svc.Delete(context.Background(), "proj-1")
	require.NoError(t, err)
}

func TestProjectService_ListEnvironments(t *testing.T) {
	environments := []models.Environment{
		{UUID: "env-1", Name: "production"},
		{UUID: "env-2", Name: "staging"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects/proj-1/environments", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(environments)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	result, err := svc.ListEnvironments(context.Background(), "proj-1")
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "env-1", result[0].UUID)
	assert.Equal(t, "production", result[0].Name)
	assert.Equal(t, "staging", result[1].Name)
}

func TestProjectService_CreateEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects/proj-1/environments", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.EnvironmentCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "staging", req.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(models.UUID{UUID: "env-new"})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	result, err := svc.CreateEnvironment(context.Background(), "proj-1", models.EnvironmentCreateRequest{
		Name: "staging",
	})
	require.NoError(t, err)
	assert.Equal(t, "env-new", result.UUID)
}

func TestProjectService_UpdateEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects/proj-1/environments/staging", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req models.EnvironmentUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.NotNil(t, req.Name)
		assert.Equal(t, "production", *req.Name)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"uuid":        "env-1",
			"name":        "production",
			"description": "Prod env",
		})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	name := "production"
	result, err := svc.UpdateEnvironment(context.Background(), "proj-1", "staging", models.EnvironmentUpdateRequest{
		Name: &name,
	})
	require.NoError(t, err)
	assert.Equal(t, "env-1", result["uuid"])
	assert.Equal(t, "production", result["name"])
}

func TestProjectService_UpdateEnvironment_PathEscape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects/proj-1/environments/my%20env", r.URL.EscapedPath())
		assert.Equal(t, "PATCH", r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"uuid": "env-1", "name": "my env"})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	name := "my env"
	_, err := svc.UpdateEnvironment(context.Background(), "proj-1", "my env", models.EnvironmentUpdateRequest{
		Name: &name,
	})
	require.NoError(t, err)
}

func TestProjectService_DeleteEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects/proj-1/environments/staging", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Environment deleted."})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	err := svc.DeleteEnvironment(context.Background(), "proj-1", "staging")
	require.NoError(t, err)
}

func TestProjectService_DeleteEnvironment_ByUUID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/projects/proj-1/environments/env-uuid-1", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewProjectService(client)

	err := svc.DeleteEnvironment(context.Background(), "proj-1", "env-uuid-1")
	require.NoError(t, err)
}

func TestProjectService_GetEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		// Environment names may contain spaces; they must be path-escaped.
		assert.Equal(t, "/api/v1/projects/proj-uuid/my%20env", r.URL.EscapedPath())
		_, _ = w.Write([]byte(`{
			"id": 3, "uuid": "env-uuid", "name": "my env", "project_id": 1, "description": "Staging",
			"applications": [{"id": 10, "uuid": "app-1", "name": "web", "status": "running:healthy", "fqdn": "https://x"}],
			"postgresqls": [{"id": 11, "uuid": "pg-1", "name": "db", "status": "exited"}],
			"redis": [], "mongodbs": [], "mysqls": [], "mariadbs": [],
			"services": [{"id": 12, "uuid": "svc-1", "name": "plausible"}]
		}`))
	}))
	defer server.Close()

	svc := NewProjectService(api.NewClient(server.URL, "test-token"))
	env, err := svc.GetEnvironment(context.Background(), "proj-uuid", "my env")
	require.NoError(t, err)
	assert.Equal(t, "env-uuid", env.UUID)
	assert.Equal(t, "my env", env.Name)
	require.NotNil(t, env.Description)
	assert.Equal(t, "Staging", *env.Description)
	require.Len(t, env.Applications, 1)
	assert.Equal(t, 10, env.Applications[0].ID)

	rows := env.ResourceRows()
	assert.Equal(t, []models.EnvironmentResourceRow{
		{Type: "application", UUID: "app-1", Name: "web", Status: "running:healthy"},
		{Type: "postgresql", UUID: "pg-1", Name: "db", Status: "exited"},
		{Type: "service", UUID: "svc-1", Name: "plausible"},
	}, rows)
}

func TestProjectService_GetEnvironment_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Environment not found."}`))
	}))
	defer server.Close()

	svc := NewProjectService(api.NewClient(server.URL, "test-token"))
	_, err := svc.GetEnvironment(context.Background(), "proj-uuid", "nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Environment not found.")
}

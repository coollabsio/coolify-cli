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

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

func TestResourceService_List(t *testing.T) {
	resources := []models.Resource{
		{
			UUID: "res-1",
			Name: "Test Resource 1",
			Type: "application",
		},
		{
			UUID: "res-2",
			Name: "Test Resource 2",
			Type: "database",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/resources", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resources)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewResourceService(client)

	result, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "res-1", result[0].UUID)
	assert.Equal(t, "Test Resource 1", result[0].Name)
}

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

func TestDomainService_List(t *testing.T) {
	domains := []models.Domain{
		{
			IP:      "192.168.1.1",
			Domains: []string{"example.com", "www.example.com"},
		},
		{
			IP:      "192.168.1.2",
			Domains: []string{"test.com"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/domains", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(domains)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDomainService(client)

	result, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "192.168.1.1", result[0].IP)
	assert.Equal(t, []string{"example.com", "www.example.com"}, result[0].Domains)
}

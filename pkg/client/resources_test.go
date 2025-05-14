package client_test

import (
	"github.com/coollabsio/coolify-cli/pkg/client"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestClient_ListResources(t *testing.T) {
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/api/v1/resources", func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := os.ReadFile("resources_test.json")
		w.Write(bytes)
	})
	server := httptest.NewServer(mockMux)
	defer server.Close()

	c := client.New(server.URL, "my-token")

	resources, _ := c.ListResources()

	assert.Len(t, resources, 1)
	assert.Equal(t, "running:healthy", resources[0].Status)
	assert.Equal(t, 2, resources[0].ID)
	assert.Equal(t, "coollabsio/coolify-examples:v4.x-zxczxcxzc", resources[0].Name)
	assert.Equal(t, "application", resources[0].Type)
	assert.Equal(t, "abc123", resources[0].Uuid)
}

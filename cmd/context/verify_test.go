package context

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
)

// TestVerifyCommand_APIIntegration tests the verify logic using the API client directly
// This tests the core functionality that the verify command relies on
func TestVerifyCommand_APIIntegration(t *testing.T) {
	t.Run("successful verification", func(t *testing.T) {
		// Create a test HTTP server that responds to /api/v1/version
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v1/version", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("4.0.0-beta.383"))
		}))
		defer server.Close()

		// Create API client and verify connection
		client := api.NewClient(server.URL, "test-token")
		version, err := client.GetVersion(context.Background())

		// Verify results
		require.NoError(t, err)
		assert.Equal(t, "4.0.0-beta.383", version)
	})

	t.Run("unauthorized - invalid token", func(t *testing.T) {
		// Create a test HTTP server that returns 401
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid token",
			})
		}))
		defer server.Close()

		// Create API client with invalid token
		client := api.NewClient(server.URL, "invalid-token")
		_, err := client.GetVersion(context.Background())

		// Verify error
		require.Error(t, err)
		assert.True(t, api.IsUnauthorized(err))
	})

	t.Run("server error", func(t *testing.T) {
		// Create a test HTTP server that returns 500
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			})
		}))
		defer server.Close()

		// Create API client
		client := api.NewClient(server.URL, "test-token", api.WithRetries(0))
		_, err := client.GetVersion(context.Background())

		// Verify error
		require.Error(t, err)
		var apiErr *api.Error
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, 500, apiErr.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		// Create a test HTTP server that returns 404
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Endpoint not found",
			})
		}))
		defer server.Close()

		// Create API client
		client := api.NewClient(server.URL, "test-token")
		_, err := client.GetVersion(context.Background())

		// Verify error
		require.Error(t, err)
		assert.True(t, api.IsNotFound(err))
	})
}

// TestNewVerifyCommand tests that the command is properly configured
func TestNewVerifyCommand(t *testing.T) {
	cmd := NewVerifyCommand()

	assert.Equal(t, "verify", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}

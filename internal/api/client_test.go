package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_DebugRedactsSensitiveJSONFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"uuid":"token-1"}`))
	}))
	defer server.Close()

	var logs bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previousWriter) })

	client := NewClient(server.URL, "coolify-token", WithDebug(true), WithRetries(0))
	var response map[string]any
	err := client.Post(context.Background(), "cloud-tokens", map[string]string{"name": "primary", "token": "provider-secret"}, &response)
	require.NoError(t, err)
	assert.NotContains(t, logs.String(), "provider-secret")
	assert.Contains(t, logs.String(), `"token":"********"`)
}

func TestClient_DebugRedactsSensitiveResponseFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"uuid":"token-1","token":"provider-secret"}`))
	}))
	defer server.Close()

	var logs bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previousWriter) })

	client := NewClient(server.URL, "coolify-token", WithDebug(true), WithRetries(0))
	var response map[string]any
	err := client.Get(context.Background(), "cloud-tokens/token-1", &response)
	require.NoError(t, err)
	assert.NotContains(t, logs.String(), "provider-secret")
	assert.Contains(t, logs.String(), `"token":"********"`)
}

func TestNewClient(t *testing.T) {
	t.Run("creates client with defaults", func(t *testing.T) {
		client := NewClient("https://app.coolify.io", "test-token")

		assert.Equal(t, "https://app.coolify.io", client.baseURL)
		assert.Equal(t, "test-token", client.token)
		assert.Equal(t, defaultTimeout, client.timeout)
		assert.Equal(t, defaultRetries, client.retries)
		assert.False(t, client.debug)
	})

	t.Run("applies options", func(t *testing.T) {
		customTimeout := 10 * time.Second
		client := NewClient(
			"https://app.coolify.io",
			"test-token",
			WithDebug(true),
			WithTimeout(customTimeout),
			WithRetries(5),
		)

		assert.True(t, client.debug)
		assert.Equal(t, customTimeout, client.timeout)
		assert.Equal(t, 5, client.retries)
	})
}

func TestClient_Get_Success(t *testing.T) {
	type ServerResponse struct {
		UUID string `json:"uuid"`
		Name string `json:"name"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v1/servers", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]ServerResponse{
			{UUID: "uuid-1", Name: "server-1"},
			{UUID: "uuid-2", Name: "server-2"},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	var servers []ServerResponse

	err := client.Get(context.Background(), "servers", &servers)

	require.NoError(t, err)
	assert.Len(t, servers, 2)
	assert.Equal(t, "uuid-1", servers[0].UUID)
	assert.Equal(t, "server-1", servers[0].Name)
}

func TestClient_Get_StringResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/version", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("4.0.0"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	var version string

	err := client.Get(context.Background(), "version", &version)

	require.NoError(t, err)
	assert.Equal(t, "4.0.0", version)
}

func TestClient_Get_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Server not found",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	var result interface{}

	err := client.Get(context.Background(), "servers/unknown", &result)

	require.Error(t, err)
	assert.True(t, IsNotFound(err))

	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.StatusCode)
	assert.Equal(t, "Server not found", apiErr.Message)
}

func TestClient_Get_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Invalid token",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "invalid-token")
	var result interface{}

	err := client.Get(context.Background(), "servers", &result)

	require.Error(t, err)
	assert.True(t, IsUnauthorized(err))
}

func TestClient_Post_Success(t *testing.T) {
	type CreateServerRequest struct {
		Name string `json:"name"`
		IP   string `json:"ip"`
	}

	type CreateServerResponse struct {
		UUID    string `json:"uuid"`
		Message string `json:"message"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/servers", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req CreateServerRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "test-server", req.Name)
		assert.Equal(t, "192.168.1.100", req.IP)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(CreateServerResponse{
			UUID:    "new-uuid",
			Message: "Server created",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	requestBody := CreateServerRequest{
		Name: "test-server",
		IP:   "192.168.1.100",
	}

	var response CreateServerResponse
	err := client.Post(context.Background(), "servers", requestBody, &response)

	require.NoError(t, err)
	assert.Equal(t, "new-uuid", response.UUID)
	assert.Equal(t, "Server created", response.Message)
}

func TestClient_Post_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Invalid IP address",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	requestBody := map[string]string{"ip": "invalid"}
	var response interface{}

	err := client.Post(context.Background(), "servers", requestBody, &response)

	require.Error(t, err)
	assert.True(t, IsBadRequest(err))
}

func TestClient_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v1/servers/test-uuid", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Server deleted",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	err := client.Delete(context.Background(), "servers/test-uuid")

	require.NoError(t, err)
}

func TestClient_GetVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/version", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("4.0.0-beta.383"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	version, err := client.GetVersion(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "4.0.0-beta.383", version)
}

func TestClient_Retry_Success(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", WithRetries(3))
	var result string

	err := client.Get(context.Background(), "test", &result)

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, attempts)
}

func TestClient_Retry_NoRetryOn4xx(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Bad request"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", WithRetries(3))
	var result interface{}

	err := client.Get(context.Background(), "test", &result)

	require.Error(t, err)
	assert.Equal(t, 1, attempts) // Should not retry on 400
	assert.True(t, IsBadRequest(err))
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	var result interface{}
	err := client.Get(ctx, "test", &result)

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", WithTimeout(100*time.Millisecond))
	var result interface{}

	err := client.Get(context.Background(), "test", &result)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "deadline exceeded")
}

func TestClient_Debug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	}))
	defer server.Close()

	// This test just verifies debug mode doesn't crash
	// In real usage, debug logs would go to stdout
	client := NewClient(server.URL, "test-token", WithDebug(true))
	var result string

	err := client.Get(context.Background(), "test", &result)

	require.NoError(t, err)
}

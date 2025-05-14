package client_test

import (
	"github.com/coollabsio/coolify-cli/pkg/client"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ReturnsAnErrorForBadHTTPCodes(t *testing.T) {
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/api/v1/resources", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bang, a non json error"))
	})
	server := httptest.NewServer(mockMux)
	defer server.Close()

	c := client.New(server.URL, "my-token")

	_, err := c.ListResources()

	assert.ErrorContains(t, err, "bang, a non json error")
}

func TestClient_ReturnsAnErrorForBadHTTPCodes_andMarshalsTheCoolifyMessageIfItCan(t *testing.T) {
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/api/v1/resources", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "bang"}`))
	})
	server := httptest.NewServer(mockMux)
	defer server.Close()

	c := client.New(server.URL, "my-token")

	_, err := c.ListResources()

	assert.ErrorContains(t, err, "bang")
}

func TestClient_SetsAppropriateHeadersWhenPerformingRequest(t *testing.T) {
	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/api/v1/resources", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer my-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
	})
	server := httptest.NewServer(mockMux)
	defer server.Close()

	c := client.New(server.URL, "my-token")

	_, _ = c.ListResources()
}

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

func TestTagService_UsesResourcePathsAndEscapesIdentifiers(t *testing.T) {
	requests := make(chan *http.Request, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Clone(r.Context())
		if r.Method == http.MethodPost {
			var body map[string]string
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, map[string]string{"tag_name": "production"}, body)
		}
		_ = json.NewEncoder(w).Encode([]models.Tag{{UUID: "tag-1", Name: "production"}})
	}))
	defer server.Close()

	svc := NewTagService(api.NewClient(server.URL, "token"))
	_, err := svc.ListForResource(context.Background(), TagResourceApplications, "app/id")
	require.NoError(t, err)
	_, err = svc.CreateForResource(context.Background(), TagResourceDatabases, "db id", "production")
	require.NoError(t, err)
	require.NoError(t, svc.DeleteForResource(context.Background(), TagResourceServices, "svc", "tag/id"))

	first, second, third := <-requests, <-requests, <-requests
	assert.Equal(t, http.MethodGet, first.Method)
	assert.Equal(t, "/api/v1/applications/app%2Fid/tags", first.URL.EscapedPath())
	assert.Equal(t, http.MethodPost, second.Method)
	assert.Equal(t, "/api/v1/databases/db%20id/tags", second.URL.EscapedPath())
	assert.Equal(t, http.MethodDelete, third.Method)
	assert.Equal(t, "/api/v1/services/svc/tags/tag%2Fid", third.URL.EscapedPath())
}

func TestTagService_RejectsUnsupportedResourceType(t *testing.T) {
	svc := NewTagService(api.NewClient("http://example.test", "token"))
	_, err := svc.ListForResource(context.Background(), TagResourceType("projects"), "uuid")
	require.ErrorContains(t, err, "unsupported tag resource type")
}

func TestDestinationService_CRUDAndServerPaths(t *testing.T) {
	type observed struct {
		method string
		path   string
		body   models.DestinationCreateRequest
	}
	requests := make(chan observed, 5)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		item := observed{method: r.Method, path: r.URL.EscapedPath()}
		if r.Method == http.MethodPost {
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&item.body))
		}
		requests <- item
		// List endpoints return arrays; Get/Create return a single object.
		// Wrong shapes make the client retry (decode errors), fill this
		// channel, and deadlock the handler on send.
		path := r.URL.EscapedPath()
		switch {
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && (path == "/api/v1/destinations" || path == "/api/v1/servers/server%2Fid/destinations"):
			_ = json.NewEncoder(w).Encode([]models.Destination{})
		default:
			// Get by uuid + CreateForServer expect a single Destination object.
			_ = json.NewEncoder(w).Encode(models.Destination{UUID: "destination-1", Name: "public"})
		}
	}))
	defer server.Close()

	svc := NewDestinationService(api.NewClient(server.URL, "token", api.WithRetries(0)))
	_, err := svc.List(context.Background())
	require.NoError(t, err)
	_, err = svc.ListByServer(context.Background(), "server/id")
	require.NoError(t, err)
	_, err = svc.Get(context.Background(), "destination/id")
	require.NoError(t, err)
	_, err = svc.CreateForServer(context.Background(), "server/id", models.DestinationCreateRequest{Name: "public", Network: "coolify", Type: "standalone"})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(context.Background(), "destination/id"))

	want := []observed{
		{method: http.MethodGet, path: "/api/v1/destinations"},
		{method: http.MethodGet, path: "/api/v1/servers/server%2Fid/destinations"},
		{method: http.MethodGet, path: "/api/v1/destinations/destination%2Fid"},
		{method: http.MethodPost, path: "/api/v1/servers/server%2Fid/destinations", body: models.DestinationCreateRequest{Name: "public", Network: "coolify", Type: "standalone"}},
		{method: http.MethodDelete, path: "/api/v1/destinations/destination%2Fid"},
	}
	for _, expected := range want {
		assert.Equal(t, expected, <-requests)
	}
}

func TestCloudTokenService_CRUDAndValidate(t *testing.T) {
	type observed struct{ method, path string }
	requests := make(chan observed, 6)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.EscapedPath()
		requests <- observed{r.Method, path}
		switch r.Method {
		case http.MethodPost, http.MethodPatch:
			// Validate posts with a nil body; only assert payload when one is present.
			if r.Body != nil && r.ContentLength != 0 {
				var body map[string]any
				assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				if path == "/api/v1/cloud-tokens" {
					assert.Equal(t, "hetzner", body["provider"])
					assert.Equal(t, "secret", body["token"])
				}
			}
		}
		// List returns an array; other endpoints return a single object.
		// Mismatched shapes cause decode-error retries that deadlock this channel.
		switch {
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && path == "/api/v1/cloud-tokens":
			_, _ = w.Write([]byte(`[{"uuid":"token-1","valid":true}]`))
		default:
			_, _ = w.Write([]byte(`{"uuid":"token-1","valid":true}`))
		}
	}))
	defer server.Close()

	svc := NewCloudTokenService(api.NewClient(server.URL, "token", api.WithRetries(0)))
	_, err := svc.List(context.Background())
	require.NoError(t, err)
	_, err = svc.Get(context.Background(), "token/id")
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), models.CloudTokenCreateRequest{Provider: models.CloudProviderHetzner, Token: "secret", Name: "primary"})
	require.NoError(t, err)
	_, err = svc.Update(context.Background(), "token/id", models.CloudTokenUpdateRequest{Name: "renamed"})
	require.NoError(t, err)
	_, err = svc.Validate(context.Background(), "token/id")
	require.NoError(t, err)
	require.NoError(t, svc.Delete(context.Background(), "token/id"))

	want := []observed{{"GET", "/api/v1/cloud-tokens"}, {"GET", "/api/v1/cloud-tokens/token%2Fid"}, {"POST", "/api/v1/cloud-tokens"}, {"PATCH", "/api/v1/cloud-tokens/token%2Fid"}, {"POST", "/api/v1/cloud-tokens/token%2Fid/validate"}, {"DELETE", "/api/v1/cloud-tokens/token%2Fid"}}
	for _, expected := range want {
		assert.Equal(t, expected, <-requests)
	}
}

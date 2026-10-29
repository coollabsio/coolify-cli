package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestDNSService_Create_ConflictReturnsRecordToReplace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/applications/app-uuid/dns", r.URL.Path)
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"message":"DNS record already points to 198.51.100.7.","conflict":{"record_id":"rec-1","current":"198.51.100.7","proposed":"203.0.113.10"}}`))
	}))
	defer server.Close()

	_, err := NewDNSService(api.NewClient(server.URL, "t")).Create(context.Background(), "applications", "app-uuid", models.DNSRecordCreateRequest{Hostname: "app.example.com"})

	var conflict *DNSConflictError
	require.True(t, errors.As(err, &conflict))
	assert.Equal(t, models.DNSConflict{RecordID: "rec-1", Current: "198.51.100.7", Proposed: "203.0.113.10"}, conflict.Conflict)
}

func TestDNSService_Create_ShowsValidationErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"Validation failed.","errors":{"address":["The server has no public IP address; add the DNS record manually."]}}`))
	}))
	defer server.Close()

	_, err := NewDNSService(api.NewClient(server.URL, "t")).Create(context.Background(), "services", "svc", models.DNSRecordCreateRequest{Hostname: "app.example.com"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "address: The server has no public IP address")
}

func TestDNSService_Release_SendsDeleteRecordFlagAndReadsResult(t *testing.T) {
	var gotURI string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURI = r.Method + " " + r.URL.RequestURI()
		_, _ = w.Write([]byte(`{"result":"kept","message":"kept"}`))
	}))
	defer server.Close()

	resp, err := NewDNSService(api.NewClient(server.URL, "t")).Release(context.Background(), "services", "svc-uuid", "App.Example.com", false)

	require.NoError(t, err)
	assert.Equal(t, "DELETE /api/v1/services/svc-uuid/dns/app.example.com?delete_record=false", gotURI)
	assert.Equal(t, "kept", resp.Result)
}

func TestDNSService_RejectsOtherResourceTypes(t *testing.T) {
	_, err := NewDNSService(api.NewClient("http://127.0.0.1:1", "t")).Status(context.Background(), "databases", "x")
	require.ErrorContains(t, err, `unsupported resource type "databases"`)
}

func TestSecretManagerService_UsesResourceTypeInPath(t *testing.T) {
	var gotURIs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURIs = append(gotURIs, r.Method+" "+r.URL.Path)
		_, _ = w.Write([]byte(`{"imported":["API_KEY"],"skipped":[]}`))
	}))
	defer server.Close()
	svc := NewSecretManagerService(api.NewClient(server.URL, "t"))

	resp, err := svc.Import(context.Background(), "databases", "db-uuid")
	require.NoError(t, err)
	_, err = svc.Keys(context.Background(), "services", "svc-uuid")
	require.NoError(t, err)

	assert.Equal(t, []string{"POST /api/v1/databases/db-uuid/secret-manager/import", "GET /api/v1/services/svc-uuid/secret-manager/keys"}, gotURIs)
	assert.Equal(t, []string{"API_KEY"}, resp.Imported)
	_, err = svc.Keys(context.Background(), "servers", "x")
	require.ErrorContains(t, err, "unsupported resource type")
}

func TestValidateIntegrationTokenOptions_Cloudflare(t *testing.T) {
	yes := true
	tests := []struct {
		name    string
		req     models.IntegrationTokenCreateRequest
		wantErr string
	}{
		{name: "dns and tunnel", req: models.IntegrationTokenCreateRequest{Provider: "cloudflare", Name: "cf", Capabilities: []string{"dns", "tunnel"}, AutomaticDNS: &yes}},
		{name: "tunnel without dns", req: models.IntegrationTokenCreateRequest{Provider: "cloudflare", Name: "cf", Capabilities: []string{"tunnel"}}, wantErr: "the tunnel capability also needs the dns capability"},
		{name: "secrets on cloudflare", req: models.IntegrationTokenCreateRequest{Provider: "cloudflare", Name: "cf", Capabilities: []string{"secrets"}}, wantErr: "--capabilities for cloudflare must be one of: dns, tunnel"},
		{name: "metadata on cloudflare", req: models.IntegrationTokenCreateRequest{Provider: "cloudflare", Name: "cf", Metadata: map[string]any{"base_url": "x"}}, wantErr: "--metadata is not used for cloudflare tokens"},
		{name: "dns options on doppler", req: models.IntegrationTokenCreateRequest{Provider: "doppler", Name: "d", AutomaticDNS: &yes}, wantErr: "only for cloudflare tokens"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntegrationTokenOptions(tt.req)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

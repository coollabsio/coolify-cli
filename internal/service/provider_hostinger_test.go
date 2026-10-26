package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestHostingerService_OptionsUseTokenQuery(t *testing.T) {
	responses := map[string]string{
		"/api/v1/hostinger/data-centers":         `[{"id":9,"name":"us-central","location":"us","city":"Boston","continent":"North America"}]`,
		"/api/v1/hostinger/catalog":              `[{"id":"hostingercom-vps-kvm1","name":"KVM 1","category":"VPS","metadata":{"cpus":"1","memory":"4096","disk_space":"51200","bandwidth":"4096000","network":"300"},"prices":[{"id":"hostingercom-vps-kvm1-usd-1m","name":"KVM 1 (monthly)","currency":"USD","price":1399,"first_period_price":649,"period":1,"period_unit":"month"}]}]`,
		"/api/v1/hostinger/templates":            `[{"id":1002,"name":"Ubuntu 24.04","description":"Plain Ubuntu","documentation":"https://example.com"}]`,
		"/api/v1/hostinger/ssh-keys":             `[{"id":5,"name":"laptop","key":"ssh-ed25519 AAAA"}]`,
		"/api/v1/hostinger/post-install-scripts": `[{"id":3,"name":"setup","content":"#!/bin/sh","created_at":"2026-01-01","updated_at":"2026-01-02"}]`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "token/id", r.URL.Query().Get("cloud_provider_token_uuid"))
		body, ok := responses[r.URL.Path]
		if !assert.True(t, ok, r.URL.Path) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()
	ctx := context.Background()
	hostinger := NewHostingerService(api.NewClient(server.URL, "token", api.WithRetries(0)))

	dataCenters, err := hostinger.DataCenters(ctx, "token/id")
	require.NoError(t, err)
	assert.Equal(t, []models.HostingerDataCenter{{ID: 9, Name: "us-central", Location: "us", City: "Boston", Continent: "North America"}}, dataCenters)

	catalog, err := hostinger.Catalog(ctx, "token/id")
	require.NoError(t, err)
	require.Len(t, catalog, 1)
	assert.Equal(t, "4096", catalog[0].Metadata.Memory)
	require.Len(t, catalog[0].Prices, 1)
	assert.Equal(t, "hostingercom-vps-kvm1-usd-1m", catalog[0].Prices[0].ID)
	assert.Equal(t, 649, catalog[0].Prices[0].FirstPeriodPrice)

	templates, err := hostinger.Templates(ctx, "token/id")
	require.NoError(t, err)
	assert.Equal(t, 1002, templates[0].ID)

	keys, err := hostinger.SSHKeys(ctx, "token/id")
	require.NoError(t, err)
	assert.Equal(t, models.HostingerSSHKey{ID: 5, Name: "laptop", Key: "ssh-ed25519 AAAA"}, keys[0])

	scripts, err := hostinger.PostInstallScripts(ctx, "token/id")
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/sh", scripts[0].Content)
}

func TestHostingerService_PostInstallScripts_ForbiddenSurfacesMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Token lacks permission"}`))
	}))
	defer server.Close()

	_, err := NewHostingerService(api.NewClient(server.URL, "token")).PostInstallScripts(context.Background(), "t")

	require.Error(t, err)
	assert.True(t, api.IsUnauthorized(err))
	assert.Contains(t, err.Error(), "Token lacks permission")
}

func TestHostingerService_Create_SendsSchemaFieldsOnly(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/servers/hostinger", r.URL.RequestURI())
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"uuid":"server-1","hostinger_virtual_machine_id":42,"ip":"192.0.2.1","provisioning":true}`))
	}))
	defer server.Close()

	response, err := NewHostingerService(api.NewClient(server.URL, "token")).Create(context.Background(), models.HostingerServerCreateRequest{
		CloudProviderTokenUUID: "token-1",
		ItemID:                 "hostingercom-vps-kvm1-usd-1m",
		DataCenterID:           9,
		TemplateID:             1002,
		PrivateKeyUUID:         "key-1",
		Name:                   "web-1",
		EnableBackups:          false,
		PublicKeyIDs:           []int{5, 6},
		PostInstallScriptID:    3,
		InstantValidate:        true,
	})

	require.NoError(t, err)
	assert.Equal(t, &models.HostingerServerCreateResponse{UUID: "server-1", HostingerVirtualMachineID: 42, IP: "192.0.2.1", Provisioning: true}, response)
	keys := make([]string, 0, len(body))
	for key := range body {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	assert.Equal(t, []string{"cloud_provider_token_uuid", "data_center_id", "enable_backups", "instant_validate", "item_id", "name", "post_install_script_id", "private_key_uuid", "public_key_ids", "template_id"}, keys)
	assert.Equal(t, "hostingercom-vps-kvm1-usd-1m", body["item_id"])
	assert.InDelta(t, 9, body["data_center_id"], 0)
	assert.Equal(t, false, body["enable_backups"])
	assert.Equal(t, []any{float64(5), float64(6)}, body["public_key_ids"])
	assert.InDelta(t, 3, body["post_install_script_id"], 0)
}

func TestHostingerService_Create_OmitsOptionalFields(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"uuid":"server-1","hostinger_virtual_machine_id":42,"ip":"192.0.2.1","provisioning":false}`))
	}))
	defer server.Close()

	_, err := NewHostingerService(api.NewClient(server.URL, "token")).Create(context.Background(), models.HostingerServerCreateRequest{CloudProviderTokenUUID: "t", ItemID: "i", DataCenterID: 1, TemplateID: 2, PrivateKeyUUID: "k", EnableBackups: true})

	require.NoError(t, err)
	assert.NotContains(t, body, "name")
	assert.NotContains(t, body, "public_key_ids")
	assert.NotContains(t, body, "post_install_script_id")
	assert.Equal(t, true, body["enable_backups"])
}

func TestHostingerService_Create_ErrorResponses(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		retryAfter   string
		response     string
		wantPending  bool
		wantContains []string
	}{
		{
			name:         "payment processing",
			status:       http.StatusAccepted,
			response:     `{"message":"Payment is being processed"}`,
			wantPending:  true,
			wantContains: []string{"hostinger payment is still processing", "Payment is being processed"},
		},
		{
			name:         "validation error",
			status:       http.StatusUnprocessableEntity,
			response:     `{"message":"The given data was invalid.","errors":{"template_id":["The template id field is required."],"item_id":["Invalid item."]}}`,
			wantContains: []string{"422", "The given data was invalid.", "item_id: Invalid item.; template_id: The template id field is required."},
		},
		{
			name:         "rate limited",
			status:       http.StatusTooManyRequests,
			retryAfter:   "60",
			response:     `{"message":"Too Many Attempts."}`,
			wantContains: []string{"429", "retry after 60 seconds"},
		},
		{
			name:         "server error after purchase",
			status:       http.StatusInternalServerError,
			response:     `{"message":"Failed to add server","hostinger_virtual_machine_id":42}`,
			wantContains: []string{"500", "Failed to add server", "Hostinger virtual machine 42 may already have been purchased"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				attempts.Add(1)
				if tt.retryAfter != "" {
					w.Header().Set("Retry-After", tt.retryAfter)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			// Default client retries 5xx/429; Create must still send exactly one request.
			response, err := NewHostingerService(api.NewClient(server.URL, "token")).Create(context.Background(), models.HostingerServerCreateRequest{CloudProviderTokenUUID: "t", ItemID: "i", DataCenterID: 1, TemplateID: 2, PrivateKeyUUID: "k"})

			require.Error(t, err)
			assert.Equal(t, int32(1), attempts.Load())
			assert.Equal(t, tt.wantPending, errorsIsPending(err))
			if tt.wantPending {
				require.NotNil(t, response)
				assert.Equal(t, "Payment is being processed", response.Message)
			} else {
				assert.Nil(t, response)
			}
			for _, want := range tt.wantContains {
				assert.Contains(t, err.Error(), want)
			}
		})
	}
}

func errorsIsPending(err error) bool {
	return errors.Is(err, ErrHostingerPaymentPending)
}

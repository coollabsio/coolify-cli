package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

const auditPageFixture = `{
	"current_page": 2,
	"data": [{
		"id": 41, "team_id": 0, "event": "api.application.updated", "source": "api", "action": "updated",
		"level": "info", "actor_type": "api_token", "actor_id": 1, "actor_name": "Root",
		"resource_type": "application", "resource_uuid": "app-uuid", "resource_name": "web",
		"description": "web Updated", "created_at": "2026-10-01T10:00:00.000000Z",
		"actor_email": "root@example.com", "metadata": {"application_uuid": "app-uuid"}, "changes": null
	}],
	"first_page_url": "http://x/api/v1/audit-events?page=1",
	"per_page": 10, "last_page": 3, "total": 21
}`

func TestAuditService_List(t *testing.T) {
	tests := []struct {
		name      string
		opts      models.AuditEventListOptions
		wantQuery map[string]string
	}{
		{
			name:      "no filters",
			opts:      models.AuditEventListOptions{},
			wantQuery: map[string]string{},
		},
		{
			name: "all filters",
			opts: models.AuditEventListOptions{PerPage: 10, Page: 2, Search: "web app", Action: "updated", Source: "api"},
			wantQuery: map[string]string{
				"per_page": "10", "page": "2", "search": "web app", "action": "updated", "source": "api",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v1/audit-events", r.URL.Path)
				query := r.URL.Query()
				assert.Len(t, query, len(tt.wantQuery))
				for key, want := range tt.wantQuery {
					assert.Equal(t, want, query.Get(key), key)
				}
				_, _ = w.Write([]byte(auditPageFixture))
			}))
			defer server.Close()

			svc := NewAuditService(api.NewClient(server.URL, "test-token"))
			page, err := svc.List(context.Background(), tt.opts)
			require.NoError(t, err)
			assert.Equal(t, 2, page.CurrentPage)
			assert.Equal(t, 3, page.LastPage)
			assert.Equal(t, 21, page.Total)
			assert.Equal(t, 10, page.PerPage)
			require.Len(t, page.Data, 1)
			event := page.Data[0]
			assert.Equal(t, 41, event.ID)
			assert.Equal(t, "api", event.Source)
			assert.Equal(t, "updated", event.Action)
			require.NotNil(t, event.ResourceUUID)
			assert.Equal(t, "app-uuid", *event.ResourceUUID)
			require.NotNil(t, event.ActorEmail)
			assert.Equal(t, "root@example.com", *event.ActorEmail)
		})
	}
}

func TestAuditService_List_PlainArrayWithTotalHeader(t *testing.T) {
	tests := []struct {
		name         string
		opts         models.AuditEventListOptions
		totalHeader  string
		wantPage     int
		wantPerPage  int
		wantTotal    int
		wantLastPage int
	}{
		{name: "explicit page", opts: models.AuditEventListOptions{PerPage: 10, Page: 2}, totalHeader: "21", wantPage: 2, wantPerPage: 10, wantTotal: 21, wantLastPage: 3},
		{name: "server defaults", opts: models.AuditEventListOptions{}, totalHeader: "26", wantPage: 1, wantPerPage: 25, wantTotal: 26, wantLastPage: 2},
		{name: "no events", opts: models.AuditEventListOptions{}, totalHeader: "0", wantPage: 1, wantPerPage: 25, wantTotal: 0, wantLastPage: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("X-Total-Count", tt.totalHeader)
				if tt.wantTotal == 0 {
					_, _ = w.Write([]byte(`[]`))
					return
				}
				_, _ = w.Write([]byte(`[{"id": 41, "event": "api.application.updated", "source": "api", "action": "updated", "level": "info", "created_at": "2026-10-01T10:00:00.000000Z"}]`))
			}))
			defer server.Close()

			page, err := NewAuditService(api.NewClient(server.URL, "test-token")).List(context.Background(), tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPage, page.CurrentPage)
			assert.Equal(t, tt.wantPerPage, page.PerPage)
			assert.Equal(t, tt.wantTotal, page.Total)
			assert.Equal(t, tt.wantLastPage, page.LastPage)
			if tt.wantTotal > 0 {
				require.Len(t, page.Data, 1)
				assert.Equal(t, 41, page.Data[0].ID)
			}
		})
	}
}

func TestAuditService_List_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Only team admins and owners can view audit logs."}`))
	}))
	defer server.Close()

	svc := NewAuditService(api.NewClient(server.URL, "test-token"))
	_, err := svc.List(context.Background(), models.AuditEventListOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Only team admins and owners")
}

func TestValidateAuditListOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    models.AuditEventListOptions
		wantErr string
	}{
		{name: "defaults", opts: models.AuditEventListOptions{}},
		{name: "bounds ok", opts: models.AuditEventListOptions{PerPage: 100, Page: 1}},
		{name: "per page too small", opts: models.AuditEventListOptions{PerPage: -1}, wantErr: "--per-page"},
		{name: "per page too large", opts: models.AuditEventListOptions{PerPage: 101}, wantErr: "--per-page"},
		{name: "negative page", opts: models.AuditEventListOptions{Page: -1}, wantErr: "--page"},
		{name: "search too long", opts: models.AuditEventListOptions{Search: strings.Repeat("a", 256)}, wantErr: "--search"},
		{name: "action too long", opts: models.AuditEventListOptions{Action: strings.Repeat("a", 256)}, wantErr: "--action"},
		{name: "invalid source", opts: models.AuditEventListOptions{Source: "cli"}, wantErr: "--source must be one of: all, ui, api, mcp, webhook, system, scheduler"},
	}
	for _, source := range models.AuditEventSources {
		tests = append(tests, struct {
			name    string
			opts    models.AuditEventListOptions
			wantErr string
		}{name: "source " + source, opts: models.AuditEventListOptions{Source: source}})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuditListOptions(tt.opts)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestAuditService_List_InvalidOptionsSkipRequest(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
	}))
	defer server.Close()

	svc := NewAuditService(api.NewClient(server.URL, "test-token"))
	_, err := svc.List(context.Background(), models.AuditEventListOptions{Source: "bogus"})
	require.Error(t, err)
	assert.Zero(t, calls.Load())
}

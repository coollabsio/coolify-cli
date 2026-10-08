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

func TestAuditEventServiceList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/audit-events", r.URL.Path)
		assert.Equal(t, "api", r.URL.Query().Get("source"))
		assert.Equal(t, "updated", r.URL.Query().Get("action"))
		assert.Equal(t, "website app", r.URL.Query().Get("search"))
		assert.Equal(t, "50", r.URL.Query().Get("per_page"))
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		_ = json.NewEncoder(w).Encode(models.AuditEventsPage{
			CurrentPage: 2,
			Data:        []models.AuditEvent{{Event: "api.application.updated", ActorEmail: "admin@example.com"}},
			LastPage:    2,
			PerPage:     50,
			Total:       51,
		})
	}))
	defer server.Close()

	result, err := NewAuditEventService(api.NewClient(server.URL, "token")).List(context.Background(), AuditEventListOptions{
		Source: "api", Action: "updated", Search: "website app", PerPage: 50, Page: 2,
	})
	require.NoError(t, err)
	require.Len(t, result.Data, 1)
	assert.Equal(t, "api.application.updated", result.Data[0].Event)
	assert.Equal(t, "admin@example.com", result.Data[0].ActorEmail)
	assert.Equal(t, 51, result.Total)
}

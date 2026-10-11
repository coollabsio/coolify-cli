package audit

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const pageFixture = `{
	"current_page": 1,
	"data": [
		{"id": 2, "team_id": 0, "event": "api.application.deleted", "source": "api", "action": "deleted", "level": "warning",
		 "actor_type": "api_token", "actor_id": 1, "actor_name": "Root", "resource_type": "application",
		 "resource_uuid": "app-uuid", "resource_name": "web", "description": "web Deleted", "created_at": "2026-10-01T10:00:00.000000Z"},
		{"id": 1, "team_id": 0, "event": "ui.server.created", "source": "ui", "action": "created", "level": "info",
		 "actor_type": "user", "actor_id": 1, "actor_name": "Root", "resource_type": "server",
		 "resource_uuid": "srv-uuid", "resource_name": "edge", "description": "edge Created", "created_at": "2026-09-30T10:00:00.000000Z"}
	],
	"per_page": 2, "last_page": 4, "total": 8
}`

func TestListCommand_TableOutput(t *testing.T) {
	fake := newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(pageFixture))
	})

	stdout, _, err := runCmd(t, "", "audit", "list", "--per-page", "2", "--source", "api", "--action", "deleted", "--search", "web app")
	require.NoError(t, err)

	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, http.MethodGet, reqs[0].Method)
	parsed, err := url.Parse(reqs[0].RequestURI)
	require.NoError(t, err)
	assert.Equal(t, "/api/v1/audit-events", parsed.Path)
	assert.Equal(t, url.Values{
		"per_page": {"2"}, "page": {"1"}, "source": {"api"}, "action": {"deleted"}, "search": {"web app"},
	}, parsed.Query())

	for _, want := range []string{"created_at", "source", "action", "level", "actor_name", "resource_type", "resource_name", "description", "web Deleted", "edge Created"} {
		assert.Contains(t, stdout, want)
	}
	// Internal ids and noisy columns are hidden from the table.
	assert.NotContains(t, stdout, "team_id")
	assert.NotContains(t, stdout, "actor_type")
	assert.Contains(t, stdout, "Page 1 of 4 (8 events). Use --page 2 for more.")
}

func TestListCommand_JSONOutputIsFullPage(t *testing.T) {
	newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(pageFixture))
	})

	stdout, _, err := runCmd(t, "", "audit-events", "list", "--format", "json")
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.EqualValues(t, 8, out["total"])
	assert.EqualValues(t, 4, out["last_page"])
	data := out["data"].([]any)
	require.Len(t, data, 2)
	assert.Equal(t, "app-uuid", data[0].(map[string]any)["resource_uuid"])
}

func TestListCommand_EmptyPage(t *testing.T) {
	newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"current_page":1,"data":[],"per_page":25,"last_page":1,"total":0}`))
	})
	stdout, _, err := runCmd(t, "", "audit", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "No data")
	assert.NotContains(t, stdout, "Page ")
}

func TestListCommand_Validation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "invalid source", args: []string{"--source", "cli"}, wantErr: "--source must be one of: all, ui, api, mcp, webhook, system, scheduler"},
		{name: "per page zero", args: []string{"--per-page", "0"}, wantErr: "--per-page must be between 1 and 100"},
		{name: "per page too large", args: []string{"--per-page", "101"}, wantErr: "--per-page must be between 1 and 100"},
		{name: "page zero", args: []string{"--page", "0"}, wantErr: "--page must be 1 or greater"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeAPI(t, func(http.ResponseWriter, *http.Request) {})
			_, _, err := runCmd(t, "", append([]string{"audit", "list"}, tt.args...)...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Empty(t, fake.recorded())
		})
	}
}

func TestListCommand_Forbidden(t *testing.T) {
	newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Only team admins and owners can view audit logs."}`))
	})
	_, _, err := runCmd(t, "", "audit", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Only team admins and owners")
}

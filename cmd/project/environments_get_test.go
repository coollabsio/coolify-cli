package project

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const environmentFixture = `{
	"id": 3, "uuid": "env-uuid", "name": "production", "project_id": 1, "description": "Prod",
	"created_at": "2026-01-01T00:00:00.000000Z", "updated_at": "2026-01-01T00:00:00.000000Z",
	"applications": [{"id": 10, "uuid": "app-1", "name": "web", "status": "running:healthy"}],
	"postgresqls": [{"id": 11, "uuid": "pg-1", "name": "db", "status": "exited"}],
	"redis": [], "mongodbs": [], "mysqls": [], "mariadbs": [], "services": []
}`

func TestEnvironmentsGet_TableOutput(t *testing.T) {
	fake := newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(environmentFixture))
	})

	stdout, _, err := runCmd(t, "", "project", "env", "get", "proj-uuid", "production")
	require.NoError(t, err)

	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, http.MethodGet, reqs[0].Method)
	assert.Equal(t, "/api/v1/projects/proj-uuid/production", reqs[0].RequestURI)

	for _, want := range []string{"env-uuid", "production", "Prod", "application", "app-1", "web", "running:healthy", "postgresql", "pg-1", "exited"} {
		assert.Contains(t, stdout, want)
	}
	assert.NotContains(t, stdout, "project_id")
}

func TestEnvironmentsGet_NoResources(t *testing.T) {
	newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id": 3, "uuid": "env-uuid", "name": "empty", "project_id": 1}`))
	})
	stdout, _, err := runCmd(t, "", "project", "environments", "get", "proj-uuid", "empty")
	require.NoError(t, err)
	assert.Contains(t, stdout, "env-uuid")
	assert.Contains(t, stdout, "No resources in this environment.")
}

func TestEnvironmentsGet_JSONOutput(t *testing.T) {
	newFakeAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(environmentFixture))
	})
	stdout, _, err := runCmd(t, "", "project", "env", "get", "proj-uuid", "env-uuid", "--format", "json")
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "env-uuid", out["uuid"])
	apps := out["applications"].([]any)
	require.Len(t, apps, 1)
	assert.Equal(t, "app-1", apps[0].(map[string]any)["uuid"])
}

func TestEnvironmentsGet_RequiresTwoArgs(t *testing.T) {
	fake := newFakeAPI(t, func(http.ResponseWriter, *http.Request) {})
	_, _, err := runCmd(t, "", "project", "env", "get", "proj-uuid")
	require.Error(t, err)
	assert.Empty(t, fake.recorded())
}

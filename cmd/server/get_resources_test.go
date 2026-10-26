package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const serverResourcesFixture = `[
	{"id": 7, "uuid": "res-1", "name": "web", "type": "application", "status": "running:healthy", "created_at": "2026-01-01T00:00:00.000000Z", "updated_at": "2026-01-02T00:00:00.000000Z"},
	{"id": 8, "uuid": "res-2", "name": "pg", "type": "standalone-postgresql", "status": "exited:unhealthy", "created_at": "2026-01-01T00:00:00.000000Z", "updated_at": "2026-01-02T00:00:00.000000Z"}
]`

func runServerGet(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := &cobra.Command{Use: "coolify", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("token", "", "")
	root.PersistentFlags().String("context", "", "")
	root.PersistentFlags().String("format", "table", "")
	root.PersistentFlags().BoolP("show-sensitive", "s", false, "")
	root.PersistentFlags().Bool("debug", false, "")
	root.AddCommand(NewServerCommand())

	var stdout bytes.Buffer
	root.SetIn(strings.NewReader(""))
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(append([]string{"server", "get"}, args...))
	err := root.ExecuteContext(context.Background())
	return stdout.String(), err
}

func TestServerGetResources_UsesDedicatedEndpoint(t *testing.T) {
	fake := newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(serverResourcesFixture))
	})

	_, err := runServerGet(t, "srv-uuid", "--resources")
	require.NoError(t, err)

	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, http.MethodGet, reqs[0].Method)
	assert.Equal(t, "/api/v1/servers/srv-uuid/resources", reqs[0].RequestURI)
}

func TestServerGetResources_JSONKeepsAllFields(t *testing.T) {
	newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(serverResourcesFixture))
	})

	stdout, err := runServerGet(t, "srv-uuid", "--resources", "--format", "json")
	require.NoError(t, err)

	var out []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	require.Len(t, out, 2)
	assert.EqualValues(t, 7, out[0]["id"])
	assert.Equal(t, "res-1", out[0]["uuid"])
	assert.Equal(t, "application", out[0]["type"])
	assert.Equal(t, "running:healthy", out[0]["status"])
	assert.Equal(t, "2026-01-01T00:00:00.000000Z", out[0]["created_at"])
}

func TestServerGetResources_TableHidesID(t *testing.T) {
	newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(serverResourcesFixture))
	})

	stdout, err := runServerGet(t, "srv-uuid", "--resources")
	require.NoError(t, err)

	for _, want := range []string{"uuid", "name", "type", "status", "res-1", "standalone-postgresql", "exited:unhealthy"} {
		assert.Contains(t, stdout, want)
	}
	assert.NotContains(t, stdout, "created_at")
	assert.NotContains(t, stdout, "Note: Use -s")
}

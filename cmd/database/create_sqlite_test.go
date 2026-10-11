package database

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runDatabaseCommand runs `coolify database <args>` against a fake API that
// records the request path and JSON body.
func runDatabaseCommand(t *testing.T, args ...string) (path, body string, err error) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		path, body = r.URL.Path, string(data)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"uuid":"sqlite-uuid"}`))
	}))
	t.Cleanup(server.Close)

	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "coolify")
	require.NoError(t, os.MkdirAll(configDir, 0o750))
	cfg, _ := json.Marshal(map[string]any{
		"instances": []map[string]any{{"name": "test", "fqdn": server.URL, "token": "test-token", "default": true}},
	})
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), cfg, 0o600))

	root := &cobra.Command{Use: "coolify", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("token", "", "")
	root.PersistentFlags().String("context", "", "")
	root.PersistentFlags().String("format", "json", "")
	root.PersistentFlags().Bool("debug", false, "")
	root.AddCommand(NewDatabaseCommand())
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(append([]string{"database"}, args...))
	err = root.ExecuteContext(context.Background())
	return path, body, err
}

func TestCreateSqlite_SendsSqliteDatabases(t *testing.T) {
	path, body, err := runDatabaseCommand(t, "create", "sqlite",
		"--server-uuid", "srv-1", "--project-uuid", "proj-1", "--environment-name", "production",
		"--name", "app", "--sqlite-databases", "app.db,cache.db", "--postgres-user", "ignored")
	require.NoError(t, err)
	assert.Equal(t, "/api/v1/databases/sqlite", path)
	assert.JSONEq(t, `{
		"server_uuid":"srv-1",
		"project_uuid":"proj-1",
		"environment_name":"production",
		"name":"app",
		"sqlite_databases":"app.db,cache.db"
	}`, body)
}

func TestCreateSqlite_RejectsPublicFlags(t *testing.T) {
	for _, flag := range [][]string{{"--is-public"}, {"--public-port", "5432"}} {
		args := append([]string{"create", "sqlite", "--server-uuid", "srv-1", "--project-uuid", "proj-1", "--environment-name", "production"}, flag...)
		path, _, err := runDatabaseCommand(t, args...)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not supported for sqlite")
		assert.Empty(t, path, "no request should be sent")
	}
}

func TestCreate_SqliteDatabasesIgnoredForOtherTypes(t *testing.T) {
	_, body, err := runDatabaseCommand(t, "create", "postgresql",
		"--server-uuid", "srv-1", "--project-uuid", "proj-1", "--environment-name", "production", "--sqlite-databases", "app.db")
	require.NoError(t, err)
	assert.NotContains(t, body, "sqlite_databases")
}

func TestDatabaseCommandRegistersImportCommands(t *testing.T) {
	command := NewDatabaseCommand()
	for _, name := range []string{"upload", "start", "status"} {
		child, _, err := command.Find([]string{"import", name})
		require.NoError(t, err)
		require.Equal(t, name, child.Name())
	}
}

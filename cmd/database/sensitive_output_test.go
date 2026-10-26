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

const databaseSentinelPassword = "sentinel-db-password"

// runDatabaseOutputCommand runs `coolify database <args>` against a fake API
// that returns one PostgreSQL database and returns what it printed to stdout.
func runDatabaseOutputCommand(t *testing.T, args ...string) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"uuid":"db-1","name":"main","status":"running","type":"standalone-postgresql",` +
			`"postgres_user":"app","postgres_password":"` + databaseSentinelPassword + `","postgres_db":"app"}]`))
	}))
	t.Cleanup(server.Close)

	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "coolify")
	require.NoError(t, os.MkdirAll(configDir, 0o750))
	cfg, err := json.Marshal(map[string]any{
		"instances": []map[string]any{{"name": "test", "fqdn": server.URL, "token": "test-token", "default": true}},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), cfg, 0o600))

	root := &cobra.Command{Use: "coolify", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("token", "", "")
	root.PersistentFlags().String("context", "", "")
	root.PersistentFlags().String("format", "table", "")
	root.PersistentFlags().Bool("show-sensitive", false, "")
	root.PersistentFlags().Bool("debug", false, "")
	root.AddCommand(NewDatabaseCommand())
	root.SetArgs(append([]string{"database"}, args...))

	read, write, err := os.Pipe()
	require.NoError(t, err)
	originalStdout := os.Stdout
	os.Stdout = write
	t.Cleanup(func() { os.Stdout = originalStdout })

	require.NoError(t, root.ExecuteContext(context.Background()))
	require.NoError(t, write.Close())
	os.Stdout = originalStdout
	var out bytes.Buffer
	_, err = io.Copy(&out, read)
	require.NoError(t, err)
	return out.String()
}

func TestListCommand_MasksPasswordsInStructuredFormats(t *testing.T) {
	for _, format := range []string{"json", "pretty"} {
		t.Run(format, func(t *testing.T) {
			out := runDatabaseOutputCommand(t, "list", "--format", format)
			assert.NotContains(t, out, databaseSentinelPassword)

			var databases []map[string]any
			require.NoError(t, json.Unmarshal([]byte(out), &databases))
			require.Len(t, databases, 1)
			assert.Equal(t, "********", databases[0]["postgres_password"])
			assert.Equal(t, "app", databases[0]["postgres_user"])
		})
	}
}

func TestListCommand_ShowSensitiveRevealsPasswordsInStructuredFormats(t *testing.T) {
	for _, format := range []string{"json", "pretty"} {
		t.Run(format, func(t *testing.T) {
			out := runDatabaseOutputCommand(t, "list", "--format", format, "--show-sensitive")

			var databases []map[string]any
			require.NoError(t, json.Unmarshal([]byte(out), &databases))
			require.Len(t, databases, 1)
			assert.Equal(t, databaseSentinelPassword, databases[0]["postgres_password"])
		})
	}
}

func TestListCommand_TableNeverShowsPasswords(t *testing.T) {
	for _, showSensitive := range []bool{false, true} {
		args := []string{"list", "--format", "table"}
		if showSensitive {
			args = append(args, "--show-sensitive")
		}
		out := runDatabaseOutputCommand(t, args...)
		assert.NotContains(t, out, databaseSentinelPassword)
		assert.NotContains(t, out, "Use -s to show sensitive information")
	}
}

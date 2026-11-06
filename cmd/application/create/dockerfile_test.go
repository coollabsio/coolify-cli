package create

import (
	"context"
	"encoding/base64"
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

// sentDockerfile runs `create dockerfile --dockerfile <value>` against a fake
// API and returns the "dockerfile" field the API received.
func sentDockerfile(t *testing.T, value string) string {
	t.Helper()
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &got))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"app-1"}`))
	}))
	t.Cleanup(server.Close)

	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "coolify")
	require.NoError(t, os.MkdirAll(configDir, 0o750))
	cfg, err := json.Marshal(map[string]any{
		"instances": []map[string]any{{"name": "test", "fqdn": server.URL, "token": "t", "default": true}},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), cfg, 0o600))

	root := &cobra.Command{Use: "coolify", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("token", "", "")
	root.PersistentFlags().String("context", "", "")
	root.PersistentFlags().String("format", "json", "")
	root.PersistentFlags().BoolP("show-sensitive", "s", false, "")
	root.PersistentFlags().Bool("debug", false, "")
	root.AddCommand(NewDockerfileCommand())
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"dockerfile", "--server-uuid", "s", "--project-uuid", "p",
		"--environment-name", "production", "--dockerfile", value})
	require.NoError(t, root.ExecuteContext(context.Background()))

	dockerfile, ok := got["dockerfile"].(string)
	require.True(t, ok, "request has no dockerfile field: %v", got)
	return dockerfile
}

func TestCreateDockerfile_SendsBase64(t *testing.T) {
	raw := "FROM busybox:1.37\nRUN exit 42\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))

	assert.Equal(t, encoded, sentDockerfile(t, raw), "raw content must be encoded")
	assert.Equal(t, encoded, sentDockerfile(t, encoded), "encoded content must not be encoded twice")
}

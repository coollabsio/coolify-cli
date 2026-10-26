package privatekeys

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

const sentinelPrivateKey = "-----BEGIN OPENSSH PRIVATE KEY-----sentinel"

// runPrivateKeysCommand runs `coolify private-key <args>` against a fake API
// that returns one key and returns what it printed to stdout.
func runPrivateKeysCommand(t *testing.T, args ...string) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"uuid":"key-1","name":"deploy","public_key":"ssh-ed25519 AAAA","private_key":"` + sentinelPrivateKey + `"}]`))
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
	root.AddCommand(NewPrivateKeysCommand())
	root.SetArgs(append([]string{"private-key"}, args...))

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

func TestListCommand_NeverExposesPrivateKeyWithoutShowSensitive(t *testing.T) {
	for _, format := range []string{"table", "json", "pretty"} {
		t.Run(format, func(t *testing.T) {
			out := runPrivateKeysCommand(t, "list", "--format", format)
			assert.NotContains(t, out, sentinelPrivateKey)
			assert.Contains(t, out, "********")
		})
	}
}

func TestListCommand_ShowSensitiveRevealsPrivateKeyInJSON(t *testing.T) {
	out := runPrivateKeysCommand(t, "list", "--format", "json", "--show-sensitive")

	var keys []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &keys))
	require.Len(t, keys, 1)
	assert.Equal(t, sentinelPrivateKey, keys[0]["private_key"])
}

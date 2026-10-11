package application

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type fakeRequest struct {
	Method     string
	RequestURI string
	Body       string
}

type fakeAPI struct {
	mu       sync.Mutex
	requests []fakeRequest
}

func (a *fakeAPI) recorded() []fakeRequest {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]fakeRequest(nil), a.requests...)
}

// newFakeAPI starts a fake Coolify API and points a temporary CLI config at it.
func newFakeAPI(t *testing.T, handler http.HandlerFunc) *fakeAPI {
	t.Helper()
	fake := &fakeAPI{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fake.mu.Lock()
		fake.requests = append(fake.requests, fakeRequest{Method: r.Method, RequestURI: r.RequestURI, Body: string(body)})
		fake.mu.Unlock()
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "coolify")
	require.NoError(t, os.MkdirAll(configDir, 0o750))
	data, err := json.Marshal(map[string]any{
		"instances": []map[string]any{{"name": "test", "fqdn": server.URL, "token": "test-token", "default": true}},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), data, 0o600))
	return fake
}

// runCmd executes the command tree under a minimal root carrying the global flags.
func runCmd(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	root := &cobra.Command{Use: "coolify", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("token", "", "")
	root.PersistentFlags().String("context", "", "")
	root.PersistentFlags().String("format", "table", "")
	root.PersistentFlags().BoolP("show-sensitive", "s", false, "")
	root.PersistentFlags().Bool("debug", false, "")
	root.AddCommand(NewAppCommand())

	var stdout, stderr bytes.Buffer
	root.SetIn(strings.NewReader(stdin))
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

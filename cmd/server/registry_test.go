package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedRequest struct {
	Method     string
	RequestURI string
	Body       string
}

type registryTestAPI struct {
	mu       sync.Mutex
	requests []recordedRequest
	server   *httptest.Server
}

func (a *registryTestAPI) recorded() []recordedRequest {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]recordedRequest(nil), a.requests...)
}

// newRegistryTestAPI starts a fake Coolify API and points a temporary CLI config at it.
func newRegistryTestAPI(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *registryTestAPI {
	t.Helper()
	fake := &registryTestAPI{}
	fake.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fake.mu.Lock()
		fake.requests = append(fake.requests, recordedRequest{Method: r.Method, RequestURI: r.RequestURI, Body: string(body)})
		fake.mu.Unlock()
		r.Body = io.NopCloser(bytes.NewReader(body))
		handler(w, r)
	}))
	t.Cleanup(fake.server.Close)

	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := filepath.Join(home, ".config", "coolify")
	require.NoError(t, os.MkdirAll(configDir, 0o750))
	cfg := map[string]any{
		"instances": []map[string]any{{
			"name": "test", "fqdn": fake.server.URL, "token": "test-token", "default": true,
		}},
	}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), data, 0o600))

	return fake
}

// runRegistryCommand runs `coolify server registry <args>` against the fake API.
func runRegistryCommand(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	root := &cobra.Command{Use: "coolify", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("token", "", "")
	root.PersistentFlags().String("context", "", "")
	root.PersistentFlags().String("format", "table", "")
	root.PersistentFlags().BoolP("show-sensitive", "s", false, "")
	root.PersistentFlags().Bool("debug", false, "")
	root.AddCommand(NewServerCommand())

	var stdout, stderr bytes.Buffer
	root.SetIn(strings.NewReader(stdin))
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(append([]string{"server", "registry"}, args...))

	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

const registryListFixture = `{
	"registries": [
		{"registry": "ghcr.io", "logged_in": true, "source": "auths", "username": "octocat", "used_by": []},
		{"registry": "gcr.io", "logged_in": true, "source": "credHelpers", "username": null, "used_by": [
			{"type": "Service", "name": "svc", "link": null}
		]},
		{"registry": "registry.example.com:5000", "logged_in": false, "source": null, "username": null, "used_by": [
			{"type": "Application", "name": "app-1", "link": "https://coolify.test/a1"},
			{"type": "Application", "name": "app-2", "link": "https://coolify.test/a2"},
			{"type": "Application", "name": "app-3", "link": "https://coolify.test/a3"},
			{"type": "Database", "name": "db-1", "link": "https://coolify.test/d1"}
		]}
	],
	"error": null
}`

func TestRegistryList_TableOutput(t *testing.T) {
	fake := newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(registryListFixture))
	})

	stdout, stderr, err := runRegistryCommand(t, "", "list", "srv-1")
	require.NoError(t, err)
	assert.Empty(t, stderr)

	requests := fake.recorded()
	require.Len(t, requests, 1)
	assert.Equal(t, http.MethodGet, requests[0].Method)
	assert.Equal(t, "/api/v1/servers/srv-1/registries", requests[0].RequestURI)

	for _, want := range []string{
		"registry", "status", "username", "used_by",
		"ghcr.io", "Logged in",
		"gcr.io", "Credential helper", "1 service",
		"registry.example.com:5000", "Not logged in", "3 applications, 1 database",
		"Note: Use -s to show usernames.",
	} {
		assert.Contains(t, stdout, want)
	}
	assert.NotContains(t, stdout, "octocat", "usernames are masked without -s")
	assert.Contains(t, stdout, "********")
	assert.NotContains(t, stdout, "app-1", "table shows only the usage summary")
}

func TestRegistryList_ShowSensitive(t *testing.T) {
	newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(registryListFixture))
	})

	stdout, _, err := runRegistryCommand(t, "", "list", "srv-1", "-s")
	require.NoError(t, err)
	assert.Contains(t, stdout, "octocat")
	assert.NotContains(t, stdout, "********")
	assert.NotContains(t, stdout, "Note: Use -s")
}

func TestRegistryList_JSONOutputIncludesFullUsage(t *testing.T) {
	newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(registryListFixture))
	})

	stdout, _, err := runRegistryCommand(t, "", "list", "srv-1", "--format", "json")
	require.NoError(t, err)

	var decoded struct {
		Registries []struct {
			Registry string  `json:"registry"`
			Username *string `json:"username"`
			UsedBy   []struct {
				Type string `json:"type"`
				Name string `json:"name"`
				Link string `json:"link"`
			} `json:"used_by"`
		} `json:"registries"`
		Error *string `json:"error"`
	}
	require.NoError(t, json.Unmarshal([]byte(stdout), &decoded))
	require.Len(t, decoded.Registries, 3)
	require.Len(t, decoded.Registries[2].UsedBy, 4)
	assert.Equal(t, "app-1", decoded.Registries[2].UsedBy[0].Name)
	assert.Equal(t, "https://coolify.test/a1", decoded.Registries[2].UsedBy[0].Link)
	assert.Nil(t, decoded.Error)
}

func TestRegistryList_PrintsErrorFieldAsWarning(t *testing.T) {
	newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"registries": [], "error": "The server is not reachable. Validate the server to read its registry logins."}`))
	})

	stdout, stderr, err := runRegistryCommand(t, "", "list", "srv-1")
	require.NoError(t, err)
	assert.Equal(t, "Warning: The server is not reachable. Validate the server to read its registry logins.\n", stderr)
	assert.Contains(t, stdout, "No data")
}

func TestRegistryList_SurfacesForbiddenMessage(t *testing.T) {
	newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Missing required permissions: read:sensitive"}`))
	})

	_, _, err := runRegistryCommand(t, "", "list", "srv-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list registry logins")
	assert.Contains(t, err.Error(), "Missing required permissions: read:sensitive")
}

func TestRegistryLogin_PasswordFromStdin(t *testing.T) {
	const secret = "ghp_SECRET'token\"with$special"
	fake := newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message":"Logged in to ghcr.io."}`))
	})

	stdout, _, err := runRegistryCommand(t, secret+"\n",
		"login", "srv-1", "--registry", "https://GHCR.io/", "--username", "octocat", "--password-stdin")
	require.NoError(t, err)
	assert.Equal(t, "Logged in to ghcr.io.\n", stdout)

	requests := fake.recorded()
	require.Len(t, requests, 1)
	assert.Equal(t, http.MethodPost, requests[0].Method)
	assert.Equal(t, "/api/v1/servers/srv-1/registries", requests[0].RequestURI)
	assert.NotContains(t, requests[0].RequestURI, "SECRET")

	var payload map[string]string
	require.NoError(t, json.Unmarshal([]byte(requests[0].Body), &payload))
	assert.Equal(t, map[string]string{"registry": "ghcr.io", "username": "octocat", "password": secret}, payload)
}

func TestRegistryLogin_StripsOnlyTrailingNewline(t *testing.T) {
	fake := newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	})

	_, _, err := runRegistryCommand(t, " pass word \r\n",
		"login", "srv-1", "--registry", "ghcr.io", "--username", "octocat", "--password-stdin")
	require.NoError(t, err)

	var payload map[string]string
	require.NoError(t, json.Unmarshal([]byte(fake.recorded()[0].Body), &payload))
	assert.Equal(t, " pass word ", payload["password"])
}

func TestRegistryLogin_HasNoPasswordFlag(t *testing.T) {
	cmd := newRegistryLoginCommand()
	assert.Nil(t, cmd.Flags().Lookup("password"))
	assert.Nil(t, cmd.Flags().ShorthandLookup("p"))
	require.NotNil(t, cmd.Flags().Lookup("password-stdin"))
}

func TestRegistryLogin_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		stdin   string
		args    []string
		wantErr string
	}{
		{name: "missing registry", stdin: "x", args: []string{"login", "srv-1", "--username", "u", "--password-stdin"}, wantErr: "--registry is required"},
		{name: "missing username", stdin: "x", args: []string{"login", "srv-1", "--registry", "ghcr.io", "--password-stdin"}, wantErr: "--username is required"},
		{name: "empty stdin", stdin: "\n", args: []string{"login", "srv-1", "--registry", "ghcr.io", "--username", "u", "--password-stdin"}, wantErr: "password must not be empty"},
		{name: "too long", stdin: strings.Repeat("a", maxRegistryPasswordBytes+1), args: []string{"login", "srv-1", "--registry", "ghcr.io", "--username", "u", "--password-stdin"}, wantErr: "must not be longer than"},
		{name: "no tty and no --password-stdin", stdin: "secret", args: []string{"login", "srv-1", "--registry", "ghcr.io", "--username", "u"}, wantErr: "--password-stdin"},
		{name: "missing server", stdin: "x", args: []string{"login", "--registry", "ghcr.io", "--username", "u", "--password-stdin"}, wantErr: "missing required arguments"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})

			_, _, err := runRegistryCommand(t, tt.stdin, tt.args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Empty(t, fake.recorded(), "no request is sent when input is invalid")
		})
	}
}

func TestRegistryLogin_PromptsWithoutEchoOnTTY(t *testing.T) {
	previousFD, previousRead := terminalFD, readHiddenPassword
	t.Cleanup(func() { terminalFD, readHiddenPassword = previousFD, previousRead })
	terminalFD = func(io.Reader) (int, bool) { return 42, true }
	var promptedFD int
	readHiddenPassword = func(fd int) ([]byte, error) {
		promptedFD = fd
		return []byte("typed-secret"), nil
	}

	fake := newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message":"Logged in to ghcr.io."}`))
	})

	stdout, stderr, err := runRegistryCommand(t, "", "login", "srv-1", "--registry", "ghcr.io", "-u", "octocat")
	require.NoError(t, err)
	assert.Equal(t, 42, promptedFD)
	assert.Contains(t, stderr, "Password: ")
	assert.NotContains(t, stdout+stderr, "typed-secret")

	var payload map[string]string
	require.NoError(t, json.Unmarshal([]byte(fake.recorded()[0].Body), &payload))
	assert.Equal(t, "typed-secret", payload["password"])
}

func TestRegistryLogin_PromptReadError(t *testing.T) {
	previousFD, previousRead := terminalFD, readHiddenPassword
	t.Cleanup(func() { terminalFD, readHiddenPassword = previousFD, previousRead })
	terminalFD = func(io.Reader) (int, bool) { return 3, true }
	readHiddenPassword = func(int) ([]byte, error) { return nil, errors.New("interrupted") }

	newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, _, err := runRegistryCommand(t, "", "login", "srv-1", "--registry", "ghcr.io", "-u", "octocat")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read password: interrupted")
}

func TestRegistryLogin_SurfacesAPIErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{name: "docker failure", status: http.StatusBadRequest, body: `{"message":"Login failed: Error response from daemon: unauthorized: bad token ***"}`, want: "Login failed: Error response from daemon: unauthorized: bad token ***"},
		{name: "validation", status: http.StatusUnprocessableEntity, body: `{"message":"Validation failed.","errors":{"username":["The username must not contain spaces."]}}`, want: "API error 422"},
		{name: "not found", status: http.StatusNotFound, body: `{"message":"Server not found."}`, want: "Server not found."},
		{name: "forbidden", status: http.StatusForbidden, body: `{"message":"Missing required permissions: write"}`, want: "Missing required permissions: write"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})

			_, _, err := runRegistryCommand(t, "secret", "login", "srv-1", "--registry", "ghcr.io", "--username", "octocat", "--password-stdin")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to log in to ghcr.io")
			assert.Contains(t, err.Error(), tt.want)
			assert.NotContains(t, err.Error(), "secret")
		})
	}
}

func TestRegistryLogin_MultipleServers(t *testing.T) {
	fake := newRegistryTestAPI(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/srv-bad/") {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"The server is not reachable. Validate the server first."}`))
			return
		}
		_, _ = w.Write([]byte(`{"message":"Logged in to ghcr.io."}`))
	})

	stdout, stderr, err := runRegistryCommand(t, "secret\n",
		"login", "srv-1", "srv-bad", "srv-2", "--registry", "ghcr.io", "--username", "octocat", "--password-stdin")
	require.Error(t, err)
	assert.Equal(t, "login to ghcr.io failed on 1 of 3 servers", err.Error())
	assert.Equal(t, "srv-1: Logged in to ghcr.io.\nsrv-2: Logged in to ghcr.io.\n", stdout)
	assert.Contains(t, stderr, "srv-bad: failed to log in to ghcr.io")
	assert.Contains(t, stderr, "The server is not reachable. Validate the server first.")

	requests := fake.recorded()
	require.Len(t, requests, 3)
	for _, req := range requests {
		assert.Contains(t, req.Body, `"password":"secret"`, "stdin is read once and reused for every server")
	}
}

func TestRegistryCheck_EscapesRegistryWithPort(t *testing.T) {
	fake := newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message":"The login for registry.example.com:5000 works."}`))
	})

	stdout, _, err := runRegistryCommand(t, "", "check", "srv-1", "registry.example.com:5000")
	require.NoError(t, err)
	assert.Equal(t, "The login for registry.example.com:5000 works.\n", stdout)

	requests := fake.recorded()
	require.Len(t, requests, 1)
	assert.Equal(t, http.MethodPost, requests[0].Method)
	assert.Equal(t, "/api/v1/servers/srv-1/registries/registry.example.com:5000/check", requests[0].RequestURI)
}

func TestRegistryCheck_SurfacesFailure(t *testing.T) {
	newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Login check failed: unauthorized"}`))
	})

	_, _, err := runRegistryCommand(t, "", "check", "srv-1", "ghcr.io")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check login for ghcr.io")
	assert.Contains(t, err.Error(), "Login check failed: unauthorized")
}

func TestRegistryLogout_WithForce(t *testing.T) {
	fake := newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message":"Logged out from registry.example.com:5000."}`))
	})

	stdout, _, err := runRegistryCommand(t, "", "logout", "srv-1", "registry.example.com:5000", "--force")
	require.NoError(t, err)
	assert.Equal(t, "Logged out from registry.example.com:5000 on server srv-1.\n", stdout)
	assert.NotContains(t, stdout, "Are you sure")

	requests := fake.recorded()
	require.Len(t, requests, 1)
	assert.Equal(t, http.MethodDelete, requests[0].Method)
	assert.Equal(t, "/api/v1/servers/srv-1/registries/registry.example.com:5000", requests[0].RequestURI)
}

func TestRegistryLogout_Confirmation(t *testing.T) {
	tests := []struct {
		name        string
		answer      string
		wantRequest bool
		wantOutput  string
	}{
		{name: "yes", answer: "yes\n", wantRequest: true, wantOutput: "Logged out from ghcr.io on server srv-1."},
		{name: "y", answer: "y\n", wantRequest: true, wantOutput: "Logged out from ghcr.io on server srv-1."},
		{name: "no", answer: "no\n", wantRequest: false, wantOutput: "Logout cancelled."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"message":"Logged out from ghcr.io."}`))
			})

			stdout, _, err := runRegistryCommand(t, tt.answer, "logout", "srv-1", "ghcr.io")
			require.NoError(t, err)
			assert.Contains(t, stdout, "Are you sure you want to log out from ghcr.io on server srv-1? (yes/no): ")
			assert.Contains(t, stdout, tt.wantOutput)
			if tt.wantRequest {
				assert.Len(t, fake.recorded(), 1)
			} else {
				assert.Empty(t, fake.recorded())
			}
		})
	}
}

func TestRegistryLogout_SurfacesFailure(t *testing.T) {
	newRegistryTestAPI(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"The server is not reachable. Validate the server first."}`))
	})

	_, _, err := runRegistryCommand(t, "", "logout", "srv-1", "ghcr.io", "-f")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to log out from ghcr.io")
	assert.Contains(t, err.Error(), "The server is not reachable. Validate the server first.")
}

func TestRegistryCommand_Registered(t *testing.T) {
	serverCmd := NewServerCommand()
	registryCmd, _, err := serverCmd.Find([]string{"registries", "list"})
	require.NoError(t, err)
	assert.Equal(t, "list", registryCmd.Name())
	assert.Equal(t, "registry", registryCmd.Parent().Name())
}

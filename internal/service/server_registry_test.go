package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func newRegistryTestService(t *testing.T, handler http.HandlerFunc) *ServerRegistryService {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewServerRegistryService(api.NewClient(server.URL, "test-token", api.WithRetries(0)))
}

func TestServerRegistryService_List(t *testing.T) {
	svc := newRegistryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/servers/srv-1/registries", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{
			"registries": [
				{"registry": "ghcr.io", "logged_in": true, "source": "auths", "username": "octocat", "used_by": []},
				{"registry": "registry.example.com:5000", "logged_in": false, "source": null, "username": null,
				 "used_by": [{"type": "Application", "name": "private-app", "link": "https://coolify.test/app"}]}
			],
			"error": null
		}`))
	})

	result, err := svc.List(context.Background(), "srv-1")
	require.NoError(t, err)
	require.Len(t, result.Registries, 2)
	assert.Nil(t, result.Error)

	ghcr := result.Registries[0]
	assert.Equal(t, "ghcr.io", ghcr.Registry)
	assert.True(t, ghcr.LoggedIn)
	require.NotNil(t, ghcr.Source)
	assert.Equal(t, models.ServerRegistrySourceAuths, *ghcr.Source)
	require.NotNil(t, ghcr.Username)
	assert.Equal(t, "octocat", *ghcr.Username)
	assert.Empty(t, ghcr.UsedBy)

	private := result.Registries[1]
	assert.False(t, private.LoggedIn)
	assert.Nil(t, private.Source)
	assert.Nil(t, private.Username)
	require.Len(t, private.UsedBy, 1)
	assert.Equal(t, "Application", private.UsedBy[0].Type)
	assert.Equal(t, "private-app", private.UsedBy[0].Name)
	require.NotNil(t, private.UsedBy[0].Link)
	assert.Equal(t, "https://coolify.test/app", *private.UsedBy[0].Link)
}

func TestServerRegistryService_List_ErrorField(t *testing.T) {
	svc := newRegistryTestService(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"registries": [], "error": "The server is not reachable. Validate the server to read its registry logins."}`))
	})

	result, err := svc.List(context.Background(), "srv-1")
	require.NoError(t, err)
	assert.Empty(t, result.Registries)
	require.NotNil(t, result.Error)
	assert.Contains(t, *result.Error, "not reachable")
}

func TestServerRegistryService_Login_SendsPasswordOnlyInBody(t *testing.T) {
	const secret = "ghp_SECRET'token\"with$special"
	svc := newRegistryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/servers/srv-1/registries", r.URL.Path)
		assert.NotContains(t, r.RequestURI, "SECRET")
		assert.Empty(t, r.URL.RawQuery)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		var payload map[string]any
		assert.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, map[string]any{"registry": "ghcr.io", "username": "octocat", "password": secret}, payload)

		_, _ = w.Write([]byte(`{"message":"Logged in to ghcr.io."}`))
	})

	resp, err := svc.Login(context.Background(), "srv-1", models.ServerRegistryLoginRequest{
		Registry: "https://GHCR.io/",
		Username: "octocat",
		Password: secret,
	})
	require.NoError(t, err)
	assert.Equal(t, "Logged in to ghcr.io.", resp.Message)
}

func TestServerRegistryService_Check_EscapesRegistryWithPort(t *testing.T) {
	svc := newRegistryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/servers/srv-1/registries/registry.example.com:5000/check", r.URL.Path)
		assert.Equal(t, "/api/v1/servers/srv-1/registries/registry.example.com:5000/check", r.RequestURI)
		_, _ = w.Write([]byte(`{"message":"The login for registry.example.com:5000 works."}`))
	})

	resp, err := svc.Check(context.Background(), "srv-1", "registry.example.com:5000")
	require.NoError(t, err)
	assert.Equal(t, "The login for registry.example.com:5000 works.", resp.Message)
}

func TestServerRegistryService_Logout(t *testing.T) {
	svc := newRegistryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/servers/srv-1/registries/ghcr.io", r.RequestURI)
		_, _ = w.Write([]byte(`{"message":"Logged out from ghcr.io."}`))
	})

	require.NoError(t, svc.Logout(context.Background(), "srv-1", "GHCR.io"))
}

func TestServerRegistryService_PathSegmentsAreEscaped(t *testing.T) {
	svc := newRegistryTestService(t, func(w http.ResponseWriter, r *http.Request) {
		// Characters that would change the request target stay inside one path segment.
		assert.Equal(t, "/api/v1/servers/srv%3F1/registries/bad%20host%3Fx=1%23frag", r.RequestURI)
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"Validation failed.","errors":{"registry":["Enter a registry host."]}}`))
	})

	err := svc.Logout(context.Background(), "srv?1", "bad host?x=1#frag")
	require.Error(t, err)
}

func TestServerRegistryService_ErrorsSurfaceAPIMessage(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		call    func(*ServerRegistryService) error
		message string
	}{
		{
			name:   "login docker failure",
			status: http.StatusBadRequest,
			body:   `{"message":"Login failed: Error response from daemon: unauthorized: bad token ***"}`,
			call: func(s *ServerRegistryService) error {
				_, err := s.Login(context.Background(), "srv-1", models.ServerRegistryLoginRequest{Registry: "ghcr.io", Username: "octocat", Password: "x"})
				return err
			},
			message: "Login failed: Error response from daemon: unauthorized: bad token ***",
		},
		{
			name:   "login validation",
			status: http.StatusUnprocessableEntity,
			body:   `{"message":"Validation failed.","errors":{"username":["The username must not contain spaces."]}}`,
			call: func(s *ServerRegistryService) error {
				_, err := s.Login(context.Background(), "srv-1", models.ServerRegistryLoginRequest{Registry: "ghcr.io", Username: "octo cat", Password: "x"})
				return err
			},
			message: "Validation failed.",
		},
		{
			name:   "check failure",
			status: http.StatusBadRequest,
			body:   `{"message":"Login check failed: Error: Cannot perform an interactive login from a non TTY device"}`,
			call: func(s *ServerRegistryService) error {
				_, err := s.Check(context.Background(), "srv-1", "ghcr.io")
				return err
			},
			message: "Login check failed: Error: Cannot perform an interactive login from a non TTY device",
		},
		{
			name:   "list forbidden",
			status: http.StatusForbidden,
			body:   `{"message":"Missing required permissions: read:sensitive"}`,
			call: func(s *ServerRegistryService) error {
				_, err := s.List(context.Background(), "srv-1")
				return err
			},
			message: "Missing required permissions: read:sensitive",
		},
		{
			name:   "logout server not found",
			status: http.StatusNotFound,
			body:   `{"message":"Server not found."}`,
			call: func(s *ServerRegistryService) error {
				return s.Logout(context.Background(), "srv-1", "ghcr.io")
			},
			message: "Server not found.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newRegistryTestService(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})

			err := tt.call(svc)
			require.Error(t, err)
			var apiErr *api.Error
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, tt.status, apiErr.StatusCode)
			assert.Equal(t, tt.message, apiErr.Message)
		})
	}
}

func TestNormalizeRegistry(t *testing.T) {
	tests := map[string]string{
		"ghcr.io":                            "ghcr.io",
		"  GHCR.io  ":                        "ghcr.io",
		"https://GHCR.io/":                   "ghcr.io",
		"http://registry.example.com:5000/":  "registry.example.com:5000",
		"registry.example.com:5000/team/app": "registry.example.com:5000",
		"docker.io":                          "docker.io",
		"":                                   "",
	}
	for input, want := range tests {
		assert.Equal(t, want, NormalizeRegistry(input), "input %q", input)
	}
}

func TestServerRegistryStatus(t *testing.T) {
	assert.Equal(t, models.ServerRegistryStatusLoggedIn, ServerRegistryStatus(models.ServerRegistry{LoggedIn: true, Source: strPtr("auths")}))
	assert.Equal(t, models.ServerRegistryStatusLoggedIn, ServerRegistryStatus(models.ServerRegistry{LoggedIn: true}))
	assert.Equal(t, models.ServerRegistryStatusCredentialHelper, ServerRegistryStatus(models.ServerRegistry{LoggedIn: true, Source: strPtr("credHelpers")}))
	assert.Equal(t, models.ServerRegistryStatusNotLoggedIn, ServerRegistryStatus(models.ServerRegistry{LoggedIn: false}))
}

func TestServerRegistryUsageSummary(t *testing.T) {
	assert.Empty(t, ServerRegistryUsageSummary(nil))
	assert.Equal(t, "3 applications, 1 database", ServerRegistryUsageSummary([]models.ServerRegistryUser{
		{Type: "Application", Name: "a"},
		{Type: "Database", Name: "db"},
		{Type: "Application", Name: "b"},
		{Type: "Application", Name: "c"},
	}))
	assert.Equal(t, "1 service, 2 builds", ServerRegistryUsageSummary([]models.ServerRegistryUser{
		{Type: "Service", Name: "s"},
		{Type: "Build", Name: "a"},
		{Type: "Build", Name: "b"},
	}))
}

func TestServerRegistryRows_MasksUsernames(t *testing.T) {
	registries := []models.ServerRegistry{
		{Registry: "ghcr.io", LoggedIn: true, Source: strPtr("auths"), Username: strPtr("octocat")},
		{Registry: "registry.example.com:5000", UsedBy: []models.ServerRegistryUser{{Type: "Application", Name: "app"}}},
	}

	masked := ServerRegistryRows(registries, false, "********")
	require.Len(t, masked, 2)
	assert.Equal(t, models.ServerRegistryRow{Registry: "ghcr.io", Status: "Logged in", Username: "********"}, masked[0])
	assert.Equal(t, models.ServerRegistryRow{Registry: "registry.example.com:5000", Status: "Not logged in", UsedBy: "1 application"}, masked[1])

	shown := ServerRegistryRows(registries, true, "********")
	assert.Equal(t, "octocat", shown[0].Username)
	assert.Empty(t, shown[1].Username)
}

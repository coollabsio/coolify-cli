package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

const githubAppsListFixture = `[
	{"id": 7, "uuid": "gh-app-uuid-7", "name": "public-github", "html_url": "https://github.com"},
	{"id": 42, "uuid": "gh-app-uuid-42", "name": "my-github-app", "html_url": "https://github.com"}
]`

// newGitHubTestServer records every "METHOD path" it serves. GET /github-apps
// returns the list fixture; any other request matching a key in routes returns
// that body; everything else is a 404.
func newGitHubTestServer(t *testing.T, routes map[string]string) (*httptest.Server, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var calls []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		mu.Lock()
		calls = append(calls, key)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		if key == "GET /api/v1/github-apps" {
			_, _ = w.Write([]byte(githubAppsListFixture))
			return
		}
		if body, ok := routes[key]; ok {
			if r.Method == http.MethodDelete && body == "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			_, _ = w.Write([]byte(body))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not found."}`))
	}))
	t.Cleanup(srv.Close)

	return srv, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), calls...)
	}
}

func assertGitHubCalls(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("calls = %v, want %v", got, want)
	}
}

func TestGitHubAppService_Get_ResolvesFromList(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantID int
	}{
		{name: "by uuid", input: "gh-app-uuid-42", wantID: 42},
		{name: "by numeric id", input: "7", wantID: 7},
		{name: "trims whitespace", input: "  gh-app-uuid-42 ", wantID: 42},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, calls := newGitHubTestServer(t, nil)
			svc := NewGitHubAppService(api.NewClient(srv.URL, "test-token"))

			app, err := svc.Get(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if app.ID != tt.wantID {
				t.Errorf("Get() ID = %d, want %d", app.ID, tt.wantID)
			}
			// Must never hit the nonexistent GET /github-apps/{x} endpoint.
			assertGitHubCalls(t, calls(), []string{"GET /api/v1/github-apps"})
		})
	}
}

func TestGitHubAppService_Get_UnknownUUID(t *testing.T) {
	srv, calls := newGitHubTestServer(t, nil)
	svc := NewGitHubAppService(api.NewClient(srv.URL, "test-token"))

	_, err := svc.Get(context.Background(), "does-not-exist")
	if err == nil {
		t.Fatal("Get() expected error for unknown UUID")
	}
	if !strings.Contains(err.Error(), `GitHub App "does-not-exist" not found`) {
		t.Errorf("Get() error = %q, want not-found message", err)
	}
	assertGitHubCalls(t, calls(), []string{"GET /api/v1/github-apps"})
}

func TestGitHubAppService_Update_UsesNumericID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCalls []string
	}{
		{
			name:      "uuid resolves via list",
			input:     "gh-app-uuid-42",
			wantCalls: []string{"GET /api/v1/github-apps", "PATCH /api/v1/github-apps/42"},
		},
		{
			name:      "numeric id skips list",
			input:     "42",
			wantCalls: []string{"PATCH /api/v1/github-apps/42"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, calls := newGitHubTestServer(t, map[string]string{
				"PATCH /api/v1/github-apps/42": `{"message":"GitHub app updated successfully"}`,
			})
			svc := NewGitHubAppService(api.NewClient(srv.URL, "test-token"))

			name := "renamed"
			if err := svc.Update(context.Background(), tt.input, &models.GitHubAppUpdateRequest{Name: &name}); err != nil {
				t.Fatalf("Update() error = %v", err)
			}
			assertGitHubCalls(t, calls(), tt.wantCalls)
		})
	}
}

func TestGitHubAppService_Update_SendsBody(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
	defer srv.Close()
	svc := NewGitHubAppService(api.NewClient(srv.URL, "test-token"))

	name := "renamed"
	if err := svc.Update(context.Background(), "42", &models.GitHubAppUpdateRequest{Name: &name}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if gotPath != "PATCH /api/v1/github-apps/42" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["name"] != "renamed" {
		t.Errorf("body = %v, want name=renamed", gotBody)
	}
}

func TestGitHubAppService_Delete_UsesNumericID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCalls []string
	}{
		{
			name:      "uuid resolves via list",
			input:     "gh-app-uuid-42",
			wantCalls: []string{"GET /api/v1/github-apps", "DELETE /api/v1/github-apps/42"},
		},
		{
			name:      "numeric id skips list",
			input:     "42",
			wantCalls: []string{"DELETE /api/v1/github-apps/42"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, calls := newGitHubTestServer(t, map[string]string{
				"DELETE /api/v1/github-apps/42": "",
			})
			svc := NewGitHubAppService(api.NewClient(srv.URL, "test-token"))

			if err := svc.Delete(context.Background(), tt.input); err != nil {
				t.Fatalf("Delete() error = %v", err)
			}
			assertGitHubCalls(t, calls(), tt.wantCalls)
		})
	}
}

func TestGitHubAppService_ListRepositories_UsesNumericID(t *testing.T) {
	srv, calls := newGitHubTestServer(t, map[string]string{
		"GET /api/v1/github-apps/42/repositories": `{"repositories":[{"id":1,"name":"repo","full_name":"owner/repo"}]}`,
	})
	svc := NewGitHubAppService(api.NewClient(srv.URL, "test-token"))

	repos, err := svc.ListRepositories(context.Background(), "gh-app-uuid-42")
	if err != nil {
		t.Fatalf("ListRepositories() error = %v", err)
	}
	if len(repos) != 1 {
		t.Errorf("ListRepositories() len = %d, want 1", len(repos))
	}
	assertGitHubCalls(t, calls(), []string{"GET /api/v1/github-apps", "GET /api/v1/github-apps/42/repositories"})
}

func TestGitHubAppService_ListBranches_UsesNumericID(t *testing.T) {
	srv, calls := newGitHubTestServer(t, map[string]string{
		"GET /api/v1/github-apps/42/repositories/owner/repo/branches": `{"branches":[{"name":"main"},{"name":"dev"}]}`,
	})
	svc := NewGitHubAppService(api.NewClient(srv.URL, "test-token"))

	branches, err := svc.ListBranches(context.Background(), "gh-app-uuid-42", "owner", "repo")
	if err != nil {
		t.Fatalf("ListBranches() error = %v", err)
	}
	if len(branches) != 2 {
		t.Errorf("ListBranches() len = %d, want 2", len(branches))
	}
	assertGitHubCalls(t, calls(), []string{"GET /api/v1/github-apps", "GET /api/v1/github-apps/42/repositories/owner/repo/branches"})
}

func TestGitHubAppService_UnknownUUID_DoesNotCallEndpoint(t *testing.T) {
	ctx := context.Background()
	ops := map[string]func(*GitHubAppService) error{
		"update": func(s *GitHubAppService) error {
			return s.Update(ctx, "missing-uuid", &models.GitHubAppUpdateRequest{})
		},
		"delete": func(s *GitHubAppService) error { return s.Delete(ctx, "missing-uuid") },
		"repos": func(s *GitHubAppService) error {
			_, err := s.ListRepositories(ctx, "missing-uuid")
			return err
		},
		"branches": func(s *GitHubAppService) error {
			_, err := s.ListBranches(ctx, "missing-uuid", "owner", "repo")
			return err
		},
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			srv, calls := newGitHubTestServer(t, nil)
			svc := NewGitHubAppService(api.NewClient(srv.URL, "test-token"))

			err := op(svc)
			if err == nil || !strings.Contains(err.Error(), `GitHub App "missing-uuid" not found`) {
				t.Fatalf("error = %v, want not-found", err)
			}
			assertGitHubCalls(t, calls(), []string{"GET /api/v1/github-apps"})
		})
	}
}

func TestGitHubAppService_ListError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthenticated."}`))
	}))
	defer srv.Close()
	svc := NewGitHubAppService(api.NewClient(srv.URL, "test-token"))

	if _, err := svc.Get(context.Background(), "gh-app-uuid-42"); err == nil {
		t.Fatal("Get() expected error when list fails")
	}
}

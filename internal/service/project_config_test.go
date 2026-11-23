package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

type recordedRequest struct {
	method string
	uri    string
	body   string
}

func newProjectConfigTestService(t *testing.T, status int, response string) (*ProjectConfigService, *[]recordedRequest) {
	t.Helper()
	var requests []recordedRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests = append(requests, recordedRequest{method: r.Method, uri: r.RequestURI, body: string(body)})
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	return NewProjectConfigService(api.NewClient(server.URL, "test-token", api.WithRetries(0))), &requests
}

func TestProjectConfigService_Export_ReturnsFileAsSent(t *testing.T) {
	svc, reqs := newProjectConfigTestService(t, http.StatusOK, `{"version":1,"resources":{"api":{"type":"application"}}}`)

	file, err := svc.Export(context.Background(), "env-uuid")

	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1,"resources":{"api":{"type":"application"}}}`, string(file))
	require.Len(t, *reqs, 1)
	assert.Equal(t, "GET", (*reqs)[0].method)
	assert.Equal(t, "/api/v1/environments/env-uuid/config/export", (*reqs)[0].uri)
}

func TestProjectConfigService_Export_Forbidden(t *testing.T) {
	svc, _ := newProjectConfigTestService(t, http.StatusForbidden, `{"message":"Exporting needs the read:sensitive ability."}`)
	_, err := svc.Export(context.Background(), "env-uuid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read:sensitive")
}

func TestProjectConfigService_Plan(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		response    string
		wantErr     bool
		wantChanges int
		wantErrors  []models.ConfigError
		wantMissing models.ConfigMissingMap
	}{
		{
			name:   "plan with empty missing_required array",
			status: http.StatusOK,
			response: `{"changes":[{"kind":"create","key":"api","path":"","destructive":false,"secret":false,"from":null,"to":{"type":"application"}}],
				"errors":[],"warnings":["w"],"missing_required":[],"state_hash":"h","destructive":false,"removes_all":false}`,
			wantChanges: 1,
		},
		{
			name:        "plan with missing required values",
			status:      http.StatusOK,
			response:    `{"changes":[],"errors":[],"warnings":[],"missing_required":{"api":["STRIPE_KEY"]},"destructive":false,"removes_all":false}`,
			wantMissing: models.ConfigMissingMap{"api": {"STRIPE_KEY"}},
		},
		{
			name:       "file errors become a plan with errors",
			status:     http.StatusUnprocessableEntity,
			response:   `{"message":"The file is not valid.","errors":[{"path":"resources.api.type","message":"Unknown type."}]}`,
			wantErrors: []models.ConfigError{{Path: "resources.api.type", Message: "Unknown type."}},
		},
		{
			name:     "request validation errors stay an error",
			status:   http.StatusUnprocessableEntity,
			response: `{"message":"Validation failed.","errors":{"file":["The file field must be present."]}}`,
			wantErr:  true,
		},
		{
			name:     "not found",
			status:   http.StatusNotFound,
			response: `{"message":"Environment not found."}`,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, reqs := newProjectConfigTestService(t, tt.status, tt.response)

			plan, err := svc.Plan(context.Background(), "env-uuid", `{"version":1}`)

			require.Len(t, *reqs, 1)
			assert.Equal(t, "POST", (*reqs)[0].method)
			assert.Equal(t, "/api/v1/environments/env-uuid/config/plan", (*reqs)[0].uri)
			assert.JSONEq(t, `{"file":"{\"version\":1}"}`, (*reqs)[0].body)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, plan.Changes, tt.wantChanges)
			if tt.wantErrors == nil {
				assert.Empty(t, plan.Errors)
			} else {
				assert.Equal(t, tt.wantErrors, plan.Errors)
			}
			if tt.wantMissing == nil {
				assert.Empty(t, plan.MissingRequired)
			} else {
				assert.Equal(t, tt.wantMissing, plan.MissingRequired)
			}
		})
	}
}

func TestProjectConfigService_Apply(t *testing.T) {
	file := `{"version":1}`
	tests := []struct {
		name       string
		file       *string
		status     int
		response   string
		wantBody   string
		wantStatus string
		wantErr    bool
	}{
		{name: "api link applies file", file: &file, status: http.StatusAccepted, response: `{"uuid":"rev-1","status":"applied"}`, wantBody: `{"file":"{\"version\":1}"}`, wantStatus: "applied"},
		{name: "git link sends no body", status: http.StatusAccepted, response: `{"uuid":"rev-2","status":"planned"}`, wantStatus: "planned"},
		{name: "waiting approval", file: &file, status: http.StatusOK, response: `{"uuid":"rev-3","status":"waiting_approval"}`, wantBody: `{"file":"{\"version\":1}"}`, wantStatus: "waiting_approval"},
		{name: "failed revision is returned", file: &file, status: http.StatusUnprocessableEntity, response: `{"uuid":"rev-4","status":"failed","error":"boom"}`, wantBody: `{"file":"{\"version\":1}"}`, wantStatus: "failed"},
		{name: "validation error stays an error", file: &file, status: http.StatusUnprocessableEntity, response: `{"message":"This environment follows its branch. Push the file instead."}`, wantBody: `{"file":"{\"version\":1}"}`, wantErr: true},
		{name: "suspended link", file: &file, status: http.StatusConflict, response: `{"message":"This link is suspended. Resume it first."}`, wantBody: `{"file":"{\"version\":1}"}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, reqs := newProjectConfigTestService(t, tt.status, tt.response)

			revision, err := svc.Apply(context.Background(), "link-uuid", tt.file)

			require.Len(t, *reqs, 1)
			assert.Equal(t, "/api/v1/config/links/link-uuid/apply", (*reqs)[0].uri)
			if tt.wantBody == "" {
				assert.Empty(t, (*reqs)[0].body)
			} else {
				assert.JSONEq(t, tt.wantBody, (*reqs)[0].body)
			}
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, revision.Status)
		})
	}
}

func TestProjectConfigService_LinkRequests(t *testing.T) {
	autoApply := false
	tests := []struct {
		name       string
		call       func(*ProjectConfigService) error
		response   string
		wantMethod string
		wantURI    string
		wantBody   string
	}{
		{
			name: "create",
			call: func(s *ProjectConfigService) error {
				_, err := s.CreateLink(context.Background(), models.ConfigLinkCreateRequest{Project: "p", Environment: "e", Mode: "api"})
				return err
			},
			response: `{"uuid":"link-1"}`, wantMethod: "POST", wantURI: "/api/v1/config/links",
			wantBody: `{"project":"p","environment":"e","mode":"api"}`,
		},
		{
			name: "get",
			call: func(s *ProjectConfigService) error {
				_, err := s.GetLink(context.Background(), "link-1")
				return err
			},
			response: `{"uuid":"link-1"}`, wantMethod: "GET", wantURI: "/api/v1/config/links/link-1",
		},
		{
			name: "update sends false auto_apply",
			call: func(s *ProjectConfigService) error {
				_, err := s.UpdateLink(context.Background(), "link-1", models.ConfigLinkUpdateRequest{AutoApply: &autoApply})
				return err
			},
			response: `{"uuid":"link-1"}`, wantMethod: "PATCH", wantURI: "/api/v1/config/links/link-1",
			wantBody: `{"auto_apply":false}`,
		},
		{
			name:     "delete",
			call:     func(s *ProjectConfigService) error { return s.DeleteLink(context.Background(), "link-1") },
			response: `{"message":"Disconnected."}`, wantMethod: "DELETE", wantURI: "/api/v1/config/links/link-1",
		},
		{
			name: "revisions",
			call: func(s *ProjectConfigService) error {
				_, err := s.Revisions(context.Background(), "link-1")
				return err
			},
			response: `[]`, wantMethod: "GET", wantURI: "/api/v1/config/links/link-1/revisions",
		},
		{
			name: "suspend",
			call: func(s *ProjectConfigService) error {
				_, err := s.Suspend(context.Background(), "link-1")
				return err
			},
			response: `{"uuid":"link-1"}`, wantMethod: "POST", wantURI: "/api/v1/config/links/link-1/suspend",
		},
		{
			name: "resume preview",
			call: func(s *ProjectConfigService) error {
				_, err := s.ResumePreview(context.Background(), "link-1")
				return err
			},
			response: `{"changes":[],"errors":[],"warnings":[],"missing_required":[]}`, wantMethod: "POST", wantURI: "/api/v1/config/links/link-1/resume",
		},
		{
			name: "resume confirmed",
			call: func(s *ProjectConfigService) error {
				_, err := s.Resume(context.Background(), "link-1")
				return err
			},
			response: `{"link":{"uuid":"link-1"}}`, wantMethod: "POST", wantURI: "/api/v1/config/links/link-1/resume?confirm=1",
		},
		{
			name: "approve without delete data sends an empty list",
			call: func(s *ProjectConfigService) error {
				_, err := s.Approve(context.Background(), "rev-1", nil)
				return err
			},
			response: `{"uuid":"rev-1"}`, wantMethod: "POST", wantURI: "/api/v1/config/revisions/rev-1/approve",
			wantBody: `{"delete_data":[]}`,
		},
		{
			name: "approve with delete data",
			call: func(s *ProjectConfigService) error {
				_, err := s.Approve(context.Background(), "rev-1", []string{"db"})
				return err
			},
			response: `{"uuid":"rev-1"}`, wantMethod: "POST", wantURI: "/api/v1/config/revisions/rev-1/approve",
			wantBody: `{"delete_data":["db"]}`,
		},
		{
			name: "release",
			call: func(s *ProjectConfigService) error {
				_, err := s.Release(context.Background(), "rev-1")
				return err
			},
			response: `{"uuid":"rev-1"}`, wantMethod: "POST", wantURI: "/api/v1/config/revisions/rev-1/release",
		},
		{
			name: "retry",
			call: func(s *ProjectConfigService) error {
				_, err := s.Retry(context.Background(), "rev-1")
				return err
			},
			response: `{"uuid":"rev-1"}`, wantMethod: "POST", wantURI: "/api/v1/config/revisions/rev-1/retry",
		},
		{
			name: "uuid is path escaped",
			call: func(s *ProjectConfigService) error {
				_, err := s.GetLink(context.Background(), "../servers")
				return err
			},
			response: `{"uuid":"x"}`, wantMethod: "GET", wantURI: "/api/v1/config/links/..%2Fservers",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, reqs := newProjectConfigTestService(t, http.StatusOK, tt.response)

			require.NoError(t, tt.call(svc))

			require.Len(t, *reqs, 1)
			assert.Equal(t, tt.wantMethod, (*reqs)[0].method)
			assert.Equal(t, tt.wantURI, (*reqs)[0].uri)
			if tt.wantBody == "" {
				assert.Empty(t, (*reqs)[0].body)
			} else {
				assert.JSONEq(t, tt.wantBody, (*reqs)[0].body)
			}
		})
	}
}

func TestProjectConfigService_RevisionPlanDecodes(t *testing.T) {
	svc, _ := newProjectConfigTestService(t, http.StatusOK, `[{"uuid":"rev-1","trigger":"push","commit_sha":"abc","status":"waiting_approval","error":null,
		"plan":{"changes":[{"kind":"remove","key":"db","path":"","destructive":true,"secret":true,"from":null,"to":null}],"errors":[],"warnings":[],"missing_required":[],"destructive":true,"removes_all":false},
		"approved_at":null,"applied_at":null,"created_at":"2026-11-01T00:00:00Z"},
		{"uuid":"rev-0","trigger":"link","status":"failed","error":"bad","plan":null,"created_at":"2026-10-01T00:00:00Z"}]`)

	revisions, err := svc.Revisions(context.Background(), "link-1")

	require.NoError(t, err)
	require.Len(t, revisions, 2)
	require.NotNil(t, revisions[0].Plan)
	assert.True(t, revisions[0].Plan.Destructive)
	assert.Equal(t, "db", revisions[0].Plan.Changes[0].Key)
	assert.Nil(t, revisions[1].Plan)
	assert.Equal(t, "bad", *revisions[1].Error)
}

func TestProjectConfigService_ValidationDetailsInError(t *testing.T) {
	svc, _ := newProjectConfigTestService(t, http.StatusUnprocessableEntity,
		`{"message":"Validation failed.","errors":{"delete_data":["Use only keys of resources that this plan removes."]}}`)

	_, err := svc.Approve(context.Background(), "rev-1", []string{"nope"})

	require.Error(t, err)
	assert.Equal(t, "failed to approve revision rev-1: API error 422 on config/revisions/rev-1/approve: Validation failed. (delete_data: Use only keys of resources that this plan removes.)", err.Error())
}

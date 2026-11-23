package config

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
	"github.com/stretchr/testify/assert"
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

// newFakeAPI starts a fake Coolify API that answers every request with status
// and body, and points a temporary CLI config at it.
func newFakeAPI(t *testing.T, status int, body string) *fakeAPI {
	t.Helper()
	fake := &fakeAPI{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := io.ReadAll(r.Body)
		fake.mu.Lock()
		fake.requests = append(fake.requests, fakeRequest{Method: r.Method, RequestURI: r.RequestURI, Body: string(reqBody)})
		fake.mu.Unlock()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
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

// runCmd executes the config command tree under a minimal root with the global flags.
func runCmd(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	root := &cobra.Command{Use: "coolify", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("token", "", "")
	root.PersistentFlags().String("context", "", "")
	root.PersistentFlags().String("format", "table", "")
	root.PersistentFlags().BoolP("show-sensitive", "s", false, "")
	root.PersistentFlags().Bool("debug", false, "")
	root.AddCommand(NewConfigCommand())

	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), err
}

func writeFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "coolify.json")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}

const planFixture = `{
	"changes": [
		{"kind":"create","key":"api","path":"","destructive":false,"secret":false,"from":null,"to":null},
		{"kind":"update","key":"db","path":"env.DATABASE_URL","destructive":true,"secret":true,"from":null,"to":null}
	],
	"errors": [],
	"warnings": ["\"legacy\" (application) is not managed by coolify.json."],
	"missing_required": {"api": ["STRIPE_KEY", "MAIL_KEY"]},
	"state_hash": "abc",
	"destructive": true,
	"removes_all": false
}`

func TestPlan_TableShowsChangesWarningsAndMissing(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, planFixture)
	path := writeFile(t, `{"version":1}`)

	stdout, err := runCmd(t, "", "config", "plan", "env-uuid", "-f", path)

	require.NoError(t, err)
	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, "/api/v1/environments/env-uuid/config/plan", reqs[0].RequestURI)
	assert.JSONEq(t, `{"file":"{\"version\":1}"}`, reqs[0].Body)
	for _, want := range []string{"resource", "change", "setting", "destructive", "api", "create", "db", "env.DATABASE_URL", "waits for approval", "Warnings:", "legacy", "Missing required values:", "api: STRIPE_KEY, MAIL_KEY"} {
		assert.Contains(t, stdout, want)
	}
	assert.NotContains(t, stdout, "Errors:")
}

func TestPlan_FileErrorsExitNonZero(t *testing.T) {
	newFakeAPI(t, http.StatusUnprocessableEntity, `{"message":"The file is not valid.","errors":[{"path":"resources.api.type","message":"Unknown type."}]}`)
	path := writeFile(t, `{"version":1}`)

	stdout, err := runCmd(t, "", "config", "plan", "env-uuid", "-f", path)

	require.Error(t, err)
	assert.Equal(t, "the file has 1 error(s)", err.Error())
	assert.Contains(t, stdout, "Errors:")
	assert.Contains(t, stdout, "resources.api.type: Unknown type.")
}

func TestPlan_JSONOutputKeepsErrorsAndExitCode(t *testing.T) {
	newFakeAPI(t, http.StatusOK, `{"changes":[],"errors":[{"path":"version","message":"Must be 1."}],"warnings":[],"missing_required":[],"destructive":false,"removes_all":false}`)
	path := writeFile(t, `{}`)

	stdout, err := runCmd(t, "", "config", "plan", "env-uuid", "-f", path, "--format", "json")

	require.Error(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "Must be 1.", out["errors"].([]any)[0].(map[string]any)["message"])
}

func TestPlan_ReadsStdin(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{"changes":[],"errors":[],"warnings":[],"missing_required":[]}`)

	stdout, err := runCmd(t, `{"version":1,"resources":{}}`, "config", "plan", "env-uuid", "-f", "-")

	require.NoError(t, err)
	assert.Contains(t, stdout, "No changes.")
	assert.JSONEq(t, `{"file":"{\"version\":1,\"resources\":{}}"}`, fake.recorded()[0].Body)
}

func TestPlan_MissingFileSendsNoRequest(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{}`)
	_, err := runCmd(t, "", "config", "plan", "env-uuid", "-f", filepath.Join(t.TempDir(), "nope.json"))
	require.Error(t, err)
	assert.Empty(t, fake.recorded())
}

func TestExport_WritesIndentedFile(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{"version":1,"resources":{"api":{"type":"application"}}}`)
	path := filepath.Join(t.TempDir(), "coolify.json")

	stdout, err := runCmd(t, "", "config", "export", "env-uuid", "-o", path)

	require.NoError(t, err)
	assert.Equal(t, "/api/v1/environments/env-uuid/config/export", fake.recorded()[0].RequestURI)
	assert.Contains(t, stdout, "Exported to "+path)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "{\n    \"version\": 1,\n    \"resources\": {\n        \"api\": {\n            \"type\": \"application\"\n        }\n    }\n}\n", string(data))
}

func TestExport_StdoutKeepsKeyOrder(t *testing.T) {
	newFakeAPI(t, http.StatusOK, `{"$schema":"s","version":1,"resources":{}}`)
	stdout, err := runCmd(t, "", "config", "export", "env-uuid")
	require.NoError(t, err)
	assert.Equal(t, "{\n    \"$schema\": \"s\",\n    \"version\": 1,\n    \"resources\": {}\n}\n", stdout)
}

func TestLinkCreate_SendsOnlySetFields(t *testing.T) {
	fake := newFakeAPI(t, http.StatusCreated, `{"uuid":"link-1","mode":"api","project":"p","environment":"production","auto_apply":false}`)

	stdout, err := runCmd(t, "", "config", "link", "create", "--project", "p", "--environment", "production")

	require.NoError(t, err)
	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, http.MethodPost, reqs[0].Method)
	assert.Equal(t, "/api/v1/config/links", reqs[0].RequestURI)
	assert.JSONEq(t, `{"project":"p","environment":"production","mode":"api"}`, reqs[0].Body)
	assert.Contains(t, stdout, "link-1")
}

func TestLinkCreate_GitModeMasksWebhookSecret(t *testing.T) {
	response := `{"uuid":"link-1","mode":"git","project":"p","environment":"production","auto_apply":true,
		"webhook_secret":"s3cr3t","webhook_urls":{"github":"https://c.example/webhooks/source/github/events/manual"}}`
	fake := newFakeAPI(t, http.StatusCreated, response)

	stdout, err := runCmd(t, "", "config", "link", "create", "--project", "p", "--environment", "production",
		"--mode", "git", "--repository", "https://github.com/acme/api", "--branch", "main",
		"--source-type", "public", "--server-uuid", "srv-1", "--auto-apply")

	require.NoError(t, err)
	assert.JSONEq(t, `{"project":"p","environment":"production","mode":"git","auto_apply":true,
		"repository":"https://github.com/acme/api","branch":"main","source_type":"public","server_uuid":"srv-1"}`, fake.recorded()[0].Body)
	assert.Contains(t, stdout, "github: https://c.example/webhooks/source/github/events/manual")
	assert.Contains(t, stdout, "Webhook secret: ********")
	assert.NotContains(t, stdout, "s3cr3t")

	stdout, err = runCmd(t, "", "config", "link", "create", "--project", "p", "--environment", "production", "--mode", "git", "-s")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Webhook secret: s3cr3t")
}

func TestLinkCreate_RequiresProjectAndEnvironment(t *testing.T) {
	fake := newFakeAPI(t, http.StatusCreated, `{}`)
	_, err := runCmd(t, "", "config", "link", "create", "--project", "p")
	require.Error(t, err)
	assert.Empty(t, fake.recorded())
}

func TestLinkUpdate(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{"uuid":"link-1","mode":"git","auto_apply":false}`)

	_, err := runCmd(t, "", "config", "link", "update", "link-1", "--auto-apply=false")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPatch, fake.recorded()[0].Method)
	assert.Equal(t, "/api/v1/config/links/link-1", fake.recorded()[0].RequestURI)
	assert.JSONEq(t, `{"auto_apply":false}`, fake.recorded()[0].Body)

	_, err = runCmd(t, "", "config", "link", "update", "link-1", "--server-uuid", "srv-2", "--destination-uuid", "dst-2")
	require.NoError(t, err)
	assert.JSONEq(t, `{"server_uuid":"srv-2","destination_uuid":"dst-2"}`, fake.recorded()[1].Body)
}

func TestLinkUpdate_NeedsAChange(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{}`)
	_, err := runCmd(t, "", "config", "link", "update", "link-1")
	require.Error(t, err)
	_, err = runCmd(t, "", "config", "link", "update", "link-1", "--destination-uuid", "dst-2")
	require.Error(t, err)
	assert.Empty(t, fake.recorded())
}

func TestLinkDelete_AsksFirst(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{"message":"Disconnected."}`)

	stdout, err := runCmd(t, "no\n", "config", "link", "delete", "link-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Disconnect cancelled.")
	assert.Empty(t, fake.recorded())

	stdout, err = runCmd(t, "yes\n", "config", "link", "delete", "link-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Config link link-1 disconnected.")
	require.Len(t, fake.recorded(), 1)
	assert.Equal(t, http.MethodDelete, fake.recorded()[0].Method)

	_, err = runCmd(t, "", "config", "link", "delete", "link-1", "--force")
	require.NoError(t, err)
	assert.Len(t, fake.recorded(), 2)
}

func TestRevisions_Table(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `[{"uuid":"rev-1","trigger":"push","commit_sha":"abc123","status":"applied","error":null,"plan":null,"created_at":"2026-11-01"}]`)

	stdout, err := runCmd(t, "", "config", "revisions", "link-1")

	require.NoError(t, err)
	assert.Equal(t, "/api/v1/config/links/link-1/revisions", fake.recorded()[0].RequestURI)
	for _, want := range []string{"rev-1", "push", "abc123", "applied"} {
		assert.Contains(t, stdout, want)
	}
}

func TestApply_FileWaitingApprovalShowsHint(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{"uuid":"rev-1","trigger":"api","status":"waiting_approval",
		"plan":{"changes":[{"kind":"remove","key":"old","path":"","destructive":true}],"errors":[],"warnings":[],"missing_required":[],"destructive":true}}`)
	path := writeFile(t, `{"version":1}`)

	stdout, err := runCmd(t, "", "config", "apply", "link-1", "-f", path)

	require.NoError(t, err)
	assert.Equal(t, "/api/v1/config/links/link-1/apply", fake.recorded()[0].RequestURI)
	assert.JSONEq(t, `{"file":"{\"version\":1}"}`, fake.recorded()[0].Body)
	assert.Contains(t, stdout, "remove")
	assert.Contains(t, stdout, "Approve with: coolify config approve rev-1")
}

func TestApply_GitLinkSendsNoFile(t *testing.T) {
	fake := newFakeAPI(t, http.StatusAccepted, `{"uuid":"rev-2","trigger":"api","status":"applied"}`)
	_, err := runCmd(t, "", "config", "apply", "link-1")
	require.NoError(t, err)
	assert.Empty(t, fake.recorded()[0].Body)
}

func TestApply_FailedRevisionExitsNonZero(t *testing.T) {
	newFakeAPI(t, http.StatusUnprocessableEntity, `{"uuid":"rev-3","trigger":"api","status":"failed","error":"Server db-1 not found."}`)
	path := writeFile(t, `{"version":1}`)

	stdout, err := runCmd(t, "", "config", "apply", "link-1", "-f", path)

	require.Error(t, err)
	assert.Equal(t, "revision rev-3 failed: Server db-1 not found.", err.Error())
	assert.Contains(t, stdout, "rev-3")
}

func TestResume_PreviewAndConfirm(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{"changes":[],"errors":[],"warnings":[],"missing_required":[]}`)
	stdout, err := runCmd(t, "", "config", "resume", "link-1")
	require.NoError(t, err)
	assert.Equal(t, "/api/v1/config/links/link-1/resume", fake.recorded()[0].RequestURI)
	assert.Contains(t, stdout, "No changes.")
	assert.Contains(t, stdout, "coolify config resume link-1 --confirm")

	fake = newFakeAPI(t, http.StatusOK, `{"link":{"uuid":"link-1","mode":"api"},"revision":{"uuid":"rev-9","status":"applied"}}`)
	stdout, err = runCmd(t, "", "config", "resume", "link-1", "--confirm")
	require.NoError(t, err)
	assert.Equal(t, "/api/v1/config/links/link-1/resume?confirm=1", fake.recorded()[0].RequestURI)
	assert.Contains(t, stdout, "link-1")
	assert.Contains(t, stdout, "rev-9")
}

func TestSuspend(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{"uuid":"link-1","suspended_at":"2026-11-23T10:00:00Z"}`)
	stdout, err := runCmd(t, "", "config", "suspend", "link-1")
	require.NoError(t, err)
	assert.Equal(t, "/api/v1/config/links/link-1/suspend", fake.recorded()[0].RequestURI)
	assert.Contains(t, stdout, "2026-11-23T10:00:00Z")
}

func TestApprove_DeleteDataAsksFirst(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{"uuid":"rev-1","status":"applying"}`)

	stdout, err := runCmd(t, "n\n", "config", "approve", "rev-1", "--delete-data", "old-db", "--delete-data", "cache")
	require.NoError(t, err)
	assert.Contains(t, stdout, "old-db, cache")
	assert.Contains(t, stdout, "Approval cancelled.")
	assert.Empty(t, fake.recorded())

	_, err = runCmd(t, "yes\n", "config", "approve", "rev-1", "--delete-data", "old-db,cache")
	require.NoError(t, err)
	require.Len(t, fake.recorded(), 1)
	assert.Equal(t, "/api/v1/config/revisions/rev-1/approve", fake.recorded()[0].RequestURI)
	assert.JSONEq(t, `{"delete_data":["old-db","cache"]}`, fake.recorded()[0].Body)
}

func TestApprove_WithoutDeleteDataDoesNotAsk(t *testing.T) {
	fake := newFakeAPI(t, http.StatusOK, `{"uuid":"rev-1","status":"applied"}`)
	_, err := runCmd(t, "", "config", "approve", "rev-1")
	require.NoError(t, err)
	assert.JSONEq(t, `{"delete_data":[]}`, fake.recorded()[0].Body)
}

func TestReleaseAndRetry(t *testing.T) {
	for _, action := range []string{"release", "retry"} {
		t.Run(action, func(t *testing.T) {
			fake := newFakeAPI(t, http.StatusOK, `{"uuid":"rev-1","status":"planned"}`)
			stdout, err := runCmd(t, "", "config", action, "rev-1", "--format", "json")
			require.NoError(t, err)
			assert.Equal(t, "/api/v1/config/revisions/rev-1/"+action, fake.recorded()[0].RequestURI)
			var out map[string]any
			require.NoError(t, json.Unmarshal([]byte(stdout), &out))
			assert.Equal(t, "planned", out["status"])
		})
	}
}

func TestConfig_RejectsUnknownArgs(t *testing.T) {
	_, err := runCmd(t, "", "config", "bogus")
	require.Error(t, err)
}

func TestExport_UnescapesSlashesOnly(t *testing.T) {
	newFakeAPI(t, http.StatusOK, `{"repository":"https:\/\/github.com\/acme\/api","path":"C:\\\/x","quote":"a\"b"}`)
	stdout, err := runCmd(t, "", "config", "export", "env-uuid")
	require.NoError(t, err)
	var out map[string]string
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, map[string]string{"repository": "https://github.com/acme/api", "path": `C:\/x`, "quote": `a"b`}, out)
	assert.Contains(t, stdout, `"https://github.com/acme/api"`)
}

func TestApprove_ApprovedRevisionHasNoApproveHint(t *testing.T) {
	newFakeAPI(t, http.StatusOK, `{"uuid":"rev-1","status":"waiting_approval","approved_at":"2026-11-23T10:00:00Z","plan":{"changes":[],"errors":[],"warnings":[],"missing_required":[]}}`)
	stdout, err := runCmd(t, "", "config", "approve", "rev-1")
	require.NoError(t, err)
	assert.NotContains(t, stdout, "Approve with:")
	assert.Contains(t, stdout, "Approved. Coolify applies the revision in the background.")
}

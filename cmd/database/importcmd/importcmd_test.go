package importcmd

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

type recordedRequest struct {
	Method string
	Path   string
	Body   string // JSON body, or the uploaded file content for multipart
	Fields map[string]string
}

type fakeAPI struct {
	mu       sync.Mutex
	requests []recordedRequest
}

func (f *fakeAPI) recorded() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]recordedRequest(nil), f.requests...)
}

// newFakeAPI starts a fake Coolify API that implements the import endpoints
// and points a temporary CLI config at it. statuses are returned in order by
// the status endpoint (the last one repeats).
func newFakeAPI(t *testing.T, statuses ...string) *fakeAPI {
	t.Helper()
	fake := &fakeAPI{}
	statusCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recordedRequest{Method: r.Method, Path: r.URL.Path}
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			require.NoError(t, r.ParseMultipartForm(1<<20))
			rec.Fields = map[string]string{}
			for key, values := range r.MultipartForm.Value {
				rec.Fields[key] = values[0]
			}
			file, _, err := r.FormFile("file")
			require.NoError(t, err)
			data, _ := io.ReadAll(file)
			_ = file.Close()
			rec.Body = string(data)
		} else {
			data, _ := io.ReadAll(r.Body)
			rec.Body = string(data)
		}
		fake.mu.Lock()
		fake.requests = append(fake.requests, rec)
		fake.mu.Unlock()

		switch {
		case strings.HasSuffix(r.URL.Path, "/imports/uploads"):
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"upload_id": rec.Fields["upload_id"], "filename": "dump.sql", "mime_type": "text/plain", "size": len(rec.Body)})
		case strings.HasSuffix(r.URL.Path, "/imports") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"id":42,"status":"queued","message":"Database import queued.","status_url":"/api/v1/x/42"}`))
		case strings.HasSuffix(r.URL.Path, "/imports/42"):
			fake.mu.Lock()
			body := statuses[min(statusCalls, len(statuses)-1)]
			statusCalls++
			fake.mu.Unlock()
			_, _ = w.Write([]byte(body))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		}
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

// run executes `coolify <parent> import <args>` where parent is either the
// standalone or service database command.
func run(t *testing.T, importCmd *cobra.Command, args ...string) (string, string, error) {
	t.Helper()
	root := &cobra.Command{Use: "coolify", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().String("token", "", "")
	root.PersistentFlags().String("context", "", "")
	root.PersistentFlags().String("format", "table", "")
	root.PersistentFlags().Bool("debug", false, "")
	root.AddCommand(importCmd)

	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(append([]string{"import"}, args...))
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

func writeDump(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dump.sql")
	require.NoError(t, os.WriteFile(path, []byte("SELECT 1;"), 0o600))
	return path
}

const finishedStatus = `{"id":42,"status":"finished","exit_code":0,"output":"restore ok","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:01:00Z","finished_at":"2026-01-01T00:01:00Z"}`

func TestStart_FileUploadsThenStartsAndWaits_Standalone(t *testing.T) {
	fake := newFakeAPI(t, `{"id":42,"status":"in_progress"}`, finishedStatus)

	stdout, _, err := run(t, NewStandaloneCommand(), "start", "db-1", "--file", writeDump(t), "--replace-existing", "--wait", "--poll-interval", "1ms")
	require.NoError(t, err)

	reqs := fake.recorded()
	require.Len(t, reqs, 4)
	assert.Equal(t, "/api/v1/databases/db-1/imports/uploads", reqs[0].Path)
	assert.Equal(t, "SELECT 1;", reqs[0].Body)
	uploadID := reqs[0].Fields["upload_id"]
	assert.NotEmpty(t, uploadID)

	assert.Equal(t, http.MethodPost, reqs[1].Method)
	assert.Equal(t, "/api/v1/databases/db-1/imports", reqs[1].Path)
	assert.JSONEq(t, `{"source":"upload","upload_id":"`+uploadID+`","replace_existing":true}`, reqs[1].Body)

	assert.Equal(t, http.MethodGet, reqs[2].Method)
	assert.Equal(t, "/api/v1/databases/db-1/imports/42", reqs[3].Path)
	assert.Contains(t, stdout, "finished")
	assert.Contains(t, stdout, "restore ok")
}

func TestStart_ServiceDatabaseS3_NoWait(t *testing.T) {
	fake := newFakeAPI(t, finishedStatus)

	stdout, _, err := run(t, NewServiceDatabaseCommand(), "start", "svc-1", "db-1", "--s3-storage-uuid", "s3-1", "--s3-path", "/b/db.dump", "--dump-all")
	require.NoError(t, err)

	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, "/api/v1/services/svc-1/databases/db-1/imports", reqs[0].Path)
	assert.JSONEq(t, `{"source":"s3","s3_storage_uuid":"s3-1","path":"/b/db.dump","dump_all":true}`, reqs[0].Body)
	assert.Contains(t, stdout, "42")
	assert.Contains(t, stdout, "coolify service database import status svc-1 db-1 42 --wait")
}

func TestStart_WaitFailsWhenImportFails(t *testing.T) {
	newFakeAPI(t, `{"id":42,"status":"error","exit_code":1,"output":"boom"}`)

	stdout, _, err := run(t, NewStandaloneCommand(), "start", "db-1", "--server-path", "/var/backups/db.sql.gz", "--wait", "--poll-interval", "1ms")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `status "error"`)
	assert.Contains(t, stdout, "boom")
}

func TestStart_JSONOutput(t *testing.T) {
	newFakeAPI(t, finishedStatus)

	stdout, _, err := run(t, NewStandaloneCommand(), "start", "db-1", "--upload-id", "11111111-2222-4333-8444-555555555555", "--format", "json")
	require.NoError(t, err)
	var started map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &started))
	assert.EqualValues(t, 42, started["id"])
	assert.NotContains(t, stdout, "Check progress")
}

func TestUpload_PrintsUploadID(t *testing.T) {
	fake := newFakeAPI(t, finishedStatus)

	stdout, _, err := run(t, NewServiceDatabaseCommand(), "upload", "svc-1", "db-1", writeDump(t), "--format", "json")
	require.NoError(t, err)
	reqs := fake.recorded()
	require.Len(t, reqs, 1)
	assert.Equal(t, "/api/v1/services/svc-1/databases/db-1/imports/uploads", reqs[0].Path)
	assert.Contains(t, stdout, reqs[0].Fields["upload_id"])
}

func TestStatus_PrintsStatus(t *testing.T) {
	fake := newFakeAPI(t, finishedStatus)

	stdout, _, err := run(t, NewStandaloneCommand(), "status", "db-1", "42")
	require.NoError(t, err)
	require.Len(t, fake.recorded(), 1)
	assert.Equal(t, "/api/v1/databases/db-1/imports/42", fake.recorded()[0].Path)
	assert.Contains(t, stdout, "finished")
	assert.Contains(t, stdout, "Output:\nrestore ok")
}

func TestStatus_InvalidActivityID(t *testing.T) {
	_, _, err := run(t, NewStandaloneCommand(), "status", "db-1", "abc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid activity ID")
}

func TestBuildStartRequest_SourceValidation(t *testing.T) {
	tests := []struct {
		name    string
		flags   map[string]string
		wantErr string
		want    string
		file    string
	}{
		{name: "no source", flags: map[string]string{}, wantErr: "exactly one source"},
		{name: "two sources", flags: map[string]string{"file": "a.sql", "server-path": "/a.sql"}, wantErr: "exactly one source"},
		{name: "s3 without path", flags: map[string]string{"s3-storage-uuid": "s3"}, wantErr: "must be used together"},
		{name: "s3 path without storage", flags: map[string]string{"s3-path": "/a"}, wantErr: "must be used together"},
		{name: "file", flags: map[string]string{"file": "a.sql"}, want: `{"source":"upload"}`, file: "a.sql"},
		{name: "server path with options", flags: map[string]string{"server-path": "/a.sql", "keep-owners": "true", "restore-mysql-users": "false", "sqlite-database": "app.db"},
			want: `{"source":"server","path":"/a.sql","keep_owners":true,"restore_mysql_users":false,"sqlite_database":"app.db"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newStartCommand(targetSpec{})
			for flag, value := range tt.flags {
				require.NoError(t, cmd.Flags().Set(flag, value))
			}
			req, file, err := buildStartRequest(cmd)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.file, file)
			body, err := json.Marshal(req)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(body))
		})
	}
}

func TestHumanBytes(t *testing.T) {
	assert.Equal(t, "512 B", humanBytes(512))
	assert.Equal(t, "1.5 KiB", humanBytes(1536))
	assert.Equal(t, "10.0 MiB", humanBytes(10*1024*1024))
}

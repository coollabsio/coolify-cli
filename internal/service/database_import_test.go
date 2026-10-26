package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestDatabaseImportTarget_Paths(t *testing.T) {
	assert.Equal(t, "databases/db-1/imports", StandaloneDatabaseImportTarget("db-1").basePath())
	assert.Equal(t, "services/svc-1/databases/db-1/imports", ServiceDatabaseImportTarget("svc-1", "db-1").basePath())
}

type uploadedChunk struct {
	fields  map[string]string
	name    string
	content string
}

func TestDatabaseImportService_Upload_Chunked(t *testing.T) {
	tests := []struct {
		name       string
		target     DatabaseImportTarget
		wantPath   string
		content    string
		chunkSize  int64
		wantChunks []string
	}{
		{
			name:       "standalone database, three chunks",
			target:     StandaloneDatabaseImportTarget("db-1"),
			wantPath:   "/api/v1/databases/db-1/imports/uploads",
			content:    "0123456789",
			chunkSize:  4,
			wantChunks: []string{"0123", "4567", "89"},
		},
		{
			name:       "service database, single chunk",
			target:     ServiceDatabaseImportTarget("svc-1", "db-1"),
			wantPath:   "/api/v1/services/svc-1/databases/db-1/imports/uploads",
			content:    "SELECT 1;",
			chunkSize:  DefaultImportChunkSize,
			wantChunks: []string{"SELECT 1;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			var chunks []uploadedChunk
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, tt.wantPath, r.URL.Path)
				assert.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data"))
				r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
				require.NoError(t, r.ParseMultipartForm(1<<20))
				file, header, err := r.FormFile("file")
				require.NoError(t, err)
				defer file.Close()
				data, _ := io.ReadAll(file)

				fields := map[string]string{}
				for key, values := range r.MultipartForm.Value {
					fields[key] = values[0]
				}
				mu.Lock()
				chunks = append(chunks, uploadedChunk{fields: fields, name: header.Filename, content: string(data)})
				last := len(chunks) == len(tt.wantChunks)
				mu.Unlock()

				if !last {
					_ = json.NewEncoder(w).Encode(map[string]any{"upload_id": fields["upload_id"], "done": 50, "status": true})
					return
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"upload_id": fields["upload_id"], "filename": header.Filename, "mime_type": "text/plain", "size": len(tt.content)})
			}))
			defer server.Close()

			path := writeTempFile(t, "dump.sql", tt.content)
			svc := NewDatabaseImportService(api.NewClient(server.URL, "test-token", api.WithRetries(0)))
			svc.ChunkSize = tt.chunkSize

			var progress []int64
			upload, err := svc.Upload(context.Background(), tt.target, path, func(sent, total int64) {
				assert.Equal(t, int64(len(tt.content)), total)
				progress = append(progress, sent)
			})
			require.NoError(t, err)

			require.Len(t, chunks, len(tt.wantChunks))
			uploadID := chunks[0].fields["upload_id"]
			assert.Regexp(t, uuidV4Pattern, uploadID)
			assert.Equal(t, uploadID, upload.UploadID)
			assert.Equal(t, "dump.sql", upload.Filename)
			assert.Equal(t, int64(len(tt.content)), upload.Size)

			var offset int
			for i, chunk := range chunks {
				assert.Equal(t, tt.wantChunks[i], chunk.content)
				assert.Equal(t, "dump.sql", chunk.name)
				assert.Equal(t, uploadID, chunk.fields["upload_id"])
				assert.Equal(t, uploadID, chunk.fields["dzuuid"])
				assert.Equal(t, strconv.Itoa(i), chunk.fields["dzchunkindex"])
				assert.Equal(t, strconv.Itoa(len(tt.wantChunks)), chunk.fields["dztotalchunkcount"])
				assert.Equal(t, strconv.Itoa(len(tt.content)), chunk.fields["dztotalfilesize"])
				assert.Equal(t, strconv.Itoa(int(tt.chunkSize)), chunk.fields["dzchunksize"])
				assert.Equal(t, strconv.Itoa(offset), chunk.fields["dzchunkbyteoffset"])
				offset += len(chunk.content)
			}
			assert.Equal(t, int64(len(tt.content)), progress[len(progress)-1])
			assert.Len(t, progress, len(tt.wantChunks))
		})
	}
}

func TestDatabaseImportService_Upload_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"Validation failed.","errors":{"file":["Unsupported backup file extension."]}}`))
	}))
	defer server.Close()

	path := writeTempFile(t, "dump.txt", "data")
	svc := NewDatabaseImportService(api.NewClient(server.URL, "test-token"))
	_, err := svc.Upload(context.Background(), StandaloneDatabaseImportTarget("db-1"), path, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chunk 1/1")
	assert.Contains(t, err.Error(), "Validation failed.")
}

func TestDatabaseImportService_Upload_IncompleteFinalResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"upload_id":"x","done":50,"status":true}`))
	}))
	defer server.Close()

	path := writeTempFile(t, "dump.sql", "data")
	svc := NewDatabaseImportService(api.NewClient(server.URL, "test-token"))
	_, err := svc.Upload(context.Background(), StandaloneDatabaseImportTarget("db-1"), path, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "did not complete")
}

func TestDatabaseImportService_Upload_LocalFileErrors(t *testing.T) {
	svc := NewDatabaseImportService(api.NewClient("http://127.0.0.1:0", "test-token"))
	target := StandaloneDatabaseImportTarget("db-1")

	_, err := svc.Upload(context.Background(), target, filepath.Join(t.TempDir(), "missing.sql"), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open backup file")

	_, err = svc.Upload(context.Background(), target, writeTempFile(t, "empty.sql", ""), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is empty")

	_, err = svc.Upload(context.Background(), target, t.TempDir(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")
}

func TestDatabaseImportService_Start(t *testing.T) {
	uploadID := "11111111-2222-4333-8444-555555555555"
	tests := []struct {
		name     string
		target   DatabaseImportTarget
		wantPath string
		req      *models.DatabaseImportRequest
		wantBody string
	}{
		{
			name:     "standalone upload source",
			target:   StandaloneDatabaseImportTarget("db-1"),
			wantPath: "/api/v1/databases/db-1/imports",
			req: &models.DatabaseImportRequest{
				Source:          models.DatabaseImportSourceUpload,
				UploadID:        &uploadID,
				ReplaceExisting: importBoolPtr(true),
			},
			wantBody: `{"source":"upload","upload_id":"11111111-2222-4333-8444-555555555555","replace_existing":true}`,
		},
		{
			name:     "service database s3 source",
			target:   ServiceDatabaseImportTarget("svc-1", "db-1"),
			wantPath: "/api/v1/services/svc-1/databases/db-1/imports",
			req: &models.DatabaseImportRequest{
				Source:        models.DatabaseImportSourceS3,
				S3StorageUUID: importStrPtr("s3-1"),
				Path:          importStrPtr("/backups/db.sql.gz"),
				DumpAll:       importBoolPtr(true),
			},
			wantBody: `{"source":"s3","s3_storage_uuid":"s3-1","path":"/backups/db.sql.gz","dump_all":true}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, tt.wantPath, r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				body, _ := io.ReadAll(r.Body)
				assert.JSONEq(t, tt.wantBody, string(body))
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(`{"id":42,"status":"queued","message":"Database import queued.","status_url":"/api/v1/x/42"}`))
			}))
			defer server.Close()

			svc := NewDatabaseImportService(api.NewClient(server.URL, "test-token"))
			started, err := svc.Start(context.Background(), tt.target, tt.req)
			require.NoError(t, err)
			assert.Equal(t, 42, started.ID)
			assert.Equal(t, "queued", started.Status)
			assert.Equal(t, "Database import queued.", started.Message)
		})
	}
}

func TestDatabaseImportService_Start_Conflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"message":"Another import is already in progress."}`))
	}))
	defer server.Close()

	svc := NewDatabaseImportService(api.NewClient(server.URL, "test-token"))
	_, err := svc.Start(context.Background(), StandaloneDatabaseImportTarget("db-1"), &models.DatabaseImportRequest{Source: "server", Path: importStrPtr("/tmp/a.sql")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already in progress")
}

func TestDatabaseImportService_Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/services/svc-1/databases/db-1/imports/42", r.URL.Path)
		_, _ = w.Write([]byte(`{"id":42,"status":"finished","exit_code":0,"output":"done","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:01:00Z","finished_at":"2026-01-01T00:01:00Z"}`))
	}))
	defer server.Close()

	svc := NewDatabaseImportService(api.NewClient(server.URL, "test-token"))
	status, err := svc.Status(context.Background(), ServiceDatabaseImportTarget("svc-1", "db-1"), 42)
	require.NoError(t, err)
	assert.Equal(t, 42, status.ID)
	assert.True(t, status.IsTerminal())
	assert.True(t, status.Succeeded())
	require.NotNil(t, status.Output)
	assert.Equal(t, "done", *status.Output)
}

func TestDatabaseImportService_Status_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Database import not found."}`))
	}))
	defer server.Close()

	svc := NewDatabaseImportService(api.NewClient(server.URL, "test-token"))
	_, err := svc.Status(context.Background(), StandaloneDatabaseImportTarget("db-1"), 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Database import not found.")
}

func TestDatabaseImportService_Wait(t *testing.T) {
	statuses := []string{
		`{"id":42,"status":"queued","exit_code":null}`,
		`{"id":42,"status":"in_progress","exit_code":null}`,
		`{"id":42,"status":"error","exit_code":1,"output":"boom"}`,
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-1/imports/42", r.URL.Path)
		_, _ = w.Write([]byte(statuses[min(calls, len(statuses)-1)]))
		calls++
	}))
	defer server.Close()

	svc := NewDatabaseImportService(api.NewClient(server.URL, "test-token"))
	status, err := svc.Wait(context.Background(), StandaloneDatabaseImportTarget("db-1"), 42, time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, 3, calls)
	assert.Equal(t, "error", status.Status)
	assert.False(t, status.Succeeded())
}

func TestDatabaseImportService_Wait_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":42,"status":"in_progress"}`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	svc := NewDatabaseImportService(api.NewClient(server.URL, "test-token"))
	_, err := svc.Wait(ctx, StandaloneDatabaseImportTarget("db-1"), 42, 5*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestNewUUIDv4(t *testing.T) {
	a, err := newUUIDv4()
	require.NoError(t, err)
	b, err := newUUIDv4()
	require.NoError(t, err)
	assert.Regexp(t, uuidV4Pattern, a)
	assert.NotEqual(t, a, b)
}

func importStrPtr(s string) *string { return &s }

func importBoolPtr(b bool) *bool { return &b }

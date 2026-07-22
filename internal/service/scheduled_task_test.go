package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestApplicationService_ScheduledTasks_CRUDPaths(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/applications/app-1/scheduled-tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode([]models.ScheduledTask{
				{UUID: "task-1", Name: "backup", Command: "echo hi", Frequency: "0 * * * *", Enabled: true, Timeout: 300},
			})
		case http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			var req models.ScheduledTaskCreateRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			if req.Name != "cleanup" || req.Command != "rm -rf /tmp" || req.Frequency != "daily" {
				t.Fatalf("unexpected create body: %#v", req)
			}
			if req.Timeout == nil || *req.Timeout != 120 {
				t.Fatalf("expected timeout 120, got %#v", req.Timeout)
			}
			_ = json.NewEncoder(w).Encode(models.ScheduledTask{
				UUID: "task-new", Name: req.Name, Command: req.Command, Frequency: req.Frequency, Enabled: true, Timeout: 120,
			})
		default:
			t.Fatalf("unexpected method %s on list/create path", r.Method)
		}
	})

	mux.HandleFunc("/api/v1/applications/app-1/scheduled-tasks/task-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			body, _ := io.ReadAll(r.Body)
			var req models.ScheduledTaskUpdateRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			if req.Name == nil || *req.Name != "renamed" {
				t.Fatalf("unexpected update body: %#v", req)
			}
			_ = json.NewEncoder(w).Encode(models.ScheduledTask{
				UUID: "task-1", Name: "renamed", Command: "echo hi", Frequency: "0 * * * *", Enabled: true, Timeout: 300,
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message":"Scheduled task deleted."}`))
		default:
			t.Fatalf("unexpected method %s on task path", r.Method)
		}
	})

	mux.HandleFunc("/api/v1/applications/app-1/scheduled-tasks/task-1/executions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode([]models.ScheduledTaskExecution{
			{UUID: "exec-1", Status: "success", RetryCount: 0},
		})
	})

	mux.HandleFunc("/api/v1/applications/app-1/scheduled-tasks/task-1/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Scheduled task execution queued."})
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	svc := NewApplicationService(api.NewClient(server.URL, "token", api.WithRetries(0)))
	ctx := context.Background()

	list, err := svc.ListScheduledTasks(ctx, "app-1")
	if err != nil || len(list) != 1 || list[0].UUID != "task-1" {
		t.Fatalf("list: %v %#v", err, list)
	}

	timeout := 120
	created, err := svc.CreateScheduledTask(ctx, "app-1", models.ScheduledTaskCreateRequest{
		Name: "cleanup", Command: "rm -rf /tmp", Frequency: "daily", Timeout: &timeout,
	})
	if err != nil || created == nil || created.UUID != "task-new" {
		t.Fatalf("create: %v %#v", err, created)
	}

	name := "renamed"
	updated, err := svc.UpdateScheduledTask(ctx, "app-1", "task-1", models.ScheduledTaskUpdateRequest{Name: &name})
	if err != nil || updated == nil || updated.Name != "renamed" {
		t.Fatalf("update: %v %#v", err, updated)
	}

	if err := svc.DeleteScheduledTask(ctx, "app-1", "task-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	execs, err := svc.ListScheduledTaskExecutions(ctx, "app-1", "task-1")
	if err != nil || len(execs) != 1 || execs[0].UUID != "exec-1" {
		t.Fatalf("executions: %v %#v", err, execs)
	}

	resp, err := svc.ExecuteScheduledTask(ctx, "app-1", "task-1")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if msg, _ := resp["message"].(string); msg != "Scheduled task execution queued." {
		t.Fatalf("execute message: %#v", resp)
	}
}

func TestService_ScheduledTasks_CRUDPaths(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/services/svc-1/scheduled-tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode([]models.ScheduledTask{
				{UUID: "task-1", Name: "ping", Command: "true", Frequency: "*/5 * * * *", Enabled: true, Timeout: 60},
			})
		case http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			var req models.ScheduledTaskCreateRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			if req.Name != "job" || req.Command != "date" || req.Frequency != "hourly" {
				t.Fatalf("unexpected create body: %#v", req)
			}
			if req.Container == nil || *req.Container != "app" {
				t.Fatalf("expected container app, got %#v", req.Container)
			}
			if req.Enabled == nil || *req.Enabled {
				t.Fatalf("expected enabled false, got %#v", req.Enabled)
			}
			_ = json.NewEncoder(w).Encode(models.ScheduledTask{
				UUID: "task-new", Name: req.Name, Command: req.Command, Frequency: req.Frequency, Enabled: false, Timeout: 300,
			})
		default:
			t.Fatalf("unexpected method %s on list/create path", r.Method)
		}
	})

	mux.HandleFunc("/api/v1/services/svc-1/scheduled-tasks/task-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			body, _ := io.ReadAll(r.Body)
			var req models.ScheduledTaskUpdateRequest
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("decode update body: %v", err)
			}
			if req.Frequency == nil || *req.Frequency != "weekly" {
				t.Fatalf("unexpected update body: %#v", req)
			}
			_ = json.NewEncoder(w).Encode(models.ScheduledTask{
				UUID: "task-1", Name: "ping", Command: "true", Frequency: "weekly", Enabled: true, Timeout: 60,
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected method %s on task path", r.Method)
		}
	})

	mux.HandleFunc("/api/v1/services/svc-1/scheduled-tasks/task-1/executions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode([]models.ScheduledTaskExecution{
			{UUID: "exec-9", Status: "failed", RetryCount: 1},
		})
	})

	mux.HandleFunc("/api/v1/services/svc-1/scheduled-tasks/task-1/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Scheduled task execution queued."})
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	svc := NewService(api.NewClient(server.URL, "token", api.WithRetries(0)))
	ctx := context.Background()

	list, err := svc.ListScheduledTasks(ctx, "svc-1")
	if err != nil || len(list) != 1 || list[0].UUID != "task-1" {
		t.Fatalf("list: %v %#v", err, list)
	}

	container := "app"
	enabled := false
	created, err := svc.CreateScheduledTask(ctx, "svc-1", models.ScheduledTaskCreateRequest{
		Name: "job", Command: "date", Frequency: "hourly", Container: &container, Enabled: &enabled,
	})
	if err != nil || created == nil || created.UUID != "task-new" {
		t.Fatalf("create: %v %#v", err, created)
	}

	freq := "weekly"
	updated, err := svc.UpdateScheduledTask(ctx, "svc-1", "task-1", models.ScheduledTaskUpdateRequest{Frequency: &freq})
	if err != nil || updated == nil || updated.Frequency != "weekly" {
		t.Fatalf("update: %v %#v", err, updated)
	}

	if err := svc.DeleteScheduledTask(ctx, "svc-1", "task-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	execs, err := svc.ListScheduledTaskExecutions(ctx, "svc-1", "task-1")
	if err != nil || len(execs) != 1 || execs[0].Status != "failed" {
		t.Fatalf("executions: %v %#v", err, execs)
	}

	resp, err := svc.ExecuteScheduledTask(ctx, "svc-1", "task-1")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if msg, _ := resp["message"].(string); msg != "Scheduled task execution queued." {
		t.Fatalf("execute message: %#v", resp)
	}
}

func TestApplicationService_ScheduledTasks_PathEscaping(t *testing.T) {
	mux := http.NewServeMux()
	// PathEscape leaves most UUIDs alone; ensure path is built under scheduled-tasks.
	mux.HandleFunc("/api/v1/applications/app%2Fuuid/scheduled-tasks/task%2Fuuid/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "ok"})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	svc := NewApplicationService(api.NewClient(server.URL, "token", api.WithRetries(0)))

	resp, err := svc.ExecuteScheduledTask(context.Background(), "app/uuid", "task/uuid")
	if err != nil {
		t.Fatalf("execute escaped: %v", err)
	}
	if msg, _ := resp["message"].(string); msg != "ok" {
		t.Fatalf("unexpected resp: %#v", resp)
	}
}

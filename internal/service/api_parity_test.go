package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestS3StorageService_CRUDPaths(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/s3-storages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode([]models.S3Storage{{UUID: "s3-1", Name: "minio"}})
		case http.MethodPost:
			_ = json.NewEncoder(w).Encode(models.UUID{UUID: "s3-new"})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	mux.HandleFunc("/api/v1/s3-storages/s3-1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(models.S3Storage{UUID: "s3-1", Name: "minio"})
		case http.MethodPatch:
			_ = json.NewEncoder(w).Encode(models.UUID{UUID: "s3-1"})
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	mux.HandleFunc("/api/v1/s3-storages/s3-1/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		_ = json.NewEncoder(w).Encode(models.S3StorageValidation{Valid: true, Message: "ok"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	svc := NewS3StorageService(api.NewClient(server.URL, "token", api.WithRetries(0)))
	ctx := context.Background()

	list, err := svc.List(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %#v", err, list)
	}
	got, err := svc.Get(ctx, "s3-1")
	if err != nil || got.UUID != "s3-1" {
		t.Fatalf("get: %v %#v", err, got)
	}
	created, err := svc.Create(ctx, models.S3StorageCreateRequest{Name: "n", Endpoint: "http://e", Bucket: "b", Region: "r", Key: "k", Secret: "s"})
	if err != nil || created.UUID != "s3-new" {
		t.Fatalf("create: %v %#v", err, created)
	}
	if _, err := svc.Update(ctx, "s3-1", models.S3StorageUpdateRequest{Name: strPtr("x")}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := svc.Delete(ctx, "s3-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	val, err := svc.Validate(ctx, "s3-1")
	if err != nil || !val.Valid {
		t.Fatalf("validate: %v %#v", err, val)
	}
}

func TestSharedEnvService_TeamPaths(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/team/envs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode([]models.SharedEnvironmentVariable{{ID: 1, Key: "A"}})
		case http.MethodPost:
			_ = json.NewEncoder(w).Encode(models.SharedEnvCreateResponse{ID: 9})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	mux.HandleFunc("/api/v1/team/envs/9", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			_ = json.NewEncoder(w).Encode(models.SharedEnvironmentVariable{ID: 9, Key: "B"})
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	svc := NewSharedEnvService(api.NewClient(server.URL, "token", api.WithRetries(0)))
	ctx := context.Background()

	list, err := svc.ListTeam(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v", err)
	}
	created, err := svc.CreateTeam(ctx, models.SharedEnvCreateRequest{Key: "A", Value: "1"})
	if err != nil || created.ID != 9 {
		t.Fatalf("create: %v %#v", err, created)
	}
	if _, err := svc.UpdateTeam(ctx, 9, models.SharedEnvUpdateRequest{Key: strPtr("B")}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := svc.DeleteTeam(ctx, 9); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestNotificationService_Paths(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/notifications/webhook", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"webhook_enabled": false})
		case http.MethodPatch:
			_ = json.NewEncoder(w).Encode(map[string]any{"webhook_enabled": true})
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	svc := NewNotificationService(api.NewClient(server.URL, "token", api.WithRetries(0)))
	ctx := context.Background()
	if _, err := svc.Get(ctx, "webhook"); err != nil {
		t.Fatalf("get: %v", err)
	}
	out, err := svc.Update(ctx, "webhook", map[string]any{"webhook_enabled": true})
	if err != nil || out["webhook_enabled"] != true {
		t.Fatalf("update: %v %#v", err, out)
	}
}

func strPtr(s string) *string { return &s }

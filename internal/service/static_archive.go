package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// MaxStaticArchiveSize is the server-side limit for a compressed site archive (1 GiB).
const MaxStaticArchiveSize int64 = 1024 * 1024 * 1024

// StaticArchiveService uploads site archives and creates or deploys static
// site applications from them.
type StaticArchiveService struct {
	client *api.Client
	// ChunkSize is the upload chunk size in bytes.
	ChunkSize int64
}

// NewStaticArchiveService creates a new static archive service.
func NewStaticArchiveService(client *api.Client) *StaticArchiveService {
	return &StaticArchiveService{client: client, ChunkSize: DefaultImportChunkSize}
}

// Upload uploads a .tar or .tar.gz site archive in chunks and returns the
// validated upload whose UploadID can create or deploy an application.
// appUUID names the application the upload is for, or is empty for a new one.
func (s *StaticArchiveService) Upload(ctx context.Context, archivePath, appUUID string, progress UploadProgress) (*models.StaticArchiveUpload, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open site archive: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat site archive: %w", err)
	}
	size := info.Size()
	if size == 0 {
		return nil, fmt.Errorf("site archive %s is empty", archivePath)
	}
	if size > MaxStaticArchiveSize {
		return nil, fmt.Errorf("site archive is %d bytes, more than the 1 GiB limit", size)
	}

	var fields map[string]string
	if appUUID != "" {
		fields = map[string]string{"application_uuid": appUUID}
	}
	upload, err := uploadChunks[models.StaticArchiveUpload](ctx, s.client, "applications/archive/uploads", file, size, s.ChunkSize, fields, progress)
	if err != nil {
		return nil, err
	}
	if upload.SHA == "" {
		return nil, errors.New("upload did not complete: the server did not assemble the uploaded chunks")
	}
	return &upload, nil
}

// Create creates a static site application from an uploaded archive.
func (s *StaticArchiveService) Create(ctx context.Context, req *models.StaticArchiveCreateRequest) (*models.StaticArchiveCreated, error) {
	var created models.StaticArchiveCreated
	if err := s.client.PostOnce(ctx, "applications/archive", req, &created); err != nil {
		return nil, fmt.Errorf("failed to create static site application: %w", err)
	}
	return &created, nil
}

// Deploy deploys an uploaded archive to an existing static site application.
func (s *StaticArchiveService) Deploy(ctx context.Context, appUUID string, req *models.StaticArchiveDeployRequest) (*models.StaticArchiveDeployed, error) {
	var deployed models.StaticArchiveDeployed
	if err := s.client.PostOnce(ctx, "applications/"+url.PathEscape(appUUID)+"/archive", req, &deployed); err != nil {
		return nil, fmt.Errorf("failed to deploy site archive to application %s: %w", appUUID, err)
	}
	return &deployed, nil
}

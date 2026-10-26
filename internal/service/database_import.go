package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

const (
	// DefaultImportChunkSize matches the chunk size of the Coolify web UI and
	// stays well below the server's per-request PHP upload limit.
	DefaultImportChunkSize int64 = 10_000_000
	// MaxImportFileSize is the server-side limit for uploaded backups (10 GiB).
	MaxImportFileSize int64 = 10 * 1024 * 1024 * 1024
	// DefaultImportPollInterval is used by Wait when no interval is given.
	DefaultImportPollInterval = 3 * time.Second
)

// DatabaseImportTarget identifies the database an import applies to: either a
// standalone database or a database inside a service.
type DatabaseImportTarget struct {
	ServiceUUID  string // empty for standalone databases
	DatabaseUUID string
}

// StandaloneDatabaseImportTarget targets a standalone database.
func StandaloneDatabaseImportTarget(databaseUUID string) DatabaseImportTarget {
	return DatabaseImportTarget{DatabaseUUID: databaseUUID}
}

// ServiceDatabaseImportTarget targets a database that belongs to a service.
func ServiceDatabaseImportTarget(serviceUUID, databaseUUID string) DatabaseImportTarget {
	return DatabaseImportTarget{ServiceUUID: serviceUUID, DatabaseUUID: databaseUUID}
}

func (t DatabaseImportTarget) basePath() string {
	if t.ServiceUUID != "" {
		return fmt.Sprintf("services/%s/databases/%s/imports", t.ServiceUUID, t.DatabaseUUID)
	}
	return fmt.Sprintf("databases/%s/imports", t.DatabaseUUID)
}

// DatabaseImportService handles database backup uploads and imports for both
// standalone and service databases.
type DatabaseImportService struct {
	client *api.Client
	// ChunkSize is the upload chunk size in bytes.
	ChunkSize int64
}

// NewDatabaseImportService creates a new database import service.
func NewDatabaseImportService(client *api.Client) *DatabaseImportService {
	return &DatabaseImportService{client: client, ChunkSize: DefaultImportChunkSize}
}

// UploadProgress is called after every uploaded chunk.
type UploadProgress func(sent, total int64)

// Upload uploads a local backup file in chunks (DropZone protocol, as used by
// the Coolify UI) and returns the upload whose UploadID can start an import.
func (s *DatabaseImportService) Upload(ctx context.Context, target DatabaseImportTarget, filePath string, progress UploadProgress) (*models.DatabaseImportUpload, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat backup file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory", filePath)
	}
	size := info.Size()
	if size == 0 {
		return nil, fmt.Errorf("backup file %s is empty", filePath)
	}
	if size > MaxImportFileSize {
		return nil, fmt.Errorf("backup file %s exceeds the 10 GiB import limit", filePath)
	}

	chunkSize := s.ChunkSize
	if chunkSize <= 0 {
		chunkSize = DefaultImportChunkSize
	}
	totalChunks := (size + chunkSize - 1) / chunkSize

	uploadID, err := newUUIDv4()
	if err != nil {
		return nil, err
	}

	path := target.basePath() + "/uploads"
	fileName := filepath.Base(filePath)
	var upload models.DatabaseImportUpload
	for index := int64(0); index < totalChunks; index++ {
		offset := index * chunkSize
		length := min(chunkSize, size-offset)
		fields := map[string]string{
			"upload_id":         uploadID,
			"dzuuid":            uploadID,
			"dzchunkindex":      strconv.FormatInt(index, 10),
			"dztotalchunkcount": strconv.FormatInt(totalChunks, 10),
			"dztotalfilesize":   strconv.FormatInt(size, 10),
			"dzchunksize":       strconv.FormatInt(chunkSize, 10),
			"dzchunkbyteoffset": strconv.FormatInt(offset, 10),
		}
		part := api.MultipartFile{
			FieldName: "file",
			FileName:  fileName,
			Content:   io.NewSectionReader(file, offset, length),
			Size:      length,
		}

		// Intermediate chunks answer {upload_id, done, status}; only the final
		// chunk returns the assembled file's metadata.
		upload = models.DatabaseImportUpload{}
		if err := s.client.PostMultipart(ctx, path, fields, part, &upload); err != nil {
			return nil, fmt.Errorf("failed to upload chunk %d/%d: %w", index+1, totalChunks, err)
		}
		if progress != nil {
			progress(offset+length, size)
		}
	}

	if upload.Filename == "" {
		return nil, errors.New("upload did not complete: the server did not assemble the uploaded chunks")
	}
	return &upload, nil
}

// Start queues an import into the target database.
func (s *DatabaseImportService) Start(ctx context.Context, target DatabaseImportTarget, req *models.DatabaseImportRequest) (*models.DatabaseImportStarted, error) {
	var started models.DatabaseImportStarted
	if err := s.client.Post(ctx, target.basePath(), req, &started); err != nil {
		return nil, fmt.Errorf("failed to start import into database %s: %w", target.DatabaseUUID, err)
	}
	return &started, nil
}

// Status returns the status of an import activity.
func (s *DatabaseImportService) Status(ctx context.Context, target DatabaseImportTarget, activityID int) (*models.DatabaseImportStatus, error) {
	var status models.DatabaseImportStatus
	if err := s.client.Get(ctx, fmt.Sprintf("%s/%d", target.basePath(), activityID), &status); err != nil {
		return nil, fmt.Errorf("failed to get import %d status: %w", activityID, err)
	}
	return &status, nil
}

// Wait polls the import status until it reaches a terminal state and returns
// the final status.
func (s *DatabaseImportService) Wait(ctx context.Context, target DatabaseImportTarget, activityID int, interval time.Duration) (*models.DatabaseImportStatus, error) {
	if interval <= 0 {
		interval = DefaultImportPollInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		status, err := s.Status(ctx, target, activityID)
		if err != nil {
			return nil, err
		}
		if status.IsTerminal() {
			return status, nil
		}
		select {
		case <-ctx.Done():
			return status, ctx.Err()
		case <-ticker.C:
		}
	}
}

// newUUIDv4 returns a random RFC 4122 version 4 UUID.
func newUUIDv4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("failed to generate upload id: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

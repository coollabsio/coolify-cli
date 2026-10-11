package models

// Database import sources accepted by POST .../imports.
const (
	DatabaseImportSourceUpload = "upload"
	DatabaseImportSourceS3     = "s3"
	DatabaseImportSourceServer = "server"
)

// DatabaseImportUpload is the response of a completed backup upload
// (POST .../imports/uploads).
type DatabaseImportUpload struct {
	UploadID string `json:"upload_id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// DatabaseImportRequest starts an import (POST .../imports). Source is one of
// "upload" (needs UploadID), "s3" (needs S3StorageUUID + Path) or "server"
// (needs Path on the database's server).
type DatabaseImportRequest struct {
	Source            string  `json:"source"`
	UploadID          *string `json:"upload_id,omitempty"`
	S3StorageUUID     *string `json:"s3_storage_uuid,omitempty"`
	Path              *string `json:"path,omitempty"`
	DumpAll           *bool   `json:"dump_all,omitempty"`
	ReplaceExisting   *bool   `json:"replace_existing,omitempty"`
	KeepOwners        *bool   `json:"keep_owners,omitempty"`
	RestoreMysqlUsers *bool   `json:"restore_mysql_users,omitempty"`
	SqliteDatabase    *string `json:"sqlite_database,omitempty"`
}

// DatabaseImportStarted is the 202 response of POST .../imports.
type DatabaseImportStarted struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	StatusURL string `json:"status_url" table:"-"`
}

// DatabaseImportStatus is the response of GET .../imports/{activity_id}.
// Output is only returned for tokens that can read sensitive data.
type DatabaseImportStatus struct {
	ID         int     `json:"id"`
	Status     string  `json:"status"`
	ExitCode   *int    `json:"exit_code"`
	Output     *string `json:"output" table:"-"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
	FinishedAt *string `json:"finished_at"`
}

// IsTerminal reports whether the import has stopped running.
func (s *DatabaseImportStatus) IsTerminal() bool {
	switch s.Status {
	case "finished", "error", "killed", "cancelled", "closed":
		return true
	}
	return false
}

// Succeeded reports whether the import finished successfully.
func (s *DatabaseImportStatus) Succeeded() bool {
	return s.Status == "finished" && (s.ExitCode == nil || *s.ExitCode == 0)
}

package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strconv"

	"github.com/coollabsio/coolify-cli/internal/api"
)

// UploadProgress is called after every uploaded chunk.
type UploadProgress func(sent, total int64)

// uploadChunks uploads file in chunks with the DropZone protocol that the
// Coolify UI uses and returns the response of the final chunk. Intermediate
// chunks answer {upload_id, done, status}; only the final chunk returns the
// assembled file's metadata, so callers check T for completion.
func uploadChunks[T any](ctx context.Context, client *api.Client, path string, file *os.File, size, chunkSize int64, extraFields map[string]string, progress UploadProgress) (T, error) {
	var result T
	if chunkSize <= 0 {
		chunkSize = DefaultImportChunkSize
	}
	totalChunks := (size + chunkSize - 1) / chunkSize

	uploadID, err := newUUIDv4()
	if err != nil {
		return result, err
	}

	fileName := filepath.Base(file.Name())
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
		maps.Copy(fields, extraFields)
		part := api.MultipartFile{
			FieldName: "file",
			FileName:  fileName,
			Content:   io.NewSectionReader(file, offset, length),
			Size:      length,
		}

		var zero T
		result = zero
		if err := client.PostMultipart(ctx, path, fields, part, &result); err != nil {
			return result, fmt.Errorf("failed to upload chunk %d/%d: %w", index+1, totalChunks, err)
		}
		if progress != nil {
			progress(offset+length, size)
		}
	}
	return result, nil
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

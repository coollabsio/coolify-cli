package updater

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// VerifyChecksum validates the integrity of the downloaded asset
func VerifyChecksum(reader io.Reader, expectedChecksum string) (io.Reader, error) {
	// Read all bytes from the reader
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading data: %w", err)
	}

	// Calculate the SHA256 checksum
	hash := sha256.Sum256(data)
	actualChecksum := hex.EncodeToString(hash[:])

	// Compare checksums
	if actualChecksum != expectedChecksum {
		return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	// Return a new reader with the same data
	return bytes.NewReader(data), nil
}

// VerifyChecksumBytes validates the integrity of a byte slice
func VerifyChecksumBytes(data []byte, expectedChecksum string) error {
	// Calculate the SHA256 checksum
	hash := sha256.Sum256(data)
	actualChecksum := hex.EncodeToString(hash[:])

	// Compare checksums
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

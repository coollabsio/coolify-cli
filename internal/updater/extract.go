package updater

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ExtractTarGz extracts a tar.gz archive to a temporary directory
// and returns the path to the binary
func ExtractTarGz(archiveFile io.Reader, binaryName string) (string, error) {
	// Create a temporary directory to extract files
	tempDir, err := os.MkdirTemp("", "coolify-update")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Extract the tar.gz file
	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		if rmErr := os.RemoveAll(tempDir); rmErr != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w (cleanup failed: %v)", err, rmErr)
		}
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if closeErr := gzipReader.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Error closing gzip reader: %v\n", closeErr)
		}
	}()

	tarReader := tar.NewReader(gzipReader)

	// Extract binary from the archive
	binaryPath := ""
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			if rmErr := os.RemoveAll(tempDir); rmErr != nil {
				return "", fmt.Errorf("error reading tar: %w (cleanup failed: %v)", err, rmErr)
			}
			return "", fmt.Errorf("error reading tar: %w", err)
		}

		// Skip directories and non-binary files
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Check if this is the binary we're looking for
		baseName := filepath.Base(header.Name)
		if baseName == binaryName {
			// Create the output file
			outPath := filepath.Join(tempDir, baseName)
			outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_RDWR, 0o755)
			if err != nil {
				if rmErr := os.RemoveAll(tempDir); rmErr != nil {
					return "", fmt.Errorf("failed to create output file: %w (cleanup failed: %v)", err, rmErr)
				}
				return "", fmt.Errorf("failed to create output file: %w", err)
			}

			// Copy the file contents
			if _, err := io.Copy(outFile, tarReader); err != nil {
				if closeErr := outFile.Close(); closeErr != nil {
					fmt.Fprintf(os.Stderr, "Error closing output file: %v\n", closeErr)
				}
				if rmErr := os.RemoveAll(tempDir); rmErr != nil {
					return "", fmt.Errorf("failed to extract file: %w (cleanup failed: %v)", err, rmErr)
				}
				return "", fmt.Errorf("failed to extract file: %w", err)
			}
			if closeErr := outFile.Close(); closeErr != nil {
				if rmErr := os.RemoveAll(tempDir); rmErr != nil {
					return "", fmt.Errorf("failed to close output file: %w (cleanup failed: %v)", closeErr, rmErr)
				}
				return "", fmt.Errorf("failed to close output file: %w", closeErr)
			}

			binaryPath = outPath
		}
	}

	if binaryPath == "" {
		if rmErr := os.RemoveAll(tempDir); rmErr != nil {
			return "", fmt.Errorf("binary %s not found in archive (cleanup failed: %v)", binaryName, rmErr)
		}
		return "", fmt.Errorf("binary %s not found in archive", binaryName)
	}

	return binaryPath, nil
}

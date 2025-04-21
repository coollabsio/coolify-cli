package updater

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Updater is the main interface for updating the CLI
type Updater struct {
	githubUpdater *GithubUpdater
	owner         string
	repo          string
	binaryName    string
}

// New creates a new updater instance
func New(owner, repo, currentVersion string) *Updater {
	return &Updater{
		githubUpdater: NewGithubUpdater(owner, repo, currentVersion),
		owner:         owner,
		repo:          repo,
		binaryName:    "coolify",
	}
}

// Check checks if there's a newer version available without performing an update
func (u *Updater) Check(ctx context.Context, includePrerelease bool) (*ReleaseInfo, bool, error) {
	return u.githubUpdater.CheckForUpdate(ctx, includePrerelease)
}

// To updates the CLI to release version passed in
func (u *Updater) To(ctx context.Context, release *ReleaseInfo) (string, error) {
	// Download the asset
	assetReader, err := u.githubUpdater.DownloadAsset(ctx, release.AssetURL)
	if err != nil {
		return "", fmt.Errorf("error downloading asset: %w", err)
	}
	defer func() {
		if closeErr := assetReader.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Error closing asset reader: %v\n", closeErr)
		}
	}()

	// Read the asset data once so we can verify and extract
	assetData, err := io.ReadAll(assetReader)
	if err != nil {
		return "", fmt.Errorf("error reading asset data: %w", err)
	}

	// Verify checksum if available
	if release.ChecksumURL != "" {
		checksums, err := u.githubUpdater.DownloadChecksums(ctx, release.ChecksumURL)
		if err != nil {
			return "", fmt.Errorf("error downloading checksums: %w", err)
		}

		expectedChecksum, ok := checksums[release.AssetName]
		if ok {
			// Verify the checksum
			err = VerifyChecksumBytes(assetData, expectedChecksum)
			if err != nil {
				return "", fmt.Errorf("checksum verification failed: %w", err)
			}
		}
	}

	// Extract the binary
	tempBinaryPath, err := ExtractTarGz(bytes.NewReader(assetData), u.binaryName)
	if err != nil {
		return "", fmt.Errorf("error extracting binary: %w", err)
	}

	// Clean up the temporary directory when we're done
	tempDir := filepath.Dir(tempBinaryPath)
	defer func() {
		if rmErr := os.RemoveAll(tempDir); rmErr != nil {
			fmt.Fprintf(os.Stderr, "Error removing temp directory: %v\n", rmErr)
		}
	}()

	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("error getting executable path: %w", err)
	}

	// Replace the binary
	if err := Replace(tempBinaryPath, executablePath); err != nil {
		return "", fmt.Errorf("error replacing binary: %w", err)
	}

	return release.Version, nil
}

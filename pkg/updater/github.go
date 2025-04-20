package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/v30/github"
	"github.com/hashicorp/go-version"
	"golang.org/x/oauth2"
)

const (
	DownloadTimeout = 5 * time.Minute
	BinaryName      = "coolify"
)

// ReleaseInfo contains information about a GitHub release
type ReleaseInfo struct {
	Version       string
	AssetURL      string
	AssetName     string
	ChecksumURL   string
	PublishedDate time.Time
	PreRelease    bool
	Notes         string
}

// GithubUpdater handles interaction with GitHub releases
type GithubUpdater struct {
	client         *github.Client
	httpClient     *http.Client
	owner          string
	repo           string
	binaryName     string
	currentVersion string
}

// NewGithubUpdater creates a new GitHub updater with appropriate configuration
func NewGithubUpdater(owner, repo, currentVersion string) *GithubUpdater {
	httpClient := &http.Client{
		Timeout: DownloadTimeout,
	}

	// Use GitHub token if available for better rate limits
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		httpClient = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
	}

	return &GithubUpdater{
		client:         github.NewClient(httpClient),
		httpClient:     httpClient,
		owner:          owner,
		repo:           repo,
		binaryName:     BinaryName,
		currentVersion: currentVersion,
	}
}

// DetectLatest finds the latest available release
func (g *GithubUpdater) DetectLatest(ctx context.Context, includePrerelease bool) (*ReleaseInfo, error) {
	var release *github.RepositoryRelease
	var err error

	if includePrerelease {
		// List all releases including pre-releases
		releases, _, err := g.client.Repositories.ListReleases(ctx, g.owner, g.repo, &github.ListOptions{PerPage: 10})
		if err != nil {
			return nil, fmt.Errorf("error listing releases: %w", err)
		}
		if len(releases) == 0 {
			return nil, fmt.Errorf("no releases found for %s/%s", g.owner, g.repo)
		}
		release = releases[0]
	} else {
		// Get only the latest stable release
		release, _, err = g.client.Repositories.GetLatestRelease(ctx, g.owner, g.repo)
		if err != nil {
			return nil, fmt.Errorf("error getting latest release: %w", err)
		}
	}

	if release == nil {
		return nil, fmt.Errorf("no release found for %s/%s", g.owner, g.repo)
	}

	assetName, assetURL, err := g.findMatchingAsset(release)
	if err != nil {
		return nil, err
	}

	checksumURL := ""
	for _, asset := range release.Assets {
		if strings.Contains(asset.GetName(), "checksums.txt") {
			checksumURL = asset.GetBrowserDownloadURL()
			break
		}
	}

	versionStr := strings.TrimPrefix(release.GetTagName(), "v")

	return &ReleaseInfo{
		Version:       versionStr,
		AssetURL:      assetURL,
		AssetName:     assetName,
		ChecksumURL:   checksumURL,
		PublishedDate: release.GetPublishedAt().Time,
		PreRelease:    release.GetPrerelease(),
		Notes:         release.GetBody(),
	}, nil
}

// findMatchingAsset finds the appropriate release asset for the current platform
func (g *GithubUpdater) findMatchingAsset(release *github.RepositoryRelease) (assetName, assetURL string, err error) {
	// Construct the asset name pattern based on current platform
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Based on goreleaser template from .goreleaser.yml
	assetNamePattern := fmt.Sprintf("%s_%s_%s_%s.tar.gz", g.binaryName,
		strings.TrimPrefix(release.GetTagName(), "v"), platform, arch)

	for _, asset := range release.Assets {
		if asset.GetName() == assetNamePattern {
			assetName = asset.GetName()
			assetURL = asset.GetBrowserDownloadURL()
			return assetName, assetURL, nil
		}
	}

	return "", "", fmt.Errorf("no matching asset found for %s/%s", platform, arch)
}

// DownloadAsset downloads an asset from GitHub
func (g *GithubUpdater) DownloadAsset(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error downloading asset: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return nil, fmt.Errorf("unexpected status code: %d (failed to close response: %v)", resp.StatusCode, closeErr)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// CheckForUpdate checks if a newer version is available
func (g *GithubUpdater) CheckForUpdate(ctx context.Context, includePrerelease bool) (*ReleaseInfo, bool, error) {
	latest, err := g.DetectLatest(ctx, includePrerelease)
	if err != nil {
		return nil, false, err
	}

	currentVer, err := version.NewVersion(g.currentVersion)
	if err != nil {
		return nil, false, fmt.Errorf("invalid current version: %w", err)
	}

	latestVer, err := version.NewVersion(latest.Version)
	if err != nil {
		return nil, false, fmt.Errorf("invalid latest version: %w", err)
	}

	return latest, currentVer.LessThan(latestVer), nil
}

// DownloadChecksums downloads the checksums file
func (g *GithubUpdater) DownloadChecksums(ctx context.Context, url string) (map[string]string, error) {
	if url == "" {
		return nil, fmt.Errorf("no checksums URL provided")
	}

	body, err := g.DownloadAsset(ctx, url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Error closing checksum response: %v\n", closeErr)
		}
	}()

	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("error reading checksums: %w", err)
	}

	checksums := make(map[string]string)
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		checksum := parts[0]
		filename := parts[1]

		checksums[filename] = checksum
	}

	return checksums, nil
}

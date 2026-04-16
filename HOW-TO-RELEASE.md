# How to Release Coolify CLI

This guide explains the release process for the Coolify CLI.

## Prerequisites

- Write access to the `coollabsio/coolify-cli` repository
- All changes merged to the target branch (`v4.x`)
- All tests passing (`go test ./internal/...`)

## Release Process

### 1. Create a GitHub Release

1. Go to https://github.com/coollabsio/coolify-cli/releases/new
2. Click "Choose a tag" and create a new tag:
   - **Tag name**: `v1.x.x` (must start with `v`, e.g., `v1.2.3`)
   - **Target**: `v4.x` (or your target branch)
3. **Release title**: `v1.x.x` (same as tag name)
4. **Description**: Write release notes describing:
   - New features
   - Bug fixes
   - Breaking changes (if any)
   - Example:
     ```markdown
     ## What's New
     - Added support for database management
     - Improved error messages for API failures

     ## Bug Fixes
     - Fixed panic when config file is missing

     ## Breaking Changes
     - None
     ```
5. Click "Publish release"

### 2. Automated Build Process

Once you publish the release:

1. GitHub Actions automatically triggers the `release-cli.yml` workflow
2. GoReleaser builds binaries for:
   - **Linux**: amd64, arm64
   - **macOS (Darwin)**: amd64, arm64
   - **Windows**: amd64, arm64
3. Goreleaser injects the version from the tag into the binaries via ldflags (into `internal/version.version`)
4. Binaries are automatically uploaded to the release
5. A follow-up `update-version` job then:
   - Updates the `version` constant in `internal/version/checker.go` to the new tag
   - Commits the bump to `v4.x` as `chore: bump version to vX.Y.Z`
   - Force-moves the release tag to point at that new commit
6. The release becomes available at:
   - GitHub: `https://github.com/coollabsio/coolify-cli/releases/tag/v1.x.x`
   - Install script: `curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash`
   - `go install`: `go install github.com/coollabsio/coolify-cli/coolify@v1.x.x`

### 3. Verify the Release

After the workflow completes (usually 2-5 minutes):

1. Check the release page has all platform binaries
2. Test the install script:
   ```bash
   curl -fsSL https://cdn.coollabs.io/coolify/install.sh | bash
   coolify version
   ```
3. Test the auto-update functionality:
   ```bash
   # If you have an older version installed
   coolify update
   coolify version  # Should show the new version
   ```
4. Verify the version matches your release

## Troubleshooting

### Build Failed
- Check the GitHub Actions logs at https://github.com/coollabsio/coolify-cli/actions
- Common issues:
  - Syntax errors in Go code
  - Test failures
  - GoReleaser configuration issues

### Version Not Updating
- The version is injected at build time via ldflags into `internal/version.version` — you do **not** need to edit it manually before releasing. The post-release `update-version` job also rewrites `internal/version/checker.go` on `v4.x`.
- If the hardcoded fallback in `internal/version/checker.go` is stale, check that the `update-version` job ran successfully after the release.
- The tag must start with `v` (e.g., `v1.2.3`, not `1.2.3`)
- Check that the workflow has write permissions (`contents: write` in `release-cli.yml`)

### Install Script Not Finding New Version
- Wait a few minutes for GitHub's CDN to update
- Check that binaries were uploaded to the release
- Verify the tag format is correct (`v1.x.x`)

## Release Checklist

Before creating a release:

- [ ] All tests pass: `go test ./internal/...`
- [ ] Code is formatted: `go fmt ./...`
- [ ] Changes merged to `v4.x` branch
- [ ] Release notes prepared

> Note: You do **not** need to bump the version manually. GoReleaser injects the tag version via ldflags, and the `update-version` CI job commits the bump to `internal/version/checker.go` after the release.

After creating a release:

- [ ] GitHub Actions workflow completed successfully
- [ ] All platform binaries are present on the release page
- [ ] Install script downloads the new version
- [ ] `coolify version` returns the correct version

## Configuration Files

The release process uses these configuration files:

- `.goreleaser.yml` - GoReleaser configuration (build matrix, archives, etc.) - points to `/coolify` as entry point
- `.github/workflows/release-cli.yml` - GitHub Actions workflow
- `scripts/install.sh` - User-facing install script
- `internal/version/checker.go` - Contains `GetVersion()` function that returns the current version
- `coolify/main.go` - Binary entry point for `go install` support

## Notes

- The CLI has auto-update checking built-in (checks every 10 minutes)
- Users can manually update with `coolify update`
- Install script supports version pinning: `bash install.sh v1.2.3`
- Releases are immutable - if you need to fix something, create a new patch version

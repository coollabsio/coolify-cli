package version

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() stdout error = %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() stderr error = %v", err)
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	fn()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, stdoutR)
	_, _ = io.Copy(&stderrBuf, stderrR)

	return stdoutBuf.String(), stderrBuf.String()
}

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Error("GetVersion() returned empty string")
	}
	// Version should start with 'v'
	if v[0] != 'v' {
		t.Errorf("GetVersion() = %q, expected to start with 'v'", v)
	}
}

func TestCheckLatestVersionOfCli_UpdateAvailable(t *testing.T) {
	// Save original values
	originalURL := GitHubAPIURL
	originalVersion := version
	defer func() {
		GitHubAPIURL = originalURL
		version = originalVersion
	}()

	// Set a low version to ensure update is available
	version = "v0.0.1"

	// Create mock server with newer version
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Return tags in GitHub API format
		_, _ = w.Write([]byte(`[{"ref":"refs/tags/v1.0.0"},{"ref":"refs/tags/v2.0.0"}]`))
	}))
	defer server.Close()

	GitHubAPIURL = server.URL

	var latestVersion string
	var err error
	stdout, stderr := captureOutput(t, func() {
		latestVersion, err = CheckLatestVersionOfCli(false)
	})

	if err != nil {
		t.Errorf("CheckLatestVersionOfCli() error = %v, want nil", err)
	}

	if latestVersion != "2.0.0" {
		t.Errorf("CheckLatestVersionOfCli() latestVersion = %q, want %q", latestVersion, "2.0.0")
	}

	// Should not write anything to stdout
	if stdout != "" {
		t.Errorf("CheckLatestVersionOfCli() stdout = %q, want empty string", stdout)
	}

	// Should print update message to stderr
	expectedMsg := "A new version (2.0.0) is available. Update with: coolify update\n"
	if stderr != expectedMsg {
		t.Errorf("CheckLatestVersionOfCli() stderr = %q, want %q", stderr, expectedMsg)
	}
}

func TestCheckLatestVersionOfCli_NoUpdate(t *testing.T) {
	// Save original values
	originalURL := GitHubAPIURL
	originalVersion := version
	defer func() {
		GitHubAPIURL = originalURL
		version = originalVersion
	}()

	// Set a high version to ensure no update is available
	version = "v99.99.99"

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"ref":"refs/tags/v1.0.0"},{"ref":"refs/tags/v2.0.0"}]`))
	}))
	defer server.Close()

	GitHubAPIURL = server.URL

	var latestVersion string
	var err error
	stdout, stderr := captureOutput(t, func() {
		latestVersion, err = CheckLatestVersionOfCli(false)
	})

	if err != nil {
		t.Errorf("CheckLatestVersionOfCli() error = %v, want nil", err)
	}

	// Function returns the latest version from GitHub (2.0.0), not the current version
	if latestVersion != "2.0.0" {
		t.Errorf("CheckLatestVersionOfCli() latestVersion = %q, want %q", latestVersion, "2.0.0")
	}

	// Should NOT print any message when already on latest (current v99.99.99 > latest v2.0.0)
	if stdout != "" {
		t.Errorf("CheckLatestVersionOfCli() should not write to stdout when on latest version, got: %q", stdout)
	}

	if stderr != "" {
		t.Errorf("CheckLatestVersionOfCli() should not write to stderr when on latest version, got: %q", stderr)
	}
}

func TestCheckLatestVersionOfCli_APIError_SilentFail(t *testing.T) {
	// Save original URL
	originalURL := GitHubAPIURL
	defer func() {
		GitHubAPIURL = originalURL
	}()

	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	GitHubAPIURL = server.URL

	var latestVersion string
	var err error
	stdout, stderr := captureOutput(t, func() {
		latestVersion, err = CheckLatestVersionOfCli(false)
	})

	// Should return empty string and nil error (silent fail)
	if err != nil {
		t.Errorf("CheckLatestVersionOfCli() error = %v, want nil on API error", err)
	}

	if latestVersion != "" {
		t.Errorf("CheckLatestVersionOfCli() latestVersion = %q, want empty string on API error", latestVersion)
	}

	// Should NOT print anything on error
	if stdout != "" {
		t.Errorf("CheckLatestVersionOfCli() should not print anything to stdout on API error, got: %q", stdout)
	}

	if stderr != "" {
		t.Errorf("CheckLatestVersionOfCli() should not print anything to stderr on API error, got: %q", stderr)
	}
}

func TestCheckLatestVersionOfCli_NetworkError_SilentFail(t *testing.T) {
	// Save original URL
	originalURL := GitHubAPIURL
	defer func() {
		GitHubAPIURL = originalURL
	}()

	// Use invalid URL to cause network error
	GitHubAPIURL = "http://localhost:1" // Port 1 should fail to connect

	var latestVersion string
	var err error
	stdout, stderr := captureOutput(t, func() {
		latestVersion, err = CheckLatestVersionOfCli(false)
	})

	// Should return empty string and nil error (silent fail)
	if err != nil {
		t.Errorf("CheckLatestVersionOfCli() error = %v, want nil on network error", err)
	}

	if latestVersion != "" {
		t.Errorf("CheckLatestVersionOfCli() latestVersion = %q, want empty string on network error", latestVersion)
	}

	// Should NOT print anything on error
	if stdout != "" {
		t.Errorf("CheckLatestVersionOfCli() should not print anything to stdout on network error, got: %q", stdout)
	}

	if stderr != "" {
		t.Errorf("CheckLatestVersionOfCli() should not print anything to stderr on network error, got: %q", stderr)
	}
}

func TestCheckLatestVersionOfCli_UpdateAvailable_LeavesStdoutAvailableForJSON(t *testing.T) {
	originalURL := GitHubAPIURL
	originalVersion := version
	defer func() {
		GitHubAPIURL = originalURL
		version = originalVersion
	}()

	version = "v0.0.1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"ref":"refs/tags/v2.0.0"}]`))
	}))
	defer server.Close()

	GitHubAPIURL = server.URL

	stdout, stderr := captureOutput(t, func() {
		_, _ = CheckLatestVersionOfCli(false)
		_, _ = os.Stdout.WriteString(`[{"uuid":"demo"}]` + "\n")
	})

	expectedStdout := `[{"uuid":"demo"}]` + "\n"
	if stdout != expectedStdout {
		t.Fatalf("stdout = %q, want %q", stdout, expectedStdout)
	}

	expectedStderr := "A new version (2.0.0) is available. Update with: coolify update\n"
	if stderr != expectedStderr {
		t.Fatalf("stderr = %q, want %q", stderr, expectedStderr)
	}
}

func TestCheckLatestVersionOfCli_InvalidJSON_SilentFail(t *testing.T) {
	// Save original URL
	originalURL := GitHubAPIURL
	defer func() {
		GitHubAPIURL = originalURL
	}()

	// Create mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	GitHubAPIURL = server.URL

	latestVersion, err := CheckLatestVersionOfCli(false)

	// Should return empty string and nil error (silent fail)
	if err != nil {
		t.Errorf("CheckLatestVersionOfCli() error = %v, want nil on invalid JSON", err)
	}

	if latestVersion != "" {
		t.Errorf("CheckLatestVersionOfCli() latestVersion = %q, want empty string on invalid JSON", latestVersion)
	}
}

func TestCheckLatestVersionOfCli_EmptyTags_SilentFail(t *testing.T) {
	// Save original URL
	originalURL := GitHubAPIURL
	defer func() {
		GitHubAPIURL = originalURL
	}()

	// Create mock server that returns empty array
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	GitHubAPIURL = server.URL

	latestVersion, err := CheckLatestVersionOfCli(false)

	// Should return empty string and nil error (silent fail)
	if err != nil {
		t.Errorf("CheckLatestVersionOfCli() error = %v, want nil on empty tags", err)
	}

	if latestVersion != "" {
		t.Errorf("CheckLatestVersionOfCli() latestVersion = %q, want empty string on empty tags", latestVersion)
	}
}

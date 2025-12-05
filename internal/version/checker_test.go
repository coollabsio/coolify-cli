package version

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

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

	// Capture stdout to check for update message
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	latestVersion, err := CheckLatestVersionOfCli(false)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Errorf("CheckLatestVersionOfCli() error = %v, want nil", err)
	}

	if latestVersion != "2.0.0" {
		t.Errorf("CheckLatestVersionOfCli() latestVersion = %q, want %q", latestVersion, "2.0.0")
	}

	// Should print update message
	expectedMsg := "A new version (2.0.0) is available. Update with: coolify update\n"
	if output != expectedMsg {
		t.Errorf("CheckLatestVersionOfCli() output = %q, want %q", output, expectedMsg)
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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	latestVersion, err := CheckLatestVersionOfCli(false)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Errorf("CheckLatestVersionOfCli() error = %v, want nil", err)
	}

	// Function returns the latest version from GitHub (2.0.0), not the current version
	if latestVersion != "2.0.0" {
		t.Errorf("CheckLatestVersionOfCli() latestVersion = %q, want %q", latestVersion, "2.0.0")
	}

	// Should NOT print any message when already on latest (current v99.99.99 > latest v2.0.0)
	if output != "" {
		t.Errorf("CheckLatestVersionOfCli() should not print anything when on latest version, got: %q", output)
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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	latestVersion, err := CheckLatestVersionOfCli(false)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should return empty string and nil error (silent fail)
	if err != nil {
		t.Errorf("CheckLatestVersionOfCli() error = %v, want nil on API error", err)
	}

	if latestVersion != "" {
		t.Errorf("CheckLatestVersionOfCli() latestVersion = %q, want empty string on API error", latestVersion)
	}

	// Should NOT print anything on error
	if output != "" {
		t.Errorf("CheckLatestVersionOfCli() should not print anything on API error, got: %q", output)
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

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	latestVersion, err := CheckLatestVersionOfCli(false)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should return empty string and nil error (silent fail)
	if err != nil {
		t.Errorf("CheckLatestVersionOfCli() error = %v, want nil on network error", err)
	}

	if latestVersion != "" {
		t.Errorf("CheckLatestVersionOfCli() latestVersion = %q, want empty string on network error", latestVersion)
	}

	// Should NOT print anything on error
	if output != "" {
		t.Errorf("CheckLatestVersionOfCli() should not print anything on network error, got: %q", output)
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

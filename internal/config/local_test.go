package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindLocal(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}

	// No file anywhere up the tree (TempDir parents have none).
	local, err := FindLocal(nested)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if local != nil {
		t.Fatalf("expected no local config, got %+v", local)
	}

	// Found by walking up from a nested dir.
	rootFile := filepath.Join(root, ".coolify.json")
	if err := os.WriteFile(rootFile, []byte(`{"context":"production"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	local, err = FindLocal(nested)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if local == nil || local.Context != "production" || local.Path != rootFile {
		t.Fatalf("got %+v, want context=production path=%s", local, rootFile)
	}

	// Nearest file wins.
	nestedFile := filepath.Join(nested, ".coolifyrc")
	if err := os.WriteFile(nestedFile, []byte(`{"context":"staging"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	local, err = FindLocal(nested)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if local.Context != "staging" || local.Path != nestedFile {
		t.Fatalf("got %+v, want context=staging path=%s", local, nestedFile)
	}

	// Malformed JSON is an error, not a silent fallback.
	if err := os.WriteFile(nestedFile, []byte("context = production"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := FindLocal(nested); err == nil {
		t.Fatal("expected parse error for malformed local config")
	}
}

func TestLocalContext(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".coolify.json"), []byte(`{"context":"prod"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	name, path, err := LocalContext()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "prod" || path == "" {
		t.Fatalf("got name=%q path=%q, want name=prod and a path", name, path)
	}
}

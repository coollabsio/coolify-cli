package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LocalFileNames are the repo-local config file names, checked in this order in
// each directory while walking up from the working directory.
var LocalFileNames = []string{".coolify.json", ".coolifyrc"}

// LocalConfig is the repo-local config. It only points at a context that
// already exists in the global config - credentials stay global, never in the
// repo.
type LocalConfig struct {
	Context string `json:"context"`

	// Path is the file this was read from (not serialized).
	Path string `json:"-"`
}

// FindLocal walks up from dir looking for a repo-local config file. Returns nil
// when none is found.
func FindLocal(dir string) (*LocalConfig, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	for {
		for _, name := range LocalFileNames {
			path := filepath.Join(dir, name)
			if !fileExists(path) {
				continue
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", path, err)
			}

			var local LocalConfig
			if err := json.Unmarshal(data, &local); err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", path, err)
			}
			local.Path = path
			return &local, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, nil
		}
		dir = parent
	}
}

// LocalContext returns the context name declared by the nearest repo-local
// config file, plus the file it came from. Both are empty when there is none.
func LocalContext() (name, path string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	local, err := FindLocal(cwd)
	if err != nil || local == nil {
		return "", "", err
	}
	return local.Context, local.Path, nil
}

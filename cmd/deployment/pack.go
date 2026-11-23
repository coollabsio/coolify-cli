package deployment

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// skippedNames are never packed, at any depth.
var skippedNames = map[string]bool{
	".git":         true,
	"node_modules": true,
	linkFileName:   true,
	".DS_Store":    true,
}

// packedSite is a gzip tar of a folder in a temporary file. Remove it with
// os.Remove(Path) when done.
type packedSite struct {
	Path      string
	Files     int
	HasIndex  bool
	TotalSize int64
}

// packFolder writes the regular files and directories of dir into a gzip tar
// with paths relative to dir. Symlinks and other special files are rejected,
// so the archive holds only what is inside the folder.
func packFolder(dir string) (_ *packedSite, err error) {
	root, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, fmt.Errorf("folder %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp("", "coolify-site-*.tar.gz")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary archive: %w", err)
	}
	site := &packedSite{Path: tmp.Name()}
	defer func() {
		if closeErr := tmp.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("failed to write archive: %w", closeErr)
		}
		if err != nil {
			os.Remove(site.Path)
		}
	}()

	gz := gzip.NewWriter(tmp)
	tw := tar.NewWriter(gz)
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if skippedNames[entry.Name()] {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)

		mode := entry.Type()
		switch {
		case mode&fs.ModeSymlink != 0:
			return fmt.Errorf("%s is a symlink; replace it with the file or folder it points to", filepath.Join(dir, rel))
		case !mode.IsDir() && !mode.IsRegular():
			return fmt.Errorf("%s is not a regular file or folder", filepath.Join(dir, rel))
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = name
		header.Uid, header.Gid, header.Uname, header.Gname = 0, 0, "", ""
		if entry.IsDir() {
			header.Name += "/"
			return tw.WriteHeader(header)
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err := io.Copy(tw, file); err != nil {
			return fmt.Errorf("failed to pack %s: %w", name, err)
		}
		site.Files++
		site.TotalSize += info.Size()
		if name == "index.html" {
			site.HasIndex = true
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	if site.Files == 0 {
		return nil, fmt.Errorf("folder %s has no files to deploy", dir)
	}
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to write archive: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to write archive: %w", err)
	}
	return site, nil
}

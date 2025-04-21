package updater

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Replace replaces the current executable with the new version
func Replace(newBinaryPath, executablePath string) error {
	// Get executable path if not provided
	if executablePath == "" {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error getting executable path: %w", err)
		}
		executablePath = exe
	}

	// Different replacement strategies based on OS
	if runtime.GOOS == "windows" {
		return replaceWindowsExecutable(newBinaryPath, executablePath)
	}

	return replaceUnixExecutable(newBinaryPath, executablePath)
}

// replaceUnixExecutable replaces the executable on Unix systems
func replaceUnixExecutable(newBinaryPath, executablePath string) error {
	// Open the new binary file
	newFile, err := os.Open(newBinaryPath)
	if err != nil {
		return fmt.Errorf("error opening new binary: %w", err)
	}
	defer func() {
		if closeErr := newFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Error closing new binary file: %v\n", closeErr)
		}
	}()

	// Create a temporary file in the same directory
	dir := filepath.Dir(executablePath)
	tempFile, err := os.CreateTemp(dir, "coolify-update-*")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Copy the new binary to the temporary file
	if _, err := io.Copy(tempFile, newFile); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Error closing temp file: %v\n", closeErr)
		}
		if rmErr := os.Remove(tempPath); rmErr != nil {
			return fmt.Errorf("error copying new binary: %w (cleanup failed: %v)", err, rmErr)
		}
		return fmt.Errorf("error copying new binary: %w", err)
	}
	if closeErr := tempFile.Close(); closeErr != nil {
		if rmErr := os.Remove(tempPath); rmErr != nil {
			return fmt.Errorf("error closing temp file: %w (cleanup failed: %v)", closeErr, rmErr)
		}
		return fmt.Errorf("error closing temp file: %w", closeErr)
	}

	// Set permissions to match the current executable
	info, err := os.Stat(executablePath)
	if err != nil {
		if rmErr := os.Remove(tempPath); rmErr != nil {
			return fmt.Errorf("error getting executable info: %w (cleanup failed: %v)", err, rmErr)
		}
		return fmt.Errorf("error getting executable info: %w", err)
	}

	if err := os.Chmod(tempPath, info.Mode()); err != nil {
		if rmErr := os.Remove(tempPath); rmErr != nil {
			return fmt.Errorf("error setting permissions: %w (cleanup failed: %v)", err, rmErr)
		}
		return fmt.Errorf("error setting permissions: %w", err)
	}

	// Rename the temporary file to the executable path
	if err := os.Rename(tempPath, executablePath); err != nil {
		if rmErr := os.Remove(tempPath); rmErr != nil {
			return fmt.Errorf("error replacing executable: %w (cleanup failed: %v)", err, rmErr)
		}
		return fmt.Errorf("error replacing executable: %w", err)
	}

	return nil
}

// replaceWindowsExecutable replaces the executable on Windows
func replaceWindowsExecutable(newBinaryPath, executablePath string) error {
	// Define paths
	dir := filepath.Dir(executablePath)
	base := filepath.Base(executablePath)
	oldPath := filepath.Join(dir, "."+base+".old")

	// Step 1: Rename the target to .target.old
	// If an old backup exists from a previous update, try to remove it first
	_ = os.Remove(oldPath) // Ignore errors if file doesn't exist

	if err := os.Rename(executablePath, oldPath); err != nil {
		return fmt.Errorf("error renaming executable to backup: %w", err)
	}

	// Step 2: Rename the new binary (.target.new) to target
	if err := os.Rename(newBinaryPath, executablePath); err != nil {
		// Step 4: If rename fails, attempt to roll back
		rollbackErr := os.Rename(oldPath, executablePath)
		if rollbackErr != nil {
			return fmt.Errorf("failed to replace executable (%w) and rollback also failed (%v)", err, rollbackErr)
		}
		return fmt.Errorf("error replacing executable (rollback successful): %w", err)
	}

	// Step 3: On Windows, we can't easily delete the old file,
	// so instead we just make it a hidden file
	// The actual implementation is in hidewindows_windows.go for Windows
	// and hidewindows_other.go for non-Windows platforms
	if err := hideWindowsFile(oldPath); err != nil {
		// Non-fatal error - log but don't fail the update
		fmt.Fprintf(os.Stderr, "Warning: Failed to hide old executable: %v\n", err)
	}

	return nil
}

// startProcess starts a new process and returns immediately
func startProcess(command string, args []string) error {
	attr := &os.ProcAttr{
		Dir:   filepath.Dir(command),
		Env:   os.Environ(),
		Files: []*os.File{nil, nil, nil},
	}

	process, err := os.StartProcess(command, append([]string{command}, args...), attr)
	if err != nil {
		return err
	}

	// Detach from the process
	if releaseErr := process.Release(); releaseErr != nil {
		return fmt.Errorf("failed to release process: %w", releaseErr)
	}

	return nil
}

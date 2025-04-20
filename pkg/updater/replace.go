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
	// Windows can't replace a running executable, so we need to use a rename script
	// that runs after the current process exits

	// Create a temporary batch file that will move the new binary
	batchFile, err := os.CreateTemp("", "coolify-update-*.bat")
	if err != nil {
		return fmt.Errorf("error creating batch file: %w", err)
	}
	defer func() {
		if closeErr := batchFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Error closing batch file: %v\n", closeErr)
		}
	}()

	batchPath := batchFile.Name()

	// Write a batch script that:
	// 1. Waits for the current process to exit
	// 2. Copies the new binary over the old one
	// 3. Deletes itself
	script := fmt.Sprintf(`
@echo off
:wait
ping -n 2 127.0.0.1 > nul
tasklist /fi "PID eq %d" | find "%d" > nul
if not errorlevel 1 goto wait
copy /y "%s" "%s"
del "%s"
`, os.Getpid(), os.Getpid(), newBinaryPath, executablePath, batchPath)

	if _, err := batchFile.WriteString(script); err != nil {
		if rmErr := os.Remove(batchPath); rmErr != nil {
			return fmt.Errorf("error writing batch file: %w (cleanup failed: %v)", err, rmErr)
		}
		return fmt.Errorf("error writing batch file: %w", err)
	}

	// Start the batch file
	if err := startProcess(batchPath, []string{}); err != nil {
		if rmErr := os.Remove(batchPath); rmErr != nil {
			return fmt.Errorf("error starting batch file: %w (cleanup failed: %v)", err, rmErr)
		}
		return fmt.Errorf("error starting batch file: %w", err)
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

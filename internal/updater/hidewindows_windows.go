//go:build windows
// +build windows

package updater

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// hideWindowsFile sets the hidden attribute on a Windows file
func hideWindowsFile(path string) error {
	// Convert to UTF16 for Windows API
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("error converting path to UTF16: %w", err)
	}

	// Get current attributes
	attrs, err := windows.GetFileAttributes(pathPtr)
	if err != nil {
		return fmt.Errorf("error getting file attributes: %w", err)
	}

	// Add hidden attribute
	attrs |= windows.FILE_ATTRIBUTE_HIDDEN

	// Set new attributes
	err = windows.SetFileAttributes(pathPtr, attrs)
	if err != nil {
		return fmt.Errorf("error setting file attributes: %w", err)
	}

	return nil
}

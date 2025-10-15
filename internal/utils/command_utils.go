package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetCommandExample generates example usage strings using the actual binary name
// rather than hardcoding it. This makes examples resistant to binary name changes.
func GetCommandExample(format string, args ...interface{}) string {
	binaryName := getBinaryName()
	return fmt.Sprintf(format, append([]interface{}{binaryName}, args...)...)
}

// getBinaryName returns the name of the current executable without path
func getBinaryName() string {
	exe, err := os.Executable()
	if err != nil {
		return "coolify-cli" // Fallback to the default name
	}
	return filepath.Base(exe)
}

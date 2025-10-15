//go:build !windows
// +build !windows

package updater

import (
	"fmt"
)

// hideWindowsFile is a no-op on non-Windows systems
func hideWindowsFile(path string) error {
	// For non-Windows systems, we don't need to do anything special
	fmt.Printf("Note: Old executable backup at %s\n", path)
	return nil
}

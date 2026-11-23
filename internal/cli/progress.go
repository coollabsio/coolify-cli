package cli

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// IsTerminal reports whether w is a terminal, so that progress lines with
// carriage returns are only drawn for people, never into pipes or logs.
func IsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

// HumanBytes formats n bytes with binary units, for example "1.5 KiB".
func HumanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

package output

import (
	"fmt"
	"io"
	"os"
)

// Format types
const (
	FormatTable  = "table"
	FormatJSON   = "json"
	FormatPretty = "pretty"
)

// Formatter is the interface for output formatting
type Formatter interface {
	// Format formats the data and writes it to the writer
	Format(data interface{}) error
}

// Options for formatter configuration
type Options struct {
	Writer        io.Writer
	ShowSensitive bool
	Color         bool
}

// NewFormatter creates a formatter based on the format type
func NewFormatter(format string, opts Options) (Formatter, error) {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	switch format {
	case FormatTable:
		return NewTableFormatter(opts), nil
	case FormatJSON:
		return NewJSONFormatter(opts), nil
	case FormatPretty:
		return NewPrettyFormatter(opts), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// SensitiveOverlay is the string used to hide sensitive information
const SensitiveOverlay = "********"

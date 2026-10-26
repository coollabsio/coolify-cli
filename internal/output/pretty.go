package output

import (
	"encoding/json"
)

// PrettyFormatter formats output as indented JSON
type PrettyFormatter struct {
	opts Options
}

// NewPrettyFormatter creates a new pretty JSON formatter
func NewPrettyFormatter(opts Options) *PrettyFormatter {
	return &PrettyFormatter{opts: opts}
}

// Format formats the data as indented JSON. Fields tagged sensitive:"true" are masked unless ShowSensitive is set.
func (f *PrettyFormatter) Format(data interface{}) error {
	encoder := json.NewEncoder(f.opts.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(visibleData(data, f.opts))
}

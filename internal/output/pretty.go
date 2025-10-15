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

// Format formats the data as indented JSON
func (f *PrettyFormatter) Format(data interface{}) error {
	encoder := json.NewEncoder(f.opts.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

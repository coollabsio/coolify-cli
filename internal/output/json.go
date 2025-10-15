package output

import (
	"encoding/json"
)

// JSONFormatter formats output as compact JSON
type JSONFormatter struct {
	opts Options
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(opts Options) *JSONFormatter {
	return &JSONFormatter{opts: opts}
}

// Format formats the data as compact JSON
func (f *JSONFormatter) Format(data interface{}) error {
	encoder := json.NewEncoder(f.opts.Writer)
	return encoder.Encode(data)
}

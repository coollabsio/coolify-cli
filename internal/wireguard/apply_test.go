package wireguard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstLine(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"single line", "single line"},
		{"first\nsecond\nthird", "first"},
		{"  spaces  \nnext", "spaces  "},
		{"\nleading newline", "leading newline"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, firstLine(tt.input), "input: %q", tt.input)
	}
}

func TestNonEmptyLines(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"line1\nline2", []string{"line1", "line2"}},
		{"line1\n\nline2", []string{"line1", "line2"}},
		{"  \n  \nactual", []string{"actual"}},
		{"only", []string{"only"}},
	}
	for _, tt := range tests {
		got := nonEmptyLines(tt.input)
		assert.Equal(t, tt.want, got, "input: %q", tt.input)
	}
}

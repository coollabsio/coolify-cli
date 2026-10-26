package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type secretChild struct {
	Name     string  `json:"name"`
	Password *string `json:"password,omitempty" sensitive:"true"`
}

type secretParent struct {
	Name     string                 `json:"name"`
	Token    string                 `json:"token" sensitive:"true"`
	Empty    string                 `json:"empty" sensitive:"true"`
	Missing  *string                `json:"missing,omitempty" sensitive:"true"`
	Port     int                    `json:"port" sensitive:"true"`
	Child    secretChild            `json:"child"`
	ChildPtr *secretChild           `json:"child_ptr"`
	Children []secretChild          `json:"children"`
	ByName   map[string]secretChild `json:"by_name"`
	Extra    any                    `json:"extra"`
}

func strPtr(s string) *string { return &s }

func newSecretParent() secretParent {
	return secretParent{
		Name:     "parent",
		Token:    "top-secret",
		Port:     2222,
		Child:    secretChild{Name: "child", Password: strPtr("child-secret")},
		ChildPtr: &secretChild{Name: "child-ptr", Password: strPtr("ptr-secret")},
		Children: []secretChild{{Name: "a", Password: strPtr("slice-secret")}, {Name: "b"}},
		ByName:   map[string]secretChild{"m": {Name: "m", Password: strPtr("map-secret")}},
		Extra:    secretChild{Name: "iface", Password: strPtr("iface-secret")},
	}
}

func formatToMap(t *testing.T, format string, showSensitive bool, data any) any {
	t.Helper()
	buf := &bytes.Buffer{}
	formatter, err := NewFormatter(format, Options{Writer: buf, ShowSensitive: showSensitive})
	require.NoError(t, err)
	require.NoError(t, formatter.Format(data))
	var decoded any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	return decoded
}

func TestJSONFormats_MaskSensitiveFieldsByDefault(t *testing.T) {
	for _, format := range []string{FormatJSON, FormatPretty} {
		t.Run(format, func(t *testing.T) {
			parent := newSecretParent()
			got := formatToMap(t, format, false, &parent)

			assert.Equal(t, map[string]any{
				"name":      "parent",
				"token":     SensitiveOverlay,
				"empty":     SensitiveOverlay,
				"port":      float64(0),
				"child":     map[string]any{"name": "child", "password": SensitiveOverlay},
				"child_ptr": map[string]any{"name": "child-ptr", "password": SensitiveOverlay},
				"children": []any{
					map[string]any{"name": "a", "password": SensitiveOverlay},
					map[string]any{"name": "b"},
				},
				"by_name": map[string]any{"m": map[string]any{"name": "m", "password": SensitiveOverlay}},
				"extra":   map[string]any{"name": "iface", "password": SensitiveOverlay},
			}, got)

			// The caller's data must not change.
			assert.Equal(t, newSecretParent(), parent)
		})
	}
}

func TestJSONFormats_ShowSensitiveRevealsValues(t *testing.T) {
	for _, format := range []string{FormatJSON, FormatPretty} {
		t.Run(format, func(t *testing.T) {
			got := formatToMap(t, format, true, newSecretParent())

			assert.Equal(t, map[string]any{
				"name":      "parent",
				"token":     "top-secret",
				"empty":     "",
				"port":      float64(2222),
				"child":     map[string]any{"name": "child", "password": "child-secret"},
				"child_ptr": map[string]any{"name": "child-ptr", "password": "ptr-secret"},
				"children": []any{
					map[string]any{"name": "a", "password": "slice-secret"},
					map[string]any{"name": "b"},
				},
				"by_name": map[string]any{"m": map[string]any{"name": "m", "password": "map-secret"}},
				"extra":   map[string]any{"name": "iface", "password": "iface-secret"},
			}, got)
		})
	}
}

func TestJSONFormats_MaskTopLevelSlicesAndPointers(t *testing.T) {
	for _, format := range []string{FormatJSON, FormatPretty} {
		t.Run(format, func(t *testing.T) {
			items := []*secretChild{{Name: "x", Password: strPtr("x-secret")}, nil}
			got := formatToMap(t, format, false, items)
			assert.Equal(t, []any{map[string]any{"name": "x", "password": SensitiveOverlay}, nil}, got)
		})
	}
}

func TestJSONFormats_DataWithoutSensitiveFieldsIsUnchanged(t *testing.T) {
	data := map[string]any{"name": "n", "list": []any{"a", float64(1)}, "nested": map[string]any{"ok": true}}
	for _, format := range []string{FormatJSON, FormatPretty} {
		assert.Equal(t, data, formatToMap(t, format, false, data))
	}
	assert.Nil(t, formatToMap(t, FormatJSON, false, nil))
}

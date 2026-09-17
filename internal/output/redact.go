package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// redactSensitive returns a copy of data safe to serialize, with any field
// tagged `sensitive:"true"` replaced by SensitiveOverlay unless showSensitive
// is true. The original value is never mutated. Struct fields are walked
// recursively through pointers, slices, arrays, and maps, mirroring the
// traversal TableFormatter uses for its own per-field sensitive check.
func redactSensitive(data any, showSensitive bool) any {
	if showSensitive {
		return data
	}
	return redactValue(reflect.ValueOf(data))
}

// isSensitiveField reports whether a struct field is tagged `sensitive:"true"`.
func isSensitiveField(field reflect.StructField) bool {
	return field.Tag.Get("sensitive") == "true"
}

func redactValue(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}

	// Respect existing custom JSON encoding (e.g. time.Time) instead of
	// walking into its unexported internals.
	if v.CanInterface() {
		if _, ok := v.Interface().(json.Marshaler); ok {
			return v.Interface()
		}
	}

	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return redactValue(v.Elem())
	case reflect.Struct:
		return redactStruct(v)
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			return nil
		}
		items := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			items[i] = redactValue(v.Index(i))
		}
		return items
	case reflect.Map:
		if v.IsNil() {
			return nil
		}
		m := newOrderedMap(v.Len())
		iter := v.MapRange()
		for iter.Next() {
			m.set(fmt.Sprint(iter.Key().Interface()), redactValue(iter.Value()))
		}
		return m
	default:
		return v.Interface()
	}
}

// redactStruct converts a struct into an orderedMap so field order in the
// rendered JSON matches the struct's declared field order, exactly like
// marshaling the struct directly would.
func redactStruct(v reflect.Value) *orderedMap {
	t := v.Type()
	result := newOrderedMap(t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name, omitempty, skip := jsonFieldName(field)
		if skip {
			continue
		}

		fv := v.Field(i)

		if isSensitiveField(field) {
			if omitempty && isEmptyValue(fv) {
				continue
			}
			result.set(name, SensitiveOverlay)
			continue
		}

		if omitempty && isEmptyValue(fv) {
			continue
		}
		result.set(name, redactValue(fv))
	}

	return result
}

// jsonFieldName mirrors encoding/json's handling of the `json` struct tag.
func jsonFieldName(field reflect.StructField) (name string, omitempty, skip bool) {
	name = field.Name

	tag := field.Tag.Get("json")
	if tag == "" {
		return name, false, false
	}

	parts := strings.Split(tag, ",")
	if parts[0] == "-" {
		return "", false, true
	}
	if parts[0] != "" {
		name = parts[0]
	}
	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			omitempty = true
		}
	}

	return name, omitempty, false
}

// isEmptyValue mirrors encoding/json's definition of "empty" for omitempty.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Array, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	default:
		return false
	}
}

// orderedMap marshals to a JSON object preserving insertion order, since
// map[string]any would otherwise sort keys alphabetically.
type orderedMap struct {
	keys   []string
	values []any
}

func newOrderedMap(capacity int) *orderedMap {
	return &orderedMap{
		keys:   make([]string, 0, capacity),
		values: make([]any, 0, capacity),
	}
}

func (m *orderedMap) set(key string, value any) {
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
}

func (m *orderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range m.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal key %q: %w", key, err)
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')
		valueBytes, err := json.Marshal(m.values[i])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value for key %q: %w", key, err)
		}
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

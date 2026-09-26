package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// redactSensitive returns a copy of data safe to serialize, with any field
// tagged `sensitive:"true"` replaced by SensitiveOverlay unless showSensitive
// is true. The original value is never mutated.
//
// Any value whose type cannot contain a sensitive field is handed back
// untouched so encoding/json marshals it natively. That keeps non-sensitive
// output byte-for-byte identical to marshaling the original value directly,
// including behavior this package does not reimplement ([]byte as base64,
// json.Marshaler, encoding.TextMarshaler, sorted map keys).
//
// Note: SensitiveOverlay is a string, so a redacted field of a non-string type
// (e.g. models.Server.Port) is rendered as a JSON string rather than its
// original type. Pass --show-sensitive to get the real, correctly typed value.
func redactSensitive(data any, showSensitive bool) any {
	if showSensitive {
		return data
	}
	return redactValue(reflect.ValueOf(data), nil)
}

// isSensitiveField reports whether a struct field is tagged `sensitive:"true"`.
func isSensitiveField(field reflect.StructField) bool {
	return field.Tag.Get("sensitive") == "true"
}

// redactValue walks v, rebuilding only the parts that can contain a sensitive
// field. visited holds the pointers on the current path so a self-referential
// value degrades to null instead of overflowing the stack.
func redactValue(v reflect.Value, visited map[uintptr]bool) any {
	if !v.IsValid() || !v.CanInterface() {
		return nil
	}

	if !containsSensitive(v.Type()) {
		return v.Interface()
	}

	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return nil
		}
		address := v.Pointer()
		if visited[address] {
			return nil
		}
		if visited == nil {
			visited = make(map[uintptr]bool)
		}
		visited[address] = true
		defer delete(visited, address)
		return redactValue(v.Elem(), visited)
	case reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return redactValue(v.Elem(), visited)
	case reflect.Struct:
		return redactStruct(v, visited)
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			return nil
		}
		items := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			items[i] = redactValue(v.Index(i), visited)
		}
		return items
	case reflect.Map:
		if v.IsNil() {
			return nil
		}
		// A plain map keeps encoding/json's own key sorting, which a
		// insertion-ordered map would replace with Go's randomized map order.
		m := make(map[string]any, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			m[fmt.Sprint(iter.Key().Interface())] = redactValue(iter.Value(), visited)
		}
		return m
	default:
		return v.Interface()
	}
}

// redactStruct converts a struct into an orderedMap so field order in the
// rendered JSON matches the struct's declared field order, exactly like
// marshaling the struct directly would.
func redactStruct(v reflect.Value, visited map[uintptr]bool) *orderedMap {
	result := newOrderedMap(v.Type().NumField())
	appendStructFields(v, result, visited)
	return result
}

func appendStructFields(v reflect.Value, result *orderedMap, visited map[uintptr]bool) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")

		// encoding/json promotes an embedded struct's fields into the parent
		// object unless the embedded field carries an explicit JSON name.
		if field.Anonymous && embeddedName(tag) == "" {
			fv := v.Field(i)
			if fv.Kind() == reflect.Pointer {
				if fv.IsNil() {
					continue
				}
				fv = fv.Elem()
			}
			if fv.Kind() == reflect.Struct {
				appendStructFields(fv, result, visited)
				continue
			}
		}

		if !field.IsExported() {
			continue
		}

		name, omitempty, skip := jsonFieldName(field)
		if skip {
			continue
		}

		fv := v.Field(i)
		if omitempty && isEmptyValue(fv) {
			continue
		}

		if isSensitiveField(field) {
			result.set(name, SensitiveOverlay)
			continue
		}

		result.set(name, redactValue(fv, visited))
	}
}

// embeddedName returns the explicit JSON name on an embedded field's tag, which
// stops encoding/json from promoting that field's contents into the parent.
func embeddedName(tag string) string {
	if tag == "" {
		return ""
	}
	name := strings.Split(tag, ",")[0]
	if name == "-" {
		// `json:"-"` drops the field; treat it as named so it is not promoted.
		return "-"
	}
	return name
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

var sensitiveTypeCache sync.Map // reflect.Type -> bool

// containsSensitive reports whether t, or anything reachable from it, has a
// field tagged `sensitive:"true"`. An interface type counts as sensitive
// because its dynamic value is only knowable at runtime.
func containsSensitive(t reflect.Type) bool {
	if cached, ok := sensitiveTypeCache.Load(t); ok {
		return cached.(bool)
	}
	result := typeContainsSensitive(t, map[reflect.Type]bool{})
	sensitiveTypeCache.Store(t, result)
	return result
}

func typeContainsSensitive(t reflect.Type, inProgress map[reflect.Type]bool) bool {
	if cached, ok := sensitiveTypeCache.Load(t); ok {
		return cached.(bool)
	}
	if inProgress[t] {
		// Recursive type: this edge adds nothing on its own.
		return false
	}
	inProgress[t] = true
	defer delete(inProgress, t)

	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
		return typeContainsSensitive(t.Elem(), inProgress)
	case reflect.Interface:
		return true
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if strings.Split(field.Tag.Get("json"), ",")[0] == "-" {
				continue
			}
			if !field.IsExported() && !field.Anonymous {
				continue
			}
			if isSensitiveField(field) {
				return true
			}
			if typeContainsSensitive(field.Type, inProgress) {
				return true
			}
		}
		return false
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

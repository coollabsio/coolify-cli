package output

import "reflect"

// visibleData returns data as JSON output may show it: unchanged when
// ShowSensitive is set, otherwise a copy with every field tagged
// sensitive:"true" masked. The caller's data is never modified.
func visibleData(data any, opts Options) any {
	if opts.ShowSensitive || data == nil {
		return data
	}
	return redact(reflect.ValueOf(data)).Interface()
}

// redact returns a deep copy of v with sensitive struct fields masked.
func redact(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return v
		}
		out := reflect.New(v.Type().Elem())
		out.Elem().Set(redact(v.Elem()))
		return out
	case reflect.Interface:
		if v.IsNil() {
			return v
		}
		out := reflect.New(v.Type()).Elem()
		out.Set(redact(v.Elem()))
		return out
	case reflect.Struct:
		out := reflect.New(v.Type()).Elem()
		out.Set(v)
		for i := range v.NumField() {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			if field.Tag.Get("sensitive") == "true" {
				mask(out.Field(i))
			} else {
				out.Field(i).Set(redact(v.Field(i)))
			}
		}
		return out
	case reflect.Slice:
		if v.IsNil() {
			return v
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := range v.Len() {
			out.Index(i).Set(redact(v.Index(i)))
		}
		return out
	case reflect.Array:
		out := reflect.New(v.Type()).Elem()
		for i := range v.Len() {
			out.Index(i).Set(redact(v.Index(i)))
		}
		return out
	case reflect.Map:
		if v.IsNil() {
			return v
		}
		out := reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()
		for iter.Next() {
			out.SetMapIndex(iter.Key(), redact(iter.Value()))
		}
		return out
	default:
		return v
	}
}

// mask hides a sensitive field the same way the table format does: a present
// string becomes SensitiveOverlay and a nil pointer stays nil. JSON cannot hold
// the overlay in other types, so they become their zero value.
func mask(field reflect.Value) {
	switch {
	case field.Kind() == reflect.String:
		field.SetString(SensitiveOverlay)
	case field.Kind() == reflect.Pointer && field.Type().Elem().Kind() == reflect.String:
		if field.IsNil() {
			return
		}
		overlay := reflect.New(field.Type().Elem())
		overlay.Elem().SetString(SensitiveOverlay)
		field.Set(overlay)
	default:
		field.SetZero()
	}
}

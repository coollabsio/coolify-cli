package output

import (
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"
)

// TableFormatter formats output as a table
type TableFormatter struct {
	opts Options
}

// NewTableFormatter creates a new table formatter
func NewTableFormatter(opts Options) *TableFormatter {
	return &TableFormatter{opts: opts}
}

// Format formats the data as a table
func (f *TableFormatter) Format(data interface{}) error {
	w := tabwriter.NewWriter(f.opts.Writer, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Handle different data types
	val := reflect.ValueOf(data)

	// Dereference pointer if needed
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		return f.formatSlice(w, val)
	case reflect.Struct:
		return f.formatStruct(w, val)
	case reflect.Map:
		return f.formatMap(w, val)
	default:
		return fmt.Errorf("unsupported data type for table format: %v", val.Kind())
	}
}

// formatSlice formats a slice of structs as a table
func (f *TableFormatter) formatSlice(w *tabwriter.Writer, val reflect.Value) error {
	if val.Len() == 0 {
		fmt.Fprintln(w, "No data")
		return nil
	}

	// Get the first element to determine columns
	firstElem := val.Index(0)
	if firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}

	if firstElem.Kind() != reflect.Struct {
		// Simple slice (e.g., []string)
		for i := 0; i < val.Len(); i++ {
			fmt.Fprintf(w, "%v\n", val.Index(i).Interface())
		}
		return nil
	}

	// Get column headers from struct tags or field names
	headers := f.getHeaders(firstElem.Type())
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Print rows
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		row := f.formatStructRow(elem)
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return nil
}

// formatStruct formats a single struct as a table (horizontal layout with headers)
func (f *TableFormatter) formatStruct(w *tabwriter.Writer, val reflect.Value) error {
	// Get headers
	headers := f.getHeaders(val.Type())
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Get row data
	row := f.formatStructRow(val)
	fmt.Fprintln(w, strings.Join(row, "\t"))

	return nil
}

// formatMap formats a map as a table
func (f *TableFormatter) formatMap(w *tabwriter.Writer, val reflect.Value) error {
	fmt.Fprintln(w, "Key\tValue")

	iter := val.MapRange()
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		fmt.Fprintf(w, "%v\t%v\n", key.Interface(), f.formatValue(value))
	}

	return nil
}

// getHeaders extracts column headers from struct type
func (f *TableFormatter) getHeaders(typ reflect.Type) []string {
	var headers []string

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Check table tag for skip
		if tableTag := field.Tag.Get("table"); tableTag == "-" {
			continue
		}

		fieldName := field.Name
		// Use json tag if available
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			fieldName = strings.Split(jsonTag, ",")[0]
			if fieldName == "-" || fieldName == "omitempty" {
				continue
			}
		}

		headers = append(headers, fieldName)
	}

	return headers
}

// formatStructRow extracts values from a struct as a row
func (f *TableFormatter) formatStructRow(val reflect.Value) []string {
	var row []string

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Check table tag for skip
		if tableTag := field.Tag.Get("table"); tableTag == "-" {
			continue
		}

		// Check json tag for skip
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			fieldName := strings.Split(jsonTag, ",")[0]
			if fieldName == "-" {
				continue
			}
		}

		value := val.Field(i)

		// Check if field is marked as sensitive
		isSensitive := field.Tag.Get("sensitive") == "true"
		if isSensitive && !f.opts.ShowSensitive {
			row = append(row, SensitiveOverlay)
		} else {
			row = append(row, f.formatValue(value))
		}
	}

	return row
}

// formatValue formats a reflect.Value for display
func (f *TableFormatter) formatValue(val reflect.Value) string {
	// Handle nil pointers
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return ""
	}

	// Dereference pointer
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Handle different types
	switch val.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Bool:
		if val.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", val.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", val.Float())
	case reflect.Slice, reflect.Array:
		if val.Len() == 0 {
			return "[]"
		}
		// Check if it's a slice of structs
		elemType := val.Index(0).Kind()
		if elemType == reflect.Struct || elemType == reflect.Ptr {
			// For complex types, try to extract Name field from all elements
			var names []string
			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i)
				if elem.Kind() == reflect.Ptr && !elem.IsNil() {
					elem = elem.Elem()
				}
				if elem.Kind() == reflect.Struct {
					nameField := elem.FieldByName("Name")
					if nameField.IsValid() && nameField.Kind() == reflect.String {
						names = append(names, nameField.String())
					}
				}
			}
			if len(names) > 0 {
				return strings.Join(names, ", ")
			}
			return fmt.Sprintf("(%d items)", val.Len())
		}
		// For simple types, show comma-separated values
		var items []string
		for i := 0; i < val.Len(); i++ {
			items = append(items, f.formatValue(val.Index(i)))
		}
		return strings.Join(items, ", ")
	case reflect.Struct:
		// For nested structs, try to show a name field if available
		nameField := val.FieldByName("Name")
		if nameField.IsValid() && nameField.Kind() == reflect.String {
			return nameField.String()
		}
		return fmt.Sprintf("(%s)", val.Type().Name())
	default:
		return fmt.Sprintf("%v", val.Interface())
	}
}

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

func (f *TableFormatter) Format(data any) (err error) {
	w := tabwriter.NewWriter(f.opts.Writer, 0, 0, 2, ' ', tabwriter.Debug)
	defer func() {
		if flushErr := w.Flush(); flushErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to flush table writer: %w", flushErr)
			}
		}
		// Add a final newline nach table output, but only if no error occurred
		if err == nil {
			if _, nlErr := fmt.Fprintln(f.opts.Writer); nlErr != nil {
				err = fmt.Errorf("failed to write trailing newline: %w", nlErr)
			}
		}
	}()

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
		if _, err := fmt.Fprintln(w, "No data"); err != nil {
			return fmt.Errorf("failed to write no data message: %w", err)
		}
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
			if _, err := fmt.Fprintf(w, "%v\n", val.Index(i).Interface()); err != nil {
				return fmt.Errorf("failed to write slice element: %w", err)
			}
		}
		return nil
	}

	// Get column headers from struct tags or field names
	headers := f.getHeaders(firstElem.Type())
	// Add # as first column header
	headersWithNum := append([]string{"#"}, headers...)
	if _, err := fmt.Fprintln(w, strings.Join(headersWithNum, "\t")); err != nil {
		return fmt.Errorf("failed to write table headers: %w", err)
	}

	// Print rows
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		row := f.formatStructRow(elem)
		// Add row number (1-indexed) as first column
		rowWithNum := append([]string{fmt.Sprintf("%d", i+1)}, row...)
		if _, err := fmt.Fprintln(w, strings.Join(rowWithNum, "\t")); err != nil {
			return fmt.Errorf("failed to write table row: %w", err)
		}
	}

	return nil
}

// formatStruct formats a single struct as a table (horizontal layout with headers)
func (f *TableFormatter) formatStruct(w *tabwriter.Writer, val reflect.Value) error {
	// Get headers
	headers := f.getHeaders(val.Type())
	if _, err := fmt.Fprintln(w, strings.Join(headers, "\t")); err != nil {
		return fmt.Errorf("failed to write struct headers: %w", err)
	}

	// Get row data
	row := f.formatStructRow(val)
	if _, err := fmt.Fprintln(w, strings.Join(row, "\t")); err != nil {
		return fmt.Errorf("failed to write struct row: %w", err)
	}

	return nil
}

// formatMap formats a map as a table
func (f *TableFormatter) formatMap(w *tabwriter.Writer, val reflect.Value) error {
	if _, err := fmt.Fprintln(w, "Key\tValue"); err != nil {
		return fmt.Errorf("failed to write map headers: %w", err)
	}

	iter := val.MapRange()
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		if _, err := fmt.Fprintf(w, "%v\t%v\n", key.Interface(), f.formatValue(value)); err != nil {
			return fmt.Errorf("failed to write map entry: %w", err)
		}
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

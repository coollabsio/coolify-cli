package output

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
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
	// Handle different data types
	val := reflect.ValueOf(data)

	// Dereference pointer if needed
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check for empty slice/array before creating the table writer
	if (val.Kind() == reflect.Slice || val.Kind() == reflect.Array) && val.Len() == 0 {
		if _, writeErr := fmt.Fprintln(f.opts.Writer, "No data"); writeErr != nil {
			return fmt.Errorf("failed to write no data message: %w", writeErr)
		}
		return nil
	}

	w := tablewriter.NewWriter(f.opts.Writer)
	defer func() {
		if renderErr := w.Render(); renderErr != nil && err == nil {
			err = fmt.Errorf("failed to render table: %w", renderErr)
		}
	}()

	// disable ALL CAPS for column headers
	w.Options(tablewriter.WithHeaderAutoFormat(tw.Off))

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
func (f *TableFormatter) formatSlice(w *tablewriter.Table, val reflect.Value) error {
	// Get the first element to determine columns
	firstElem := val.Index(0)
	if firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}

	if firstElem.Kind() != reflect.Struct {
		// Simple slice (e.g., []string)
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			row := []string{f.formatValue(elem)}
			if err := w.Append(row); err != nil {
				return fmt.Errorf("failed to write slice element: %w", err)
			}
		}
		return nil
	}

	// Get column headers from struct tags or field names
	headers := f.getHeaders(firstElem.Type())
	// Add # as first column header
	headersWithNum := append([]string{"#"}, headers...)
	w.Header(headersWithNum)

	// Print rows
	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		row := f.formatStructRow(elem)
		// Add row number (1-indexed) as first column
		rowWithNum := append([]string{fmt.Sprintf("%d", i+1)}, row...)
		if err := w.Append(rowWithNum); err != nil {
			return fmt.Errorf("failed to write table row: %w", err)
		}
	}

	return nil
}

// formatStruct formats a single struct as a table (horizontal layout with headers)
func (f *TableFormatter) formatStruct(w *tablewriter.Table, val reflect.Value) error {
	// Get headers
	headers := f.getHeaders(val.Type())
	w.Header(headers)

	// Get row data
	row := f.formatStructRow(val)
	if err := w.Append(row); err != nil {
		return fmt.Errorf("failed to write struct row: %w", err)
	}

	return nil
}

// formatMap formats a map as a table
func (f *TableFormatter) formatMap(w *tablewriter.Table, val reflect.Value) error {
	w.Header([]string{"Key", "Value"})

	iter := val.MapRange()
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		row := []string{f.formatValue(key), f.formatValue(value)}
		if err := w.Append(row); err != nil {
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

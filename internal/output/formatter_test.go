package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestServer struct {
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{"table format", FormatTable, false},
		{"json format", FormatJSON, false},
		{"pretty format", FormatPretty, false},
		{"invalid format", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{Writer: &bytes.Buffer{}}
			formatter, err := NewFormatter(tt.format, opts)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, formatter)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, formatter)
			}
		})
	}
}

func TestJSONFormatter(t *testing.T) {
	servers := []TestServer{
		{UUID: "uuid-1", Name: "server-1", Status: "running"},
		{UUID: "uuid-2", Name: "server-2", Status: "stopped"},
	}

	buf := &bytes.Buffer{}
	formatter := NewJSONFormatter(Options{Writer: buf})

	err := formatter.Format(servers)
	require.NoError(t, err)

	// Verify valid JSON
	var result []TestServer
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "uuid-1", result[0].UUID)
	assert.Equal(t, "server-1", result[0].Name)
}

func TestPrettyFormatter(t *testing.T) {
	servers := []TestServer{
		{UUID: "uuid-1", Name: "server-1", Status: "running"},
	}

	buf := &bytes.Buffer{}
	formatter := NewPrettyFormatter(Options{Writer: buf})

	err := formatter.Format(servers)
	require.NoError(t, err)

	output := buf.String()

	// Verify it's indented JSON
	assert.Contains(t, output, "  ")
	assert.Contains(t, output, "uuid-1")
	assert.Contains(t, output, "server-1")

	// Verify valid JSON
	var result []TestServer
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
}

func TestTableFormatter_Slice(t *testing.T) {
	servers := []TestServer{
		{UUID: "uuid-1", Name: "server-1", Status: "running"},
		{UUID: "uuid-2", Name: "server-2", Status: "stopped"},
	}

	buf := &bytes.Buffer{}
	formatter := NewTableFormatter(Options{Writer: buf})

	err := formatter.Format(servers)
	require.NoError(t, err)

	output := buf.String()

	// Check headers
	assert.Contains(t, output, "uuid")
	assert.Contains(t, output, "name")
	assert.Contains(t, output, "status")

	// Check data
	assert.Contains(t, output, "uuid-1")
	assert.Contains(t, output, "server-1")
	assert.Contains(t, output, "running")
	assert.Contains(t, output, "uuid-2")
	assert.Contains(t, output, "server-2")
	assert.Contains(t, output, "stopped")
}

func TestTableFormatter_SingleStruct(t *testing.T) {
	server := TestServer{
		UUID:   "uuid-1",
		Name:   "server-1",
		Status: "running",
	}

	buf := &bytes.Buffer{}
	formatter := NewTableFormatter(Options{Writer: buf})

	err := formatter.Format(server)
	require.NoError(t, err)

	output := buf.String()

	// Check field names and values
	assert.Contains(t, output, "uuid")
	assert.Contains(t, output, "uuid-1")
	assert.Contains(t, output, "name")
	assert.Contains(t, output, "server-1")
	assert.Contains(t, output, "status")
	assert.Contains(t, output, "running")
}

func TestTableFormatter_EmptySlice(t *testing.T) {
	var servers []TestServer

	buf := &bytes.Buffer{}
	formatter := NewTableFormatter(Options{Writer: buf})

	err := formatter.Format(servers)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No data")
}

func TestTableFormatter_Map(t *testing.T) {
	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	buf := &bytes.Buffer{}
	formatter := NewTableFormatter(Options{Writer: buf})

	err := formatter.Format(data)
	require.NoError(t, err)

	output := buf.String()

	// Check headers
	assert.Contains(t, output, "Key")
	assert.Contains(t, output, "Value")
}

func TestTableFormatter_SimpleSlice(t *testing.T) {
	data := []string{"item1", "item2", "item3"}

	buf := &bytes.Buffer{}
	formatter := NewTableFormatter(Options{Writer: buf})

	err := formatter.Format(data)
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, "item1")
	assert.Contains(t, output, "item2")
	assert.Contains(t, output, "item3")
}

func TestTableFormatter_BooleanValues(t *testing.T) {
	type TestStruct struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
		Active  bool   `json:"active"`
	}

	data := []TestStruct{
		{Name: "test1", Enabled: true, Active: false},
		{Name: "test2", Enabled: false, Active: true},
	}

	buf := &bytes.Buffer{}
	formatter := NewTableFormatter(Options{Writer: buf})

	err := formatter.Format(data)
	require.NoError(t, err)

	output := buf.String()

	// Check boolean formatting
	lines := strings.Split(output, "\n")
	assert.Contains(t, lines[1], "true")
	assert.Contains(t, lines[1], "false")
	assert.Contains(t, lines[2], "false")
	assert.Contains(t, lines[2], "true")
}

func TestTableFormatter_NilPointer(t *testing.T) {
	type TestStruct struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}

	desc := "test description"
	data := []TestStruct{
		{Name: "test1", Description: &desc},
		{Name: "test2", Description: nil},
	}

	buf := &bytes.Buffer{}
	formatter := NewTableFormatter(Options{Writer: buf})

	err := formatter.Format(data)
	require.NoError(t, err)

	output := buf.String()

	// First row should have description
	assert.Contains(t, output, "test description")
	// Second row should handle nil gracefully
	assert.Contains(t, output, "test2")
}

func TestTableFormatter_SliceField(t *testing.T) {
	type TestStruct struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	data := []TestStruct{
		{Name: "test1", Tags: []string{"tag1", "tag2", "tag3"}},
	}

	buf := &bytes.Buffer{}
	formatter := NewTableFormatter(Options{Writer: buf})

	err := formatter.Format(data)
	require.NoError(t, err)

	output := buf.String()

	// Tags should be comma-separated
	assert.Contains(t, output, "tag1, tag2, tag3")
}

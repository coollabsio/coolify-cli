package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAndFormatLogs(t *testing.T) {
	tests := []struct {
		name           string
		logsJSON       string
		showHidden     bool
		expectedOutput string
		expectLines    int
	}{
		{
			name: "parse deployment logs without hidden",
			logsJSON: `[
				{
					"command": null,
					"output": "Starting deployment",
					"type": "stdout",
					"timestamp": "2025-11-10T11:49:38.449766Z",
					"hidden": false,
					"batch": 1,
					"order": 1
				},
				{
					"command": "docker stop --time=1 xyz",
					"output": "Flag --time has been deprecated",
					"type": "stdout",
					"timestamp": "2025-11-10T11:49:38.687014Z",
					"hidden": true,
					"batch": 1,
					"order": 2
				},
				{
					"command": null,
					"output": "Deployment successful",
					"type": "stdout",
					"timestamp": "2025-11-10T11:49:40.449766Z",
					"hidden": false,
					"batch": 2,
					"order": 3
				}
			]`,
			showHidden:  false,
			expectLines: 2,
		},
		{
			name: "parse deployment logs with hidden",
			logsJSON: `[
				{
					"command": null,
					"output": "Starting deployment",
					"type": "stdout",
					"timestamp": "2025-11-10T11:49:38.449766Z",
					"hidden": false,
					"batch": 1,
					"order": 1
				},
				{
					"command": "docker stop --time=1 xyz",
					"output": "Flag --time has been deprecated",
					"type": "stdout",
					"timestamp": "2025-11-10T11:49:38.687014Z",
					"hidden": true,
					"batch": 1,
					"order": 2
				},
				{
					"command": null,
					"output": "Deployment successful",
					"type": "stdout",
					"timestamp": "2025-11-10T11:49:40.449766Z",
					"hidden": false,
					"batch": 2,
					"order": 3
				}
			]`,
			showHidden:  true,
			expectLines: 3,
		},
		{
			name: "parse with errors",
			logsJSON: `[
				{
					"command": null,
					"output": "Starting deployment",
					"type": "stdout",
					"timestamp": "2025-11-10T11:49:38.449766Z",
					"hidden": false,
					"batch": 1,
					"order": 1
				},
				{
					"command": "docker build",
					"output": "Error: No such container",
					"type": "stderr",
					"timestamp": "2025-11-10T11:49:38.687014Z",
					"hidden": false,
					"batch": 1,
					"order": 2
				}
			]`,
			showHidden:  false,
			expectLines: 2,
		},
		{
			name:           "invalid json returns original",
			logsJSON:       "This is not JSON",
			showHidden:     false,
			expectedOutput: "This is not JSON",
			expectLines:    0,
		},
		{
			name: "filter empty outputs",
			logsJSON: `[
				{
					"command": null,
					"output": "Starting deployment",
					"type": "stdout",
					"timestamp": "2025-11-10T11:49:38.449766Z",
					"hidden": false,
					"batch": 1,
					"order": 1
				},
				{
					"command": null,
					"output": "",
					"type": "stdout",
					"timestamp": "2025-11-10T11:49:38.687014Z",
					"hidden": false,
					"batch": 1,
					"order": 2
				},
				{
					"command": null,
					"output": "Done",
					"type": "stdout",
					"timestamp": "2025-11-10T11:49:40.449766Z",
					"hidden": false,
					"batch": 2,
					"order": 3
				}
			]`,
			showHidden:  false,
			expectLines: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAndFormatLogs(tt.logsJSON, tt.showHidden)
			require.NoError(t, err)

			if tt.expectedOutput != "" {
				assert.Equal(t, tt.expectedOutput, result)
			} else if tt.expectLines > 0 {
				// Count non-empty lines
				lines := strings.Split(strings.TrimSpace(result), "\n")
				nonEmptyLines := 0
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						nonEmptyLines++
					}
				}
				assert.Equal(t, tt.expectLines, nonEmptyLines)
			}
		})
	}
}

func TestParseAndFormatLogs_Ordering(t *testing.T) {
	logsJSON := `[
		{
			"output": "Third",
			"type": "stdout",
			"hidden": false,
			"batch": 2,
			"order": 3
		},
		{
			"output": "First",
			"type": "stdout",
			"hidden": false,
			"batch": 1,
			"order": 1
		},
		{
			"output": "Second",
			"type": "stdout",
			"hidden": false,
			"batch": 1,
			"order": 2
		}
	]`

	result, err := ParseAndFormatLogs(logsJSON, false)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	assert.Equal(t, "First", lines[0])
	assert.Equal(t, "Second", lines[1])
	assert.Equal(t, "Third", lines[2])
}

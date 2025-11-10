package models

import (
	"encoding/json"
	"sort"
	"strings"
)

// LogEntry represents a single log entry from a deployment
type LogEntry struct {
	Command   *string `json:"command"`
	Output    string  `json:"output"`
	Type      string  `json:"type"`
	Timestamp string  `json:"timestamp"`
	Hidden    bool    `json:"hidden"`
	Batch     int     `json:"batch"`
	Order     int     `json:"order,omitempty"`
}

// ParseAndFormatLogs parses the JSON logs string and formats it as human-readable text
func ParseAndFormatLogs(logsJSON string, showHidden bool) (string, error) {
	// Try to parse as JSON array
	var logs []LogEntry
	// Ignore parse errors - if it's not JSON, just return original
	_ = json.Unmarshal([]byte(logsJSON), &logs)
	if len(logs) == 0 {
		// If parsing failed or array is empty, return original string
		return logsJSON, nil
	}

	// Sort logs by batch and order to ensure correct sequence
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].Batch != logs[j].Batch {
			return logs[i].Batch < logs[j].Batch
		}
		return logs[i].Order < logs[j].Order
	})

	var output strings.Builder
	for _, log := range logs {
		// Skip hidden logs unless requested
		if log.Hidden && !showHidden {
			continue
		}

		// Format the log entry
		line := formatLogEntry(log)
		if line != "" {
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	return output.String(), nil
}

// formatLogEntry formats a single log entry as a human-readable line
func formatLogEntry(log LogEntry) string {
	// Just return the output text, one line per log entry
	return log.Output
}

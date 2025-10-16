package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// EnvVar represents a parsed environment variable
type EnvVar struct {
	Key   string
	Value string
}

// ParseEnvFile parses a .env file and returns a slice of environment variables
// Supports:
// - KEY=value
// - KEY="value"
// - KEY='value'
// - Multiline values with quotes
// - Comments (lines starting with #)
// - Empty lines
func ParseEnvFile(filepath string) ([]EnvVar, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var envVars []EnvVar
	scanner := bufio.NewScanner(file)
	lineNum := 0
	var currentVar *EnvVar
	var inMultiline bool
	var quoteChar rune

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Handle multiline continuation
		if inMultiline {
			currentVar.Value += "\n" + line
			// Check if this line closes the multiline value
			if strings.HasSuffix(line, string(quoteChar)) {
				// Remove the closing quote
				currentVar.Value = strings.TrimSuffix(currentVar.Value, string(quoteChar))
				envVars = append(envVars, *currentVar)
				currentVar = nil
				inMultiline = false
				quoteChar = 0
			}
			continue
		}

		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Find the first = sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format at line %d: missing '='", lineNum)
		}

		key := strings.TrimSpace(parts[0])
		value := parts[1]

		if key == "" {
			return nil, fmt.Errorf("invalid format at line %d: empty key", lineNum)
		}

		// Handle quoted values
		if len(value) >= 2 {
			firstChar := rune(value[0])
			if firstChar == '"' || firstChar == '\'' {
				// Check if the closing quote is on the same line
				if strings.HasSuffix(value, string(firstChar)) && len(value) > 1 {
					// Single-line quoted value
					value = value[1 : len(value)-1]
					envVars = append(envVars, EnvVar{Key: key, Value: value})
				} else {
					// Start of multiline quoted value
					currentVar = &EnvVar{Key: key, Value: value[1:]} // Remove opening quote
					inMultiline = true
					quoteChar = firstChar
				}
				continue
			}
		}

		// Unquoted value
		envVars = append(envVars, EnvVar{Key: key, Value: value})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if inMultiline {
		return nil, fmt.Errorf("unclosed quoted value for key '%s'", currentVar.Key)
	}

	return envVars, nil
}

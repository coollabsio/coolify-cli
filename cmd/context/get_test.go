package context

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCommand_NeverExposesTokens(t *testing.T) {
	for _, format := range []string{"table", "json", "pretty"} {
		for _, showSensitive := range []bool{false, true} {
			name := format
			args := []string{"get", "production", "--format", format}
			if showSensitive {
				name += " with show-sensitive"
				args = append(args, "--show-sensitive")
			}
			t.Run(name, func(t *testing.T) {
				output := executeContextCommand(t, NewGetCommand(), args...)
				assert.NotContains(t, output, sentinelToken)
				assert.NotContains(t, strings.ToLower(output), "token")
			})
		}
	}
}

func TestGetCommand_StructuredFormatsContainOnlySafeContextMetadata(t *testing.T) {
	for _, format := range []string{"json", "pretty"} {
		t.Run(format, func(t *testing.T) {
			output := executeContextCommand(t, NewGetCommand(), "get", "staging", "--format", format)

			var contexts []map[string]any
			require.NoError(t, json.Unmarshal([]byte(output), &contexts))
			assert.Equal(t, []map[string]any{{
				"name":    "staging",
				"fqdn":    "https://staging.example.com",
				"default": false,
			}}, contexts)
		})
	}
}

package context

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sentinelToken = "sentinel-secret-token"

func executeListCommand(t *testing.T, format string, showSensitive bool) string {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("instances", []any{map[string]any{
		"name":    "production",
		"fqdn":    "https://coolify.example.com",
		"token":   sentinelToken,
		"default": true,
	}, map[string]any{
		"name":    "staging",
		"fqdn":    "https://staging.example.com",
		"token":   sentinelToken,
		"default": false,
	}})

	root := &cobra.Command{Use: "test"}
	root.PersistentFlags().String("format", "table", "")
	root.PersistentFlags().Bool("show-sensitive", false, "")
	root.AddCommand(NewListCommand())
	root.SetArgs([]string{"list", "--format", format})
	if showSensitive {
		root.SetArgs([]string{"list", "--format", format, "--show-sensitive"})
	}

	read, write, err := os.Pipe()
	require.NoError(t, err)
	originalStdout := os.Stdout
	os.Stdout = write
	t.Cleanup(func() { os.Stdout = originalStdout })

	var output bytes.Buffer
	require.NoError(t, root.Execute())
	require.NoError(t, write.Close())
	os.Stdout = originalStdout
	_, err = io.Copy(&output, read)
	require.NoError(t, err)
	require.NoError(t, read.Close())
	return output.String()
}

func TestListCommand_NeverExposesTokens(t *testing.T) {
	for _, format := range []string{"table", "json", "pretty"} {
		for _, showSensitive := range []bool{false, true} {
			name := format
			if showSensitive {
				name += " with show-sensitive"
			}
			t.Run(name, func(t *testing.T) {
				output := executeListCommand(t, format, showSensitive)
				assert.NotContains(t, output, sentinelToken)
				assert.NotContains(t, strings.ToLower(output), "token")
			})
		}
	}
}

func TestListCommand_StructuredFormatsContainOnlySafeContextMetadata(t *testing.T) {
	for _, format := range []string{"json", "pretty"} {
		t.Run(format, func(t *testing.T) {
			output := executeListCommand(t, format, false)

			var contexts []map[string]any
			require.NoError(t, json.Unmarshal([]byte(output), &contexts))
			require.Len(t, contexts, 2)
			assert.Equal(t, map[string]any{
				"name":    "production",
				"fqdn":    "https://coolify.example.com",
				"default": true,
			}, contexts[0])
			assert.Equal(t, map[string]any{
				"name":    "staging",
				"fqdn":    "https://staging.example.com",
				"default": false,
			}, contexts[1])
		})
	}
}

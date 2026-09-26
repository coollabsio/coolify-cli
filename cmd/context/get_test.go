package context

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const getSentinelToken = "sentinel-secret-token"

func executeGetCommand(t *testing.T, format string, showSensitive bool) string {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("instances", []any{map[string]any{
		"name":    "production",
		"fqdn":    "https://coolify.example.com",
		"token":   getSentinelToken,
		"default": true,
	}})

	root := &cobra.Command{Use: "test"}
	root.PersistentFlags().String("format", "table", "")
	root.PersistentFlags().Bool("show-sensitive", false, "")
	root.AddCommand(NewGetCommand())
	args := []string{"get", "production", "--format", format}
	if showSensitive {
		args = append(args, "--show-sensitive")
	}
	root.SetArgs(args)

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

// TestGetCommand_NeverExposesTokenByDefault guards against the general
// json/pretty sensitive-field leak (fixed in internal/output/redact.go)
// for the one context command that renders the real config.Instance
// (including its token) rather than a stripped-down summary.
func TestGetCommand_NeverExposesTokenByDefault(t *testing.T) {
	for _, format := range []string{"table", "json", "pretty"} {
		t.Run(format, func(t *testing.T) {
			rendered := executeGetCommand(t, format, false)
			assert.NotContains(t, rendered, getSentinelToken)
		})
	}
}

func TestGetCommand_ShowSensitiveRevealsToken(t *testing.T) {
	for _, format := range []string{"table", "json", "pretty"} {
		t.Run(format, func(t *testing.T) {
			rendered := executeGetCommand(t, format, true)
			assert.True(t, strings.Contains(rendered, getSentinelToken), "expected token to be revealed with --show-sensitive")
		})
	}
}

package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEnvFile_Simple(t *testing.T) {
	content := `KEY1=value1
KEY2=value2
KEY3=value3`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	envVars, err := ParseEnvFile(tmpFile)
	require.NoError(t, err)
	assert.Len(t, envVars, 3)
	assert.Equal(t, "KEY1", envVars[0].Key)
	assert.Equal(t, "value1", envVars[0].Value)
	assert.Equal(t, "KEY2", envVars[1].Key)
	assert.Equal(t, "value2", envVars[1].Value)
	assert.Equal(t, "KEY3", envVars[2].Key)
	assert.Equal(t, "value3", envVars[2].Value)
}

func TestParseEnvFile_WithQuotes(t *testing.T) {
	content := `KEY1="value with spaces"
KEY2='single quoted value'
KEY3=unquoted`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	envVars, err := ParseEnvFile(tmpFile)
	require.NoError(t, err)
	assert.Len(t, envVars, 3)
	assert.Equal(t, "KEY1", envVars[0].Key)
	assert.Equal(t, "value with spaces", envVars[0].Value)
	assert.Equal(t, "KEY2", envVars[1].Key)
	assert.Equal(t, "single quoted value", envVars[1].Value)
	assert.Equal(t, "KEY3", envVars[2].Key)
	assert.Equal(t, "unquoted", envVars[2].Value)
}

func TestParseEnvFile_WithComments(t *testing.T) {
	content := `# This is a comment
KEY1=value1
# Another comment
KEY2=value2`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	envVars, err := ParseEnvFile(tmpFile)
	require.NoError(t, err)
	assert.Len(t, envVars, 2)
	assert.Equal(t, "KEY1", envVars[0].Key)
	assert.Equal(t, "value1", envVars[0].Value)
	assert.Equal(t, "KEY2", envVars[1].Key)
	assert.Equal(t, "value2", envVars[1].Value)
}

func TestParseEnvFile_WithEmptyLines(t *testing.T) {
	content := `KEY1=value1

KEY2=value2

`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	envVars, err := ParseEnvFile(tmpFile)
	require.NoError(t, err)
	assert.Len(t, envVars, 2)
	assert.Equal(t, "KEY1", envVars[0].Key)
	assert.Equal(t, "value1", envVars[0].Value)
	assert.Equal(t, "KEY2", envVars[1].Key)
	assert.Equal(t, "value2", envVars[1].Value)
}

func TestParseEnvFile_Multiline(t *testing.T) {
	content := `KEY1="line1
line2
line3"
KEY2=single`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	envVars, err := ParseEnvFile(tmpFile)
	require.NoError(t, err)
	assert.Len(t, envVars, 2)
	assert.Equal(t, "KEY1", envVars[0].Key)
	assert.Equal(t, "line1\nline2\nline3", envVars[0].Value)
	assert.Equal(t, "KEY2", envVars[1].Key)
	assert.Equal(t, "single", envVars[1].Value)
}

func TestParseEnvFile_MultilineWithSingleQuotes(t *testing.T) {
	content := `PRIVATE_KEY='-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC
-----END PRIVATE KEY-----'
OTHER=value`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	envVars, err := ParseEnvFile(tmpFile)
	require.NoError(t, err)
	assert.Len(t, envVars, 2)
	assert.Equal(t, "PRIVATE_KEY", envVars[0].Key)
	assert.Contains(t, envVars[0].Value, "BEGIN PRIVATE KEY")
	assert.Contains(t, envVars[0].Value, "END PRIVATE KEY")
	assert.Equal(t, "OTHER", envVars[1].Key)
	assert.Equal(t, "value", envVars[1].Value)
}

func TestParseEnvFile_EmptyValue(t *testing.T) {
	content := `KEY1=
KEY2=value`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	envVars, err := ParseEnvFile(tmpFile)
	require.NoError(t, err)
	assert.Len(t, envVars, 2)
	assert.Equal(t, "KEY1", envVars[0].Key)
	assert.Empty(t, envVars[0].Value)
	assert.Equal(t, "KEY2", envVars[1].Key)
	assert.Equal(t, "value", envVars[1].Value)
}

func TestParseEnvFile_EqualsInValue(t *testing.T) {
	content := `KEY1=value=with=equals
KEY2="quoted=value=with=equals"`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	envVars, err := ParseEnvFile(tmpFile)
	require.NoError(t, err)
	assert.Len(t, envVars, 2)
	assert.Equal(t, "KEY1", envVars[0].Key)
	assert.Equal(t, "value=with=equals", envVars[0].Value)
	assert.Equal(t, "KEY2", envVars[1].Key)
	assert.Equal(t, "quoted=value=with=equals", envVars[1].Value)
}

func TestParseEnvFile_InvalidFormat_MissingEquals(t *testing.T) {
	content := `KEY1
KEY2=value`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	_, err := ParseEnvFile(tmpFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing '='")
}

func TestParseEnvFile_InvalidFormat_EmptyKey(t *testing.T) {
	content := `=value
KEY2=value`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	_, err := ParseEnvFile(tmpFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty key")
}

func TestParseEnvFile_UnclosedQuote(t *testing.T) {
	content := `KEY1="unclosed quote
KEY2=value`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	_, err := ParseEnvFile(tmpFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unclosed quoted value")
}

func TestParseEnvFile_FileNotFound(t *testing.T) {
	_, err := ParseEnvFile("/nonexistent/file.env")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestParseEnvFile_EmptyFile(t *testing.T) {
	content := ``

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	envVars, err := ParseEnvFile(tmpFile)
	require.NoError(t, err)
	assert.Empty(t, envVars)
}

func TestParseEnvFile_OnlyComments(t *testing.T) {
	content := `# Comment 1
# Comment 2
# Comment 3`

	tmpFile := createTempEnvFile(t, content)
	defer os.Remove(tmpFile)

	envVars, err := ParseEnvFile(tmpFile)
	require.NoError(t, err)
	assert.Empty(t, envVars)
}

// Helper function to create a temporary .env file
func createTempEnvFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	require.NoError(t, err)
	return tmpFile
}

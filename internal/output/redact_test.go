package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/config"
	"github.com/coollabsio/coolify-cli/internal/models"
)

const dummySecret = "dummy-secret-abc123"

type sensitiveNested struct {
	Label string `json:"label"`
	Token string `json:"token" sensitive:"true"`
}

type sensitiveStruct struct {
	Name    string            `json:"name"`
	Secret  string            `json:"secret" sensitive:"true"`
	PtrKey  *string           `json:"ptr_key,omitempty" sensitive:"true"`
	Nested  sensitiveNested   `json:"nested"`
	Nests   []sensitiveNested `json:"nests"`
	NestPtr *sensitiveNested  `json:"nest_ptr,omitempty"`
	Tags    map[string]string `json:"tags"`
}

func renderAllFormats(t *testing.T, data any, showSensitive bool) map[string]string {
	t.Helper()
	out := make(map[string]string)
	for _, format := range []string{FormatTable, FormatJSON, FormatPretty} {
		var buf bytes.Buffer
		formatter, err := NewFormatter(format, Options{Writer: &buf, ShowSensitive: showSensitive})
		require.NoError(t, err)
		require.NoError(t, formatter.Format(data))
		out[format] = buf.String()
	}
	return out
}

func TestRedactSensitive_HidesTopLevelAndNestedFields(t *testing.T) {
	ptrVal := dummySecret
	data := sensitiveStruct{
		Name:   "widget",
		Secret: dummySecret,
		PtrKey: &ptrVal,
		Nested: sensitiveNested{Label: "n1", Token: dummySecret},
		Nests: []sensitiveNested{
			{Label: "n2", Token: dummySecret},
		},
		NestPtr: &sensitiveNested{Label: "n3", Token: dummySecret},
		Tags:    map[string]string{"env": "prod"},
	}

	outputs := renderAllFormats(t, data, false)
	for format, rendered := range outputs {
		assert.NotContains(t, rendered, dummySecret, "format %s leaked secret", format)
	}

	// JSON/pretty must still be valid, structurally-correct JSON with the overlay in place.
	for _, format := range []string{FormatJSON, FormatPretty} {
		var decoded map[string]any
		require.NoError(t, json.Unmarshal([]byte(outputs[format]), &decoded), "format %s produced invalid JSON", format)
		assert.Equal(t, SensitiveOverlay, decoded["secret"])
		assert.Equal(t, SensitiveOverlay, decoded["ptr_key"])
		nested, ok := decoded["nested"].(map[string]any)
		require.True(t, ok, "nested field missing in %s output", format)
		assert.Equal(t, SensitiveOverlay, nested["token"])
		assert.Equal(t, "n1", nested["label"])

		nests, ok := decoded["nests"].([]any)
		require.True(t, ok)
		require.Len(t, nests, 1)
		nestItem, ok := nests[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, SensitiveOverlay, nestItem["token"])

		nestPtr, ok := decoded["nest_ptr"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, SensitiveOverlay, nestPtr["token"])

		// Non-sensitive data must round-trip untouched.
		assert.Equal(t, "widget", decoded["name"])
	}
}

func TestRedactSensitive_ShowSensitiveRevealsRealValues(t *testing.T) {
	ptrVal := dummySecret
	data := sensitiveStruct{
		Name:   "widget",
		Secret: dummySecret,
		PtrKey: &ptrVal,
		Nested: sensitiveNested{Label: "n1", Token: dummySecret},
	}

	outputs := renderAllFormats(t, data, true)
	for format, rendered := range outputs {
		assert.Contains(t, rendered, dummySecret, "format %s should reveal secret with ShowSensitive", format)
	}
}

func TestRedactSensitive_DoesNotMutateOriginal(t *testing.T) {
	data := sensitiveStruct{Name: "widget", Secret: dummySecret}

	var buf bytes.Buffer
	formatter, err := NewFormatter(FormatJSON, Options{Writer: &buf, ShowSensitive: false})
	require.NoError(t, err)
	require.NoError(t, formatter.Format(data))

	assert.Equal(t, dummySecret, data.Secret, "original struct must not be mutated by redaction")
}

func TestRedactSensitive_PrivateKeyList(t *testing.T) {
	keys := []models.PrivateKey{
		{UUID: "uuid-1", Name: "key-one", PublicKey: "ssh-ed25519 " + dummySecret, PrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\n" + dummySecret},
	}

	hidden := renderAllFormats(t, keys, false)
	for format, rendered := range hidden {
		assert.NotContains(t, rendered, dummySecret, "format %s leaked private key material", format)
	}

	shown := renderAllFormats(t, keys, true)
	for format, rendered := range shown {
		assert.Contains(t, rendered, dummySecret, "format %s should reveal key material with ShowSensitive", format)
	}
}

func TestRedactSensitive_InstanceToken(t *testing.T) {
	instances := []config.Instance{
		{Name: "prod", FQDN: "https://coolify.example.com", Token: dummySecret, Default: true},
	}

	hidden := renderAllFormats(t, instances, false)
	for format, rendered := range hidden {
		assert.NotContains(t, rendered, dummySecret, "format %s leaked instance token", format)
		assert.NotContains(t, strings.ToLower(rendered), strings.ToLower(dummySecret))
	}

	shown := renderAllFormats(t, instances, true)
	for format, rendered := range shown {
		assert.Contains(t, rendered, dummySecret, "format %s should reveal token with ShowSensitive", format)
	}
}

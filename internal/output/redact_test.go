package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

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

// parityCarrier mixes a sensitive field with the encoding/json edge cases the
// redactor must not disturb: base64 byte slices, a json.Marshaler, nil vs empty
// slices and maps, and every omitempty flavour.
type parityCarrier struct {
	Secret    string            `json:"secret" sensitive:"true"`
	Str       string            `json:"str"`
	Zero      int               `json:"zero"`
	ZeroOmit  int               `json:"zero_omit,omitempty"`
	BoolFalse bool              `json:"bool_false"`
	BoolOmit  bool              `json:"bool_omit,omitempty"`
	NilPtr    *string           `json:"nil_ptr"`
	NilSlice  []string          `json:"nil_slice"`
	EmptySl   []string          `json:"empty_sl"`
	NilMap    map[string]string `json:"nil_map"`
	EmptyMap  map[string]string `json:"empty_map"`
	EmptyArr  [0]int            `json:"empty_arr,omitempty"`
	Arr       [2]int            `json:"arr"`
	Iface     any               `json:"iface"`
	Bytes     []byte            `json:"bytes"`
	When      time.Time         `json:"when"`
	Float     float64           `json:"float"`
	Excluded  string            `json:"-"`
}

// TestRedactSensitive_LeavesNonSensitiveOutputByteIdentical is the guard
// against the redactor quietly reshaping output it only meant to pass through.
// Everything except the redacted field must marshal exactly as encoding/json
// would have marshaled it.
func TestRedactSensitive_LeavesNonSensitiveOutputByteIdentical(t *testing.T) {
	data := parityCarrier{
		Secret:   dummySecret,
		Str:      "x",
		Arr:      [2]int{1, 2},
		EmptySl:  []string{},
		Bytes:    []byte("hello"),
		When:     time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		Float:    1.5,
		Excluded: dummySecret,
	}

	expectation := data
	expectation.Secret = SensitiveOverlay
	want, err := json.Marshal(expectation)
	require.NoError(t, err)

	got, err := json.Marshal(redactSensitive(data, false))
	require.NoError(t, err)

	assert.JSONEq(t, string(want), string(got))
	assert.Equal(t, string(want), string(got), "redaction must not reshape non-sensitive output")
}

// TestRedactSensitive_MapKeysStayStablyOrdered guards against rendering maps
// through an insertion-ordered container, which would expose Go's randomized
// map iteration order instead of encoding/json's sorted keys.
func TestRedactSensitive_MapKeysStayStablyOrdered(t *testing.T) {
	type carrier struct {
		Options map[string]string `json:"options"`
		Secret  string            `json:"secret" sensitive:"true"`
	}
	data := carrier{
		Options: map[string]string{"a": "1", "b": "2", "c": "3", "d": "4", "e": "5", "f": "6", "g": "7", "h": "8"},
		Secret:  dummySecret,
	}

	first, err := json.Marshal(redactSensitive(data, false))
	require.NoError(t, err)

	for i := 0; i < 50; i++ {
		again, err := json.Marshal(redactSensitive(data, false))
		require.NoError(t, err)
		require.Equal(t, string(first), string(again), "map key order must be stable across renders")
	}

	expectation := data
	expectation.Secret = SensitiveOverlay
	want, err := json.Marshal(expectation)
	require.NoError(t, err)
	assert.Equal(t, string(want), string(first))
}

type embeddedBase struct {
	Shared    string `json:"shared"`
	BaseToken string `json:"base_token" sensitive:"true"`
}

type embeddingOuter struct {
	embeddedBase
	Name string `json:"name"`
}

// TestRedactSensitive_PromotesEmbeddedFields covers encoding/json's promotion
// of embedded struct fields into the parent object. Nesting them under the type
// name (or dropping them, when the embedded type is unexported) would silently
// change the shape of any model that gains an embedded struct.
func TestRedactSensitive_PromotesEmbeddedFields(t *testing.T) {
	data := embeddingOuter{
		embeddedBase: embeddedBase{Shared: "visible", BaseToken: dummySecret},
		Name:         "outer",
	}

	rendered, err := json.Marshal(redactSensitive(data, false))
	require.NoError(t, err)

	assert.NotContains(t, string(rendered), dummySecret)
	assert.Equal(t, `{"shared":"visible","base_token":"********","name":"outer"}`, string(rendered))
}

type marshalerCarrier struct {
	Inner sensitiveNested
}

func (m marshalerCarrier) MarshalJSON() ([]byte, error) { return json.Marshal(m.Inner) }

// TestRedactSensitive_RedactsThroughCustomMarshaler ensures a type's own
// MarshalJSON cannot be used to smuggle a sensitive field past the redactor.
func TestRedactSensitive_RedactsThroughCustomMarshaler(t *testing.T) {
	data := struct {
		Wrapped marshalerCarrier `json:"wrapped"`
	}{Wrapped: marshalerCarrier{Inner: sensitiveNested{Label: "l", Token: dummySecret}}}

	rendered, err := json.Marshal(redactSensitive(data, false))
	require.NoError(t, err)
	assert.NotContains(t, string(rendered), dummySecret)
}

type selfReferential struct {
	Name   string           `json:"name"`
	Secret string           `json:"secret" sensitive:"true"`
	Self   *selfReferential `json:"self,omitempty"`
}

// TestRedactSensitive_SurvivesSelfReferentialValue pins the cycle guard. Without
// it the walk recurses until the goroutine stack overflows, which is a fatal
// error Go cannot recover from — it kills the CLI outright.
func TestRedactSensitive_SurvivesSelfReferentialValue(t *testing.T) {
	data := &selfReferential{Name: "a", Secret: dummySecret}
	data.Self = data

	rendered, err := json.Marshal(redactSensitive(data, false))
	require.NoError(t, err)
	assert.NotContains(t, string(rendered), dummySecret)
}

// TestRedactSensitive_DatabaseCredentials covers the database credentials that
// `database get`/`database list` return on the real model.
func TestRedactSensitive_DatabaseCredentials(t *testing.T) {
	password := dummySecret
	databases := []models.Database{{
		UUID:                    "db-1",
		Name:                    "primary",
		PostgresPassword:        &password,
		MysqlRootPassword:       &password,
		MysqlPassword:           &password,
		MariadbRootPassword:     &password,
		MariadbPassword:         &password,
		MongoInitdbRootPassword: &password,
		RedisPassword:           &password,
		KeydbPassword:           &password,
		ClickhouseAdminPassword: &password,
		DragonflyPassword:       &password,
	}}

	hidden := renderAllFormats(t, databases, false)
	for format, rendered := range hidden {
		assert.NotContains(t, rendered, dummySecret, "format %s leaked a database password", format)
	}

	// Table output excludes these columns entirely (`table:"-"`), so only the
	// structured formats can reveal them.
	shown := renderAllFormats(t, databases, true)
	for _, format := range []string{FormatJSON, FormatPretty} {
		assert.Contains(t, shown[format], dummySecret, "format %s should reveal database passwords with ShowSensitive", format)
	}
}

// TestRedactSensitive_OmitsEmptySensitivePointer keeps an unset optional secret
// out of the output entirely rather than claiming a value exists.
func TestRedactSensitive_OmitsEmptySensitivePointer(t *testing.T) {
	rendered, err := json.Marshal(redactSensitive(sensitiveStruct{Name: "widget"}, false))
	require.NoError(t, err)
	assert.NotContains(t, string(rendered), "ptr_key")
}

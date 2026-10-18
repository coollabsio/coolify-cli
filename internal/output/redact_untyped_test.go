package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestRedact_UntypedMapCredentialKeys(t *testing.T) {
	data := map[string]any{
		"telegram_enabled":  true,
		"telegram_token":    "123:ABC",
		"smtp_password":     "hunter2",
		"smtp_host":         "mail.example.com",
		"resend_api_key":    "",
		"pushover_token":    nil,
		"nested":            map[string]any{"discord_webhook_url": "https://discord.example/webhooks/1/SECRET", "label": "x"},
		"list":              []any{map[string]any{"sentinel_token": "sent-tok"}},
		"use_build_secrets": true,
	}

	var hidden bytes.Buffer
	f, err := NewFormatter(FormatJSON, Options{Writer: &hidden})
	require.NoError(t, err)
	require.NoError(t, f.Format(data))

	for _, secret := range []string{"123:ABC", "hunter2", "SECRET", "sent-tok"} {
		assert.NotContains(t, hidden.String(), secret)
	}

	var got map[string]any
	require.NoError(t, json.Unmarshal(hidden.Bytes(), &got))
	assert.Equal(t, "********", got["telegram_token"])
	assert.Equal(t, "********", got["smtp_password"])
	assert.Equal(t, "mail.example.com", got["smtp_host"])
	assert.Equal(t, true, got["telegram_enabled"], "bool values keep their type")
	assert.Equal(t, true, got["use_build_secrets"], "bool values keep their type")
	assert.Contains(t, got, "resend_api_key")
	assert.Empty(t, got["resend_api_key"], "empty values stay empty so unset secrets are not reported as set")
	assert.Nil(t, got["pushover_token"])

	var shown bytes.Buffer
	f, err = NewFormatter(FormatJSON, Options{Writer: &shown, ShowSensitive: true})
	require.NoError(t, err)
	require.NoError(t, f.Format(data))
	assert.Contains(t, shown.String(), "123:ABC")
	assert.Contains(t, shown.String(), "hunter2")
}

func TestRedact_AuditEventUntypedPayloadAndPII(t *testing.T) {
	email, ip := "dev@example.com", "203.0.113.9"
	event := models.AuditEvent{
		ActorEmail: &email,
		IPAddress:  &ip,
		Changes:    map[string]any{"token": "tok-secret", "name": "visible"},
		Metadata:   map[string]any{"smtp_password": "meta-secret"},
	}

	var buf bytes.Buffer
	f, err := NewFormatter(FormatPretty, Options{Writer: &buf})
	require.NoError(t, err)
	require.NoError(t, f.Format([]models.AuditEvent{event}))

	for _, secret := range []string{email, ip, "tok-secret", "meta-secret"} {
		assert.NotContains(t, buf.String(), secret)
	}
	assert.Contains(t, buf.String(), "visible")
}

func TestRedact_NilSensitivePointerStaysNull(t *testing.T) {
	var buf bytes.Buffer
	f, err := NewFormatter(FormatJSON, Options{Writer: &buf})
	require.NoError(t, err)
	require.NoError(t, f.Format([]models.ServerRegistry{{Registry: "ghcr.io"}}))

	var got []map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Nil(t, got[0]["username"], "an absent credential must not be rendered as a present, masked one")
}

func TestRedact_EmptySensitiveValuesAreNotMasked(t *testing.T) {
	data := map[string]any{"server": models.Server{IP: "", User: "", Port: 0}}

	var buf bytes.Buffer
	f, err := NewFormatter(FormatJSON, Options{Writer: &buf})
	require.NoError(t, err)
	require.NoError(t, f.Format(data))

	var got map[string]map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Empty(t, got["server"]["ip"])
	assert.Empty(t, got["server"]["user"])
	assert.InDelta(t, 0, got["server"]["port"], 0, "an unset port stays a number")
}

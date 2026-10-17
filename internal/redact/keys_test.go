package redact

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCredentialKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"telegram_token", true},
		{"sentinel_token", true},
		{"smtp_password", true},
		{"resend_api_key", true},
		{"logdrain_axiom_api_key", true},
		{"pushover_user_key", true},
		{"geoip_maxmind_license_key", true},
		{"discord_webhook_url", true},
		{"client_secret", true},
		{"Telegram_Token", true},
		{"token_name", false},
		{"actor_token_name", false},
		{"integration_token_uuid", false},
		{"private_key_uuid", false},
		{"hetzner_ssh_key_ids", false},
		{"cloud_provider_token_uuid", false},
		{"smtp_password_set_at", false},
		{"telegram_enabled", false},
		{"fqdn", false},
		{"key", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.want, IsCredentialKey(tt.key))
		})
	}
}

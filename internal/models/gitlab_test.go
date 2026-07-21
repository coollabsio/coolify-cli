package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitLabAppCreateRequestOmitsOptionalFields(t *testing.T) {
	req := GitLabAppCreateRequest{
		Name:    "app",
		HTMLURL: "https://gitlab.com",
	}

	payload, err := json.Marshal(req)
	require.NoError(t, err)
	require.NotContains(t, string(payload), "api_url")
	require.NotContains(t, string(payload), "client_secret")
	require.NotContains(t, string(payload), "webhook_token")
	require.Contains(t, string(payload), "html_url")
	require.Contains(t, string(payload), "name")
}

func TestGitLabAppCreateRequestIncludesSecretsWhenSet(t *testing.T) {
	secret := "s3cret"
	token := "hook"
	req := GitLabAppCreateRequest{
		Name:         "app",
		HTMLURL:      "https://gitlab.com",
		ClientSecret: &secret,
		WebhookToken: &token,
	}

	payload, err := json.Marshal(req)
	require.NoError(t, err)
	require.Contains(t, string(payload), "client_secret")
	require.Contains(t, string(payload), "webhook_token")
}

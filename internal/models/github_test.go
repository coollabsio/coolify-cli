package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitHubAppCreateRequestOmitsDerivedAPIURL(t *testing.T) {
	req := GitHubAppCreateRequest{
		Name:           "app",
		HTMLURL:        "https://github.example.com",
		AppID:          1,
		InstallationID: 2,
		ClientID:       "client",
		ClientSecret:   "secret",
		PrivateKeyUUID: "key-uuid",
	}

	payload, err := json.Marshal(req)
	require.NoError(t, err)
	require.NotContains(t, string(payload), "api_url")
	require.Contains(t, string(payload), "html_url")
}

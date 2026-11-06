package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The Coolify API reads the environment UUID from "environment_uuid".
// Any other key is rejected with a 422.
func TestCreateRequests_EnvironmentUUIDJSONKey(t *testing.T) {
	uuid := "env-uuid"
	tests := map[string]any{
		"public":     ApplicationCreatePublicRequest{EnvironmentUUID: &uuid},
		"github":     ApplicationCreateGitHubAppRequest{EnvironmentUUID: &uuid},
		"deploykey":  ApplicationCreateDeployKeyRequest{EnvironmentUUID: &uuid},
		"dockerfile": ApplicationCreateDockerfileRequest{EnvironmentUUID: &uuid},
		"dockerimg":  ApplicationCreateDockerImageRequest{EnvironmentUUID: &uuid},
		"service":    ServiceCreateRequest{EnvironmentUUID: &uuid},
	}

	for name, req := range tests {
		t.Run(name, func(t *testing.T) {
			data, err := json.Marshal(req)
			require.NoError(t, err)

			var got map[string]any
			require.NoError(t, json.Unmarshal(data, &got))
			assert.Equal(t, "env-uuid", got["environment_uuid"])
			assert.NotContains(t, got, "environment")
		})
	}
}

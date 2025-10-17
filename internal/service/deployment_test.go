package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeploymentService_Deploy(t *testing.T) {
	tests := []struct {
		name          string
		uuid          string
		force         bool
		expectedPath  string
		response      DeployResponse
	}{
		{
			name:         "deploy without force",
			uuid:         "res-123",
			force:        false,
			expectedPath: "/api/v1/deploy?uuid=res-123",
			response: DeployResponse{
				Deployments: []DeploymentInfo{
					{
						Message:        "Deployment started",
						ResourceUUID:   "res-123",
						DeploymentUUID: "dep-456",
					},
				},
			},
		},
		{
			name:         "deploy with force",
			uuid:         "res-789",
			force:        true,
			expectedPath: "/api/v1/deploy?uuid=res-789&force=true",
			response: DeployResponse{
				Deployments: []DeploymentInfo{
					{
						Message:        "Force deployment started",
						ResourceUUID:   "res-789",
						DeploymentUUID: "dep-999",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedPath, r.URL.Path+"?"+r.URL.RawQuery)
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewDeploymentService(client)

			result, err := svc.Deploy(context.Background(), tt.uuid, tt.force)
			require.NoError(t, err)
			assert.Len(t, result.Deployments, len(tt.response.Deployments))
			if len(result.Deployments) > 0 {
				assert.Equal(t, tt.response.Deployments[0].Message, result.Deployments[0].Message)
				assert.Equal(t, tt.response.Deployments[0].DeploymentUUID, result.Deployments[0].DeploymentUUID)
			}
		})
	}
}

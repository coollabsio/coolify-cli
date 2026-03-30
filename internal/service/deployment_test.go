package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestDeploymentService_Deploy(t *testing.T) {
	tests := []struct {
		name           string
		request        models.DeployRequest
		expectedPath   string
		expectedMethod string
		expectedBody   string
		response       DeployResponse
	}{
		{
			name:           "deploy without optional fields",
			request:        models.DeployRequest{UUID: "res-123"},
			expectedPath:   "/api/v1/deploy",
			expectedMethod: "POST",
			expectedBody:   `{"uuid":"res-123"}`,
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
			name: "deploy with force and extra payload fields",
			request: models.DeployRequest{
				UUID:          "res-789",
				Force:         deployBoolPtr(true),
				PullRequestID: deployIntPtr(2345),
				DockerTag:     deployStringPtr("1.28.3"),
			},
			expectedPath:   "/api/v1/deploy",
			expectedMethod: "POST",
			expectedBody:   `{"uuid":"res-789","force":true,"pull_request_id":2345,"docker_tag":"1.28.3"}`,
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
				assert.Equal(t, tt.expectedPath, r.URL.Path)
				assert.Equal(t, tt.expectedMethod, r.Method)

				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, tt.expectedBody, string(body))

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewDeploymentService(client)

			result, err := svc.Deploy(context.Background(), tt.request)
			require.NoError(t, err)
			assert.Len(t, result.Deployments, len(tt.response.Deployments))
			if len(result.Deployments) > 0 {
				assert.Equal(t, tt.response.Deployments[0].Message, result.Deployments[0].Message)
				assert.Equal(t, tt.response.Deployments[0].DeploymentUUID, result.Deployments[0].DeploymentUUID)
			}
		})
	}
}

func deployBoolPtr(v bool) *bool {
	return &v
}

func deployIntPtr(v int) *int {
	return &v
}

func deployStringPtr(v string) *string {
	return &v
}

func TestDeploymentService_ListByApplication(t *testing.T) {
	tests := []struct {
		name         string
		appUUID      string
		expectedPath string
		response     string
		wantErr      bool
		wantCount    int
	}{
		{
			name:         "successful list with deployments",
			appUUID:      "app-123",
			expectedPath: "/api/v1/deployments/applications/app-123",
			response: `{
				"count": 2,
				"deployments": [
					{
						"id": 1,
						"deployment_uuid": "dep-123",
						"application_name": "my-app",
						"server_name": "server-1",
						"status": "finished",
						"commit": "abc123"
					},
					{
						"id": 2,
						"deployment_uuid": "dep-456",
						"application_name": "my-app",
						"server_name": "server-1",
						"status": "in_progress",
						"commit": "def456"
					}
				]
			}`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:         "empty list",
			appUUID:      "app-empty",
			expectedPath: "/api/v1/deployments/applications/app-empty",
			response:     `{"count": 0, "deployments": []}`,
			wantErr:      false,
			wantCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedPath, r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewDeploymentService(client)

			result, err := svc.ListByApplication(context.Background(), tt.appUUID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}

func TestDeploymentService_GetLogsByApplication(t *testing.T) {
	tests := []struct {
		name              string
		appUUID           string
		lines             int
		deploymentsList   string
		deploymentDetails string
		expectedLogs      string
		wantErr           bool
		noDeployments     bool
		emptyLogs         bool
	}{
		{
			name:    "get logs successfully",
			appUUID: "app-123",
			lines:   0,
			deploymentsList: `{
				"count": 1,
				"deployments": [
					{
						"id": 1,
						"deployment_uuid": "dep-123",
						"application_name": "my-app",
						"status": "finished"
					}
				]
			}`,
			deploymentDetails: `{
				"id": 1,
				"deployment_uuid": "dep-123",
				"application_name": "my-app",
				"status": "finished",
				"logs": "Starting deployment...\nBuilding image...\nDeployment successful!"
			}`,
			expectedLogs: "Starting deployment...\nBuilding image...\nDeployment successful!",
			wantErr:      false,
		},
		{
			name:            "no deployments found",
			appUUID:         "app-empty",
			lines:           0,
			deploymentsList: `{"count": 0, "deployments": []}`,
			expectedLogs:    "No deployments found for this application",
			wantErr:         false,
			noDeployments:   true,
		},
		{
			name:    "empty logs",
			appUUID: "app-no-logs",
			lines:   0,
			deploymentsList: `{
				"count": 1,
				"deployments": [
					{
						"id": 1,
						"deployment_uuid": "dep-456",
						"application_name": "my-app",
						"status": "in_progress"
					}
				]
			}`,
			deploymentDetails: `{
				"id": 1,
				"deployment_uuid": "dep-456",
				"application_name": "my-app",
				"status": "in_progress",
				"logs": null
			}`,
			expectedLogs: "No logs available for deployment dep-456 (Status: in_progress)",
			wantErr:      false,
			emptyLogs:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				// Handle list deployments request
				if r.URL.Path == "/api/v1/deployments/applications/"+tt.appUUID {
					_, _ = w.Write([]byte(tt.deploymentsList))
					return
				}

				// Handle get deployment details request
				if !tt.noDeployments && r.URL.Path == "/api/v1/deployments/dep-123" || r.URL.Path == "/api/v1/deployments/dep-456" {
					_, _ = w.Write([]byte(tt.deploymentDetails))
					return
				}

				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewDeploymentService(client)

			result, err := svc.GetLogsByApplication(context.Background(), tt.appUUID, tt.lines)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLogs, result)
			}
		})
	}
}

func TestDeploymentService_ListByApplication_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "application not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDeploymentService(client)

	_, err := svc.ListByApplication(context.Background(), "nonexistent")
	require.Error(t, err)
}

func TestDeploymentService_GetLogsByApplication_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDeploymentService(client)

	_, err := svc.GetLogsByApplication(context.Background(), "app-123", 0)
	require.Error(t, err)
}

func TestDeploymentService_GetLogsByDeployment(t *testing.T) {
	tests := []struct {
		name             string
		deploymentUUID   string
		deploymentDetail string
		expectedLogs     string
		wantErr          bool
	}{
		{
			name:           "get logs successfully",
			deploymentUUID: "dep-123",
			deploymentDetail: `{
				"id": 1,
				"deployment_uuid": "dep-123",
				"application_name": "my-app",
				"status": "finished",
				"logs": "Starting deployment...\nBuilding image...\nDeployment successful!"
			}`,
			expectedLogs: "Starting deployment...\nBuilding image...\nDeployment successful!",
			wantErr:      false,
		},
		{
			name:           "empty logs",
			deploymentUUID: "dep-no-logs",
			deploymentDetail: `{
				"id": 1,
				"deployment_uuid": "dep-no-logs",
				"application_name": "my-app",
				"status": "in_progress",
				"logs": null
			}`,
			expectedLogs: "No logs available for deployment dep-no-logs (Status: in_progress)",
			wantErr:      false,
		},
		{
			name:           "logs with empty string",
			deploymentUUID: "dep-empty-string",
			deploymentDetail: `{
				"id": 1,
				"deployment_uuid": "dep-empty-string",
				"application_name": "my-app",
				"status": "queued",
				"logs": ""
			}`,
			expectedLogs: "No logs available for deployment dep-empty-string (Status: queued)",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				// Handle get deployment details request
				if r.URL.Path == "/api/v1/deployments/"+tt.deploymentUUID {
					_, _ = w.Write([]byte(tt.deploymentDetail))
					return
				}

				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			svc := NewDeploymentService(client)

			result, err := svc.GetLogsByDeployment(context.Background(), tt.deploymentUUID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLogs, result)
			}
		})
	}
}

func TestDeploymentService_GetLogsByDeployment_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "deployment not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDeploymentService(client)

	_, err := svc.GetLogsByDeployment(context.Background(), "nonexistent")
	require.Error(t, err)
}

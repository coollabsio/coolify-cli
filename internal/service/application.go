package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// ApplicationService handles application-related operations
type ApplicationService struct {
	client *api.Client
}

// NewApplicationService creates a new application service
func NewApplicationService(client *api.Client) *ApplicationService {
	return &ApplicationService{
		client: client,
	}
}

// List retrieves all applications
func (s *ApplicationService) List(ctx context.Context) ([]models.Application, error) {
	var apps []models.Application
	err := s.client.Get(ctx, "applications", &apps)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}
	return apps, nil
}

// Get retrieves a specific application by UUID
func (s *ApplicationService) Get(ctx context.Context, uuid string) (*models.Application, error) {
	var app models.Application
	err := s.client.Get(ctx, fmt.Sprintf("applications/%s", uuid), &app)
	if err != nil {
		return nil, fmt.Errorf("failed to get application %s: %w", uuid, err)
	}
	return &app, nil
}

// Update updates an application
func (s *ApplicationService) Update(ctx context.Context, uuid string, req models.ApplicationUpdateRequest) (*models.Application, error) {
	var app models.Application
	err := s.client.Patch(ctx, fmt.Sprintf("applications/%s", uuid), req, &app)
	if err != nil {
		return nil, fmt.Errorf("failed to update application %s: %w", uuid, err)
	}
	return &app, nil
}

// Delete deletes an application
func (s *ApplicationService) Delete(ctx context.Context, uuid string) error {
	err := s.client.Delete(ctx, fmt.Sprintf("applications/%s", uuid))
	if err != nil {
		return fmt.Errorf("failed to delete application %s: %w", uuid, err)
	}
	return nil
}

// Start starts an application (initiates deployment)
func (s *ApplicationService) Start(ctx context.Context, uuid string, force bool, instantDeploy bool) (*models.ApplicationLifecycleResponse, error) {
	var resp models.ApplicationLifecycleResponse

	// Build URL with query parameters
	url := fmt.Sprintf("applications/%s/start", uuid)
	if force || instantDeploy {
		url += "?"
		if force {
			url += "force=true"
		}
		if instantDeploy {
			if force {
				url += "&"
			}
			url += "instant_deploy=true"
		}
	}

	err := s.client.Get(ctx, url, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to start application %s: %w", uuid, err)
	}
	return &resp, nil
}

// Stop stops an application
func (s *ApplicationService) Stop(ctx context.Context, uuid string) (*models.ApplicationLifecycleResponse, error) {
	var resp models.ApplicationLifecycleResponse
	err := s.client.Get(ctx, fmt.Sprintf("applications/%s/stop", uuid), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to stop application %s: %w", uuid, err)
	}
	return &resp, nil
}

// Restart restarts an application
func (s *ApplicationService) Restart(ctx context.Context, uuid string) (*models.ApplicationLifecycleResponse, error) {
	var resp models.ApplicationLifecycleResponse
	err := s.client.Get(ctx, fmt.Sprintf("applications/%s/restart", uuid), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to restart application %s: %w", uuid, err)
	}
	return &resp, nil
}

// Logs retrieves logs for an application
func (s *ApplicationService) Logs(ctx context.Context, uuid string, lines int) (*models.ApplicationLogsResponse, error) {
	url := fmt.Sprintf("applications/%s/logs", uuid)

	// Add lines parameter if specified
	if lines > 0 {
		url = fmt.Sprintf("%s?lines=%d", url, lines)
	}

	var resp models.ApplicationLogsResponse
	err := s.client.Get(ctx, url, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for application %s: %w", uuid, err)
	}
	return &resp, nil
}

// ListEnvs retrieves all environment variables for an application
func (s *ApplicationService) ListEnvs(ctx context.Context, uuid string) ([]models.EnvironmentVariable, error) {
	var envs []models.EnvironmentVariable
	err := s.client.Get(ctx, fmt.Sprintf("applications/%s/envs", uuid), &envs)
	if err != nil {
		return nil, fmt.Errorf("failed to list environment variables for application %s: %w", uuid, err)
	}
	return envs, nil
}

// CreateEnv creates a new environment variable for an application
func (s *ApplicationService) CreateEnv(ctx context.Context, uuid string, req *models.EnvironmentVariableCreateRequest) (*models.EnvironmentVariable, error) {
	var env models.EnvironmentVariable
	err := s.client.Post(ctx, fmt.Sprintf("applications/%s/envs", uuid), req, &env)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment variable for application %s: %w", uuid, err)
	}
	return &env, nil
}

// UpdateEnv updates an existing environment variable for an application
func (s *ApplicationService) UpdateEnv(ctx context.Context, appUUID string, req *models.EnvironmentVariableUpdateRequest) (*models.EnvironmentVariable, error) {
	var env models.EnvironmentVariable
	err := s.client.Patch(ctx, fmt.Sprintf("applications/%s/envs", appUUID), req, &env)
	if err != nil {
		return nil, fmt.Errorf("failed to update environment variable for application %s: %w", appUUID, err)
	}
	return &env, nil
}

// GetEnv retrieves a single environment variable by UUID or key
func (s *ApplicationService) GetEnv(ctx context.Context, appUUID, envIdentifier string) (*models.EnvironmentVariable, error) {
	envs, err := s.ListEnvs(ctx, appUUID)
	if err != nil {
		return nil, err
	}

	// Try to find by UUID first, then by key
	for _, env := range envs {
		if env.UUID == envIdentifier || env.Key == envIdentifier {
			return &env, nil
		}
	}

	return nil, fmt.Errorf("environment variable '%s' not found in application %s", envIdentifier, appUUID)
}

// DeleteEnv deletes an environment variable from an application
func (s *ApplicationService) DeleteEnv(ctx context.Context, appUUID string, envUUID string) error {
	err := s.client.Delete(ctx, fmt.Sprintf("applications/%s/envs/%s", appUUID, envUUID))
	if err != nil {
		return fmt.Errorf("failed to delete environment variable %s for application %s: %w", envUUID, appUUID, err)
	}
	return nil
}

// BulkUpdateEnvsRequest represents a bulk update request for environment variables
type BulkUpdateEnvsRequest struct {
	Data []models.EnvironmentVariableCreateRequest `json:"data"`
}

// BulkUpdateEnvsResponse represents the response from bulk update
type BulkUpdateEnvsResponse struct {
	Message string `json:"message"`
}

// BulkUpdateEnvs updates multiple environment variables in a single request
func (s *ApplicationService) BulkUpdateEnvs(ctx context.Context, appUUID string, req *BulkUpdateEnvsRequest) (*BulkUpdateEnvsResponse, error) {
	var response BulkUpdateEnvsResponse
	err := s.client.Patch(ctx, fmt.Sprintf("applications/%s/envs/bulk", appUUID), req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update environment variables for application %s: %w", appUUID, err)
	}
	return &response, nil
}

// CreatePublic creates an application from a public git repository
func (s *ApplicationService) CreatePublic(ctx context.Context, req *models.ApplicationCreatePublicRequest) (*models.Application, error) {
	var app models.Application
	err := s.client.Post(ctx, "applications/public", req, &app)
	if err != nil {
		return nil, fmt.Errorf("failed to create application from public repository: %w", err)
	}
	return &app, nil
}

// CreateGitHubApp creates an application from a private repository using GitHub App
func (s *ApplicationService) CreateGitHubApp(ctx context.Context, req *models.ApplicationCreateGitHubAppRequest) (*models.Application, error) {
	var app models.Application
	err := s.client.Post(ctx, "applications/private-github-app", req, &app)
	if err != nil {
		return nil, fmt.Errorf("failed to create application from private GitHub repository: %w", err)
	}
	return &app, nil
}

// CreateDeployKey creates an application from a private repository using SSH deploy key
func (s *ApplicationService) CreateDeployKey(ctx context.Context, req *models.ApplicationCreateDeployKeyRequest) (*models.Application, error) {
	var app models.Application
	err := s.client.Post(ctx, "applications/private-deploy-key", req, &app)
	if err != nil {
		return nil, fmt.Errorf("failed to create application from private repository with deploy key: %w", err)
	}
	return &app, nil
}

// CreateDockerfile creates an application from a custom Dockerfile
func (s *ApplicationService) CreateDockerfile(ctx context.Context, req *models.ApplicationCreateDockerfileRequest) (*models.Application, error) {
	var app models.Application
	err := s.client.Post(ctx, "applications/dockerfile", req, &app)
	if err != nil {
		return nil, fmt.Errorf("failed to create application from Dockerfile: %w", err)
	}
	return &app, nil
}

// CreateDockerImage creates an application from a pre-built Docker image
func (s *ApplicationService) CreateDockerImage(ctx context.Context, req *models.ApplicationCreateDockerImageRequest) (*models.Application, error) {
	var app models.Application
	err := s.client.Post(ctx, "applications/dockerimage", req, &app)
	if err != nil {
		return nil, fmt.Errorf("failed to create application from Docker image: %w", err)
	}
	return &app, nil
}

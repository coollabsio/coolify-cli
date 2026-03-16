package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// Service handles service-related operations
type Service struct {
	client *api.Client
}

// NewService creates a new service instance
func NewService(client *api.Client) *Service {
	return &Service{client: client}
}

// List retrieves all services
func (s *Service) List(ctx context.Context) ([]models.Service, error) {
	var services []models.Service
	err := s.client.Get(ctx, "services", &services)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	return services, nil
}

// Get retrieves a service by UUID
func (s *Service) Get(ctx context.Context, uuid string) (*models.Service, error) {
	var service models.Service
	err := s.client.Get(ctx, fmt.Sprintf("services/%s", uuid), &service)
	if err != nil {
		return nil, fmt.Errorf("failed to get service %s: %w", uuid, err)
	}
	return &service, nil
}

// Create creates a new service
func (s *Service) Create(ctx context.Context, req *models.ServiceCreateRequest) (*models.Service, error) {
	var service models.Service
	err := s.client.Post(ctx, "services", req, &service)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}
	return &service, nil
}

// Update updates a service
func (s *Service) Update(ctx context.Context, uuid string, req *models.ServiceUpdateRequest) (*models.Service, error) {
	var service models.Service
	err := s.client.Patch(ctx, fmt.Sprintf("services/%s", uuid), req, &service)
	if err != nil {
		return nil, fmt.Errorf("failed to update service %s: %w", uuid, err)
	}
	return &service, nil
}

// Delete deletes a service
func (s *Service) Delete(ctx context.Context, uuid string, deleteConfigurations, deleteVolumes, dockerCleanup, deleteConnectedNetworks bool) error {
	url := fmt.Sprintf("services/%s?delete_configurations=%t&delete_volumes=%t&docker_cleanup=%t&delete_connected_networks=%t",
		uuid, deleteConfigurations, deleteVolumes, dockerCleanup, deleteConnectedNetworks)

	err := s.client.Delete(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to delete service %s: %w", uuid, err)
	}
	return nil
}

// Start starts a service
func (s *Service) Start(ctx context.Context, uuid string) (*models.ServiceLifecycleResponse, error) {
	var resp models.ServiceLifecycleResponse
	err := s.client.Post(ctx, fmt.Sprintf("services/%s/start", uuid), nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to start service %s: %w", uuid, err)
	}
	return &resp, nil
}

// Stop stops a service
func (s *Service) Stop(ctx context.Context, uuid string) (*models.ServiceLifecycleResponse, error) {
	var resp models.ServiceLifecycleResponse
	err := s.client.Post(ctx, fmt.Sprintf("services/%s/stop", uuid), nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to stop service %s: %w", uuid, err)
	}
	return &resp, nil
}

// Restart restarts a service
func (s *Service) Restart(ctx context.Context, uuid string) (*models.ServiceLifecycleResponse, error) {
	var resp models.ServiceLifecycleResponse
	err := s.client.Post(ctx, fmt.Sprintf("services/%s/restart", uuid), nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to restart service %s: %w", uuid, err)
	}
	return &resp, nil
}

// ListEnvs retrieves all environment variables for a service
func (s *Service) ListEnvs(ctx context.Context, uuid string) ([]models.ServiceEnvironmentVariable, error) {
	var envs []models.ServiceEnvironmentVariable
	err := s.client.Get(ctx, fmt.Sprintf("services/%s/envs", uuid), &envs)
	if err != nil {
		return nil, fmt.Errorf("failed to list environment variables for service %s: %w", uuid, err)
	}
	return envs, nil
}

// GetEnv retrieves a single environment variable by UUID or key
func (s *Service) GetEnv(ctx context.Context, serviceUUID, envIdentifier string) (*models.ServiceEnvironmentVariable, error) {
	envs, err := s.ListEnvs(ctx, serviceUUID)
	if err != nil {
		return nil, err
	}

	// Try to find by UUID first, then by key
	for _, env := range envs {
		if env.UUID == envIdentifier || env.Key == envIdentifier {
			return &env, nil
		}
	}

	return nil, fmt.Errorf("environment variable '%s' not found in service %s", envIdentifier, serviceUUID)
}

// CreateEnv creates a new environment variable for a service
func (s *Service) CreateEnv(ctx context.Context, uuid string, req *models.ServiceEnvironmentVariableCreateRequest) (*models.ServiceEnvironmentVariable, error) {
	var env models.ServiceEnvironmentVariable
	err := s.client.Post(ctx, fmt.Sprintf("services/%s/envs", uuid), req, &env)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment variable for service %s: %w", uuid, err)
	}
	return &env, nil
}

// UpdateEnv updates an environment variable for a service.
// Uses the bulk update endpoint as the single-update endpoint
// (PATCH /services/{uuid}/envs) returns 404 for services.
// See: https://github.com/coollabsio/coolify-cli/issues/48
func (s *Service) UpdateEnv(ctx context.Context, serviceUUID string, req *models.ServiceEnvironmentVariableUpdateRequest) (*models.ServiceEnvironmentVariable, error) {
	// Fetch the current env var to get existing values for fields not being updated
	existing, err := s.GetEnv(ctx, serviceUUID, req.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch environment variable for update: %w", err)
	}

	// Build the bulk update entry, applying requested changes over existing values
	entry := models.ServiceEnvironmentVariableCreateRequest{
		Key:   existing.Key,
		Value: existing.Value,
	}
	if req.Key != nil {
		entry.Key = *req.Key
	}
	if req.Value != nil {
		entry.Value = *req.Value
	}
	isBuildTime := existing.IsBuildTime
	if req.IsBuildTime != nil {
		isBuildTime = *req.IsBuildTime
	}
	entry.IsBuildTime = &isBuildTime
	isLiteral := existing.IsLiteralValue
	if req.IsLiteral != nil {
		isLiteral = *req.IsLiteral
	}
	entry.IsLiteral = &isLiteral
	isRuntime := existing.IsRuntime
	if req.IsRuntime != nil {
		isRuntime = *req.IsRuntime
	}
	entry.IsRuntime = &isRuntime

	bulkReq := &models.ServiceEnvBulkUpdateRequest{
		Data: []models.ServiceEnvironmentVariableCreateRequest{entry},
	}
	err = s.client.Patch(ctx, fmt.Sprintf("services/%s/envs/bulk", serviceUUID), bulkReq, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update environment variable for service %s: %w", serviceUUID, err)
	}

	// Fetch the updated env var to return the current state
	updated, err := s.GetEnv(ctx, serviceUUID, entry.Key)
	if err != nil {
		return nil, fmt.Errorf("environment variable updated but failed to fetch result: %w", err)
	}
	return updated, nil
}

// DeleteEnv deletes an environment variable from a service
func (s *Service) DeleteEnv(ctx context.Context, serviceUUID, envUUID string) error {
	err := s.client.Delete(ctx, fmt.Sprintf("services/%s/envs/%s", serviceUUID, envUUID))
	if err != nil {
		return fmt.Errorf("failed to delete environment variable %s from service %s: %w", envUUID, serviceUUID, err)
	}
	return nil
}

// BulkUpdateEnvs updates multiple environment variables in a single request
func (s *Service) BulkUpdateEnvs(ctx context.Context, serviceUUID string, req *models.ServiceEnvBulkUpdateRequest) (*models.ServiceEnvBulkUpdateResponse, error) {
	var response models.ServiceEnvBulkUpdateResponse
	err := s.client.Patch(ctx, fmt.Sprintf("services/%s/envs/bulk", serviceUUID), req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update environment variables for service %s: %w", serviceUUID, err)
	}
	return &response, nil
}

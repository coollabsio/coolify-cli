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

// UpdateEnv updates an environment variable for a service
func (s *Service) UpdateEnv(ctx context.Context, serviceUUID string, req *models.ServiceEnvironmentVariableUpdateRequest) (*models.ServiceEnvironmentVariable, error) {
	var env models.ServiceEnvironmentVariable
	err := s.client.Patch(ctx, fmt.Sprintf("services/%s/envs", serviceUUID), req, &env)
	if err != nil {
		return nil, fmt.Errorf("failed to update environment variable for service %s: %w", serviceUUID, err)
	}
	return &env, nil
}

// DeleteEnv deletes an environment variable from a service
func (s *Service) DeleteEnv(ctx context.Context, serviceUUID, envUUID string) error {
	err := s.client.Delete(ctx, fmt.Sprintf("services/%s/envs/%s", serviceUUID, envUUID))
	if err != nil {
		return fmt.Errorf("failed to delete environment variable %s from service %s: %w", envUUID, serviceUUID, err)
	}
	return nil
}

// ListStorages retrieves all storages for a service
func (s *Service) ListStorages(ctx context.Context, uuid string) ([]models.StorageListItem, error) {
	var resp models.StoragesResponse
	err := s.client.Get(ctx, fmt.Sprintf("services/%s/storages", uuid), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list storages for service %s: %w", uuid, err)
	}
	return models.MergeStorages(resp), nil
}

// CreateStorage creates a new storage for a service
func (s *Service) CreateStorage(ctx context.Context, uuid string, req *models.ServiceStorageCreateRequest) error {
	err := s.client.Post(ctx, fmt.Sprintf("services/%s/storages", uuid), req, nil)
	if err != nil {
		return fmt.Errorf("failed to create storage for service %s: %w", uuid, err)
	}
	return nil
}

// UpdateStorage updates a storage for a service
func (s *Service) UpdateStorage(ctx context.Context, uuid string, req *models.StorageUpdateRequest) error {
	err := s.client.Patch(ctx, fmt.Sprintf("services/%s/storages", uuid), req, nil)
	if err != nil {
		return fmt.Errorf("failed to update storage for service %s: %w", uuid, err)
	}
	return nil
}

// DeleteStorage deletes a storage from a service
func (s *Service) DeleteStorage(ctx context.Context, svcUUID, storageUUID string) error {
	err := s.client.Delete(ctx, fmt.Sprintf("services/%s/storages/%s", svcUUID, storageUUID))
	if err != nil {
		return fmt.Errorf("failed to delete storage %s from service %s: %w", storageUUID, svcUUID, err)
	}
	return nil
}

// BulkUpdateEnvs updates multiple environment variables in a single request
func (s *Service) BulkUpdateEnvs(ctx context.Context, serviceUUID string, req *models.ServiceEnvBulkUpdateRequest) (models.ServiceEnvBulkUpdateResponse, error) {
	var response models.ServiceEnvBulkUpdateResponse
	err := s.client.Patch(ctx, fmt.Sprintf("services/%s/envs/bulk", serviceUUID), req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update environment variables for service %s: %w", serviceUUID, err)
	}
	return response, nil
}

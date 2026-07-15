package service

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

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

// Logs retrieves logs for one sub-resource in a service.
func (s *Service) Logs(ctx context.Context, uuid, subServiceName string, lines int, showTimestamps bool) (*models.LogsResponse, error) {
	query := url.Values{}
	query.Set("sub_service_name", subServiceName)
	query.Set("lines", strconv.Itoa(lines))
	query.Set("show_timestamps", strconv.FormatBool(showTimestamps))

	var response models.LogsResponse
	err := s.client.Get(ctx, fmt.Sprintf("services/%s/logs?%s", uuid, query.Encode()), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for service %s sub-resource %s: %w", uuid, subServiceName, err)
	}
	return &response, nil
}

// Move moves a service to another environment.
func (s *Service) Move(ctx context.Context, uuid, environmentUUID string) (*models.MoveResourceResponse, error) {
	var response models.MoveResourceResponse
	request := &models.MoveResourceRequest{EnvironmentUUID: environmentUUID}
	err := s.client.Post(ctx, fmt.Sprintf("services/%s/move", uuid), request, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to move service %s: %w", uuid, err)
	}
	return &response, nil
}

// ListApplications lists compose applications belonging to a service.
func (s *Service) ListApplications(ctx context.Context, serviceUUID string) ([]models.ServiceApplication, error) {
	var applications []models.ServiceApplication
	err := s.client.Get(ctx, fmt.Sprintf("services/%s/applications", serviceUUID), &applications)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications for service %s: %w", serviceUUID, err)
	}
	return applications, nil
}

// GetApplication retrieves one compose application belonging to a service.
func (s *Service) GetApplication(ctx context.Context, serviceUUID, applicationUUID string) (*models.ServiceApplication, error) {
	var application models.ServiceApplication
	err := s.client.Get(ctx, fmt.Sprintf("services/%s/applications/%s", serviceUUID, applicationUUID), &application)
	if err != nil {
		return nil, fmt.Errorf("failed to get application %s for service %s: %w", applicationUUID, serviceUUID, err)
	}
	return &application, nil
}

// UpdateApplication updates one compose application belonging to a service.
func (s *Service) UpdateApplication(ctx context.Context, serviceUUID, applicationUUID string, forceDomainOverride bool, request *models.ServiceApplicationUpdateRequest) (*models.ServiceApplication, error) {
	query := url.Values{}
	query.Set("force_domain_override", strconv.FormatBool(forceDomainOverride))
	var application models.ServiceApplication
	err := s.client.Patch(ctx, fmt.Sprintf("services/%s/applications/%s?%s", serviceUUID, applicationUUID, query.Encode()), request, &application)
	if err != nil {
		return nil, fmt.Errorf("failed to update application %s for service %s: %w", applicationUUID, serviceUUID, err)
	}
	return &application, nil
}

// ApplicationLogs retrieves logs for one compose application belonging to a service.
func (s *Service) ApplicationLogs(ctx context.Context, serviceUUID, applicationUUID string, lines int) (*models.LogsResponse, error) {
	query := url.Values{}
	query.Set("lines", strconv.Itoa(lines))
	var response models.LogsResponse
	err := s.client.Get(ctx, fmt.Sprintf("services/%s/applications/%s/logs?%s", serviceUUID, applicationUUID, query.Encode()), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for application %s in service %s: %w", applicationUUID, serviceUUID, err)
	}
	return &response, nil
}

// StartApplication deploys one compose application belonging to a service.
func (s *Service) StartApplication(ctx context.Context, serviceUUID, applicationUUID string, force, latest bool) (*models.ServiceLifecycleResponse, error) {
	query := url.Values{}
	query.Set("force", strconv.FormatBool(force))
	query.Set("latest", strconv.FormatBool(latest))
	return s.applicationLifecycle(ctx, serviceUUID, applicationUUID, "start", query)
}

// RestartApplication restarts one compose application belonging to a service.
func (s *Service) RestartApplication(ctx context.Context, serviceUUID, applicationUUID string) (*models.ServiceLifecycleResponse, error) {
	return s.applicationLifecycle(ctx, serviceUUID, applicationUUID, "restart", nil)
}

// StopApplication stops one compose application belonging to a service.
func (s *Service) StopApplication(ctx context.Context, serviceUUID, applicationUUID string) (*models.ServiceLifecycleResponse, error) {
	return s.applicationLifecycle(ctx, serviceUUID, applicationUUID, "stop", nil)
}

func (s *Service) applicationLifecycle(ctx context.Context, serviceUUID, applicationUUID, action string, query url.Values) (*models.ServiceLifecycleResponse, error) {
	path := fmt.Sprintf("services/%s/applications/%s/%s", serviceUUID, applicationUUID, action)
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	var response models.ServiceLifecycleResponse
	err := s.client.Post(ctx, path, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to %s application %s for service %s: %w", action, applicationUUID, serviceUUID, err)
	}
	return &response, nil
}

// ListDatabases lists compose databases belonging to a service.
func (s *Service) ListDatabases(ctx context.Context, serviceUUID string) ([]models.ServiceDatabase, error) {
	var databases []models.ServiceDatabase
	err := s.client.Get(ctx, fmt.Sprintf("services/%s/databases", serviceUUID), &databases)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases for service %s: %w", serviceUUID, err)
	}
	return databases, nil
}

// GetDatabase retrieves one compose database belonging to a service.
func (s *Service) GetDatabase(ctx context.Context, serviceUUID, databaseUUID string) (*models.ServiceDatabase, error) {
	var database models.ServiceDatabase
	err := s.client.Get(ctx, fmt.Sprintf("services/%s/databases/%s", serviceUUID, databaseUUID), &database)
	if err != nil {
		return nil, fmt.Errorf("failed to get database %s for service %s: %w", databaseUUID, serviceUUID, err)
	}
	return &database, nil
}

// UpdateDatabase updates one compose database belonging to a service.
func (s *Service) UpdateDatabase(ctx context.Context, serviceUUID, databaseUUID string, request *models.ServiceDatabaseUpdateRequest) (*models.ServiceDatabase, error) {
	var database models.ServiceDatabase
	err := s.client.Patch(ctx, fmt.Sprintf("services/%s/databases/%s", serviceUUID, databaseUUID), request, &database)
	if err != nil {
		return nil, fmt.Errorf("failed to update database %s for service %s: %w", databaseUUID, serviceUUID, err)
	}
	return &database, nil
}

// DatabaseLogs retrieves logs for one compose database belonging to a service.
func (s *Service) DatabaseLogs(ctx context.Context, serviceUUID, databaseUUID string, lines int) (*models.LogsResponse, error) {
	query := url.Values{}
	query.Set("lines", strconv.Itoa(lines))
	var response models.LogsResponse
	err := s.client.Get(ctx, fmt.Sprintf("services/%s/databases/%s/logs?%s", serviceUUID, databaseUUID, query.Encode()), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for database %s in service %s: %w", databaseUUID, serviceUUID, err)
	}
	return &response, nil
}

// StartDatabase deploys one compose database belonging to a service.
func (s *Service) StartDatabase(ctx context.Context, serviceUUID, databaseUUID string, force, latest bool) (*models.ServiceLifecycleResponse, error) {
	query := url.Values{}
	query.Set("force", strconv.FormatBool(force))
	query.Set("latest", strconv.FormatBool(latest))
	return s.databaseLifecycle(ctx, serviceUUID, databaseUUID, "start", query)
}

// RestartDatabase restarts one compose database belonging to a service.
func (s *Service) RestartDatabase(ctx context.Context, serviceUUID, databaseUUID string) (*models.ServiceLifecycleResponse, error) {
	return s.databaseLifecycle(ctx, serviceUUID, databaseUUID, "restart", nil)
}

// StopDatabase stops one compose database belonging to a service.
func (s *Service) StopDatabase(ctx context.Context, serviceUUID, databaseUUID string) (*models.ServiceLifecycleResponse, error) {
	return s.databaseLifecycle(ctx, serviceUUID, databaseUUID, "stop", nil)
}

func (s *Service) databaseLifecycle(ctx context.Context, serviceUUID, databaseUUID, action string, query url.Values) (*models.ServiceLifecycleResponse, error) {
	path := fmt.Sprintf("services/%s/databases/%s/%s", serviceUUID, databaseUUID, action)
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	var response models.ServiceLifecycleResponse
	err := s.client.Post(ctx, path, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to %s database %s for service %s: %w", action, databaseUUID, serviceUUID, err)
	}
	return &response, nil
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

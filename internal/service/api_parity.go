package service

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// --- S3 ---

type S3StorageService struct{ client *api.Client }

func NewS3StorageService(client *api.Client) *S3StorageService {
	return &S3StorageService{client: client}
}

func (s *S3StorageService) List(ctx context.Context) ([]models.S3Storage, error) {
	var items []models.S3Storage
	err := s.client.Get(ctx, "s3-storages", &items)
	return items, err
}

func (s *S3StorageService) Get(ctx context.Context, uuid string) (*models.S3Storage, error) {
	var item models.S3Storage
	err := s.client.Get(ctx, "s3-storages/"+url.PathEscape(uuid), &item)
	return &item, err
}

func (s *S3StorageService) Create(ctx context.Context, req models.S3StorageCreateRequest) (*models.UUID, error) {
	var resp models.UUID
	err := s.client.Post(ctx, "s3-storages", req, &resp)
	return &resp, err
}

func (s *S3StorageService) Update(ctx context.Context, uuid string, req models.S3StorageUpdateRequest) (*models.UUID, error) {
	var resp models.UUID
	err := s.client.Patch(ctx, "s3-storages/"+url.PathEscape(uuid), req, &resp)
	return &resp, err
}

func (s *S3StorageService) Delete(ctx context.Context, uuid string) error {
	return s.client.Delete(ctx, "s3-storages/"+url.PathEscape(uuid))
}

func (s *S3StorageService) Validate(ctx context.Context, uuid string) (*models.S3StorageValidation, error) {
	var resp models.S3StorageValidation
	err := s.client.Post(ctx, "s3-storages/"+url.PathEscape(uuid)+"/validate", map[string]any{}, &resp)
	return &resp, err
}

// --- Notifications ---

type NotificationService struct{ client *api.Client }

func NewNotificationService(client *api.Client) *NotificationService {
	return &NotificationService{client: client}
}

func (s *NotificationService) Get(ctx context.Context, channel string) (map[string]any, error) {
	var out map[string]any
	err := s.client.Get(ctx, "notifications/"+url.PathEscape(channel), &out)
	return out, err
}

func (s *NotificationService) Update(ctx context.Context, channel string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.client.Patch(ctx, "notifications/"+url.PathEscape(channel), body, &out)
	return out, err
}

// --- Shared envs ---

type SharedEnvService struct{ client *api.Client }

func NewSharedEnvService(client *api.Client) *SharedEnvService {
	return &SharedEnvService{client: client}
}

func (s *SharedEnvService) ListTeam(ctx context.Context) ([]models.SharedEnvironmentVariable, error) {
	var items []models.SharedEnvironmentVariable
	err := s.client.Get(ctx, "team/envs", &items)
	return items, err
}

func (s *SharedEnvService) CreateTeam(ctx context.Context, req models.SharedEnvCreateRequest) (*models.SharedEnvCreateResponse, error) {
	var resp models.SharedEnvCreateResponse
	err := s.client.Post(ctx, "team/envs", req, &resp)
	return &resp, err
}

func (s *SharedEnvService) UpdateTeam(ctx context.Context, id int, req models.SharedEnvUpdateRequest) (*models.SharedEnvironmentVariable, error) {
	var resp models.SharedEnvironmentVariable
	err := s.client.Patch(ctx, "team/envs/"+strconv.Itoa(id), req, &resp)
	return &resp, err
}

func (s *SharedEnvService) DeleteTeam(ctx context.Context, id int) error {
	return s.client.Delete(ctx, "team/envs/"+strconv.Itoa(id))
}

func (s *SharedEnvService) ListProject(ctx context.Context, projectUUID string) ([]models.SharedEnvironmentVariable, error) {
	var items []models.SharedEnvironmentVariable
	err := s.client.Get(ctx, "projects/"+url.PathEscape(projectUUID)+"/envs", &items)
	return items, err
}

func (s *SharedEnvService) CreateProject(ctx context.Context, projectUUID string, req models.SharedEnvCreateRequest) (*models.SharedEnvCreateResponse, error) {
	var resp models.SharedEnvCreateResponse
	err := s.client.Post(ctx, "projects/"+url.PathEscape(projectUUID)+"/envs", req, &resp)
	return &resp, err
}

func (s *SharedEnvService) UpdateProject(ctx context.Context, projectUUID string, id int, req models.SharedEnvUpdateRequest) (*models.SharedEnvironmentVariable, error) {
	var resp models.SharedEnvironmentVariable
	err := s.client.Patch(ctx, "projects/"+url.PathEscape(projectUUID)+"/envs/"+strconv.Itoa(id), req, &resp)
	return &resp, err
}

func (s *SharedEnvService) DeleteProject(ctx context.Context, projectUUID string, id int) error {
	return s.client.Delete(ctx, "projects/"+url.PathEscape(projectUUID)+"/envs/"+strconv.Itoa(id))
}

func (s *SharedEnvService) ListEnvironment(ctx context.Context, projectUUID, envNameOrUUID string) ([]models.SharedEnvironmentVariable, error) {
	var items []models.SharedEnvironmentVariable
	path := fmt.Sprintf("projects/%s/environments/%s/envs", url.PathEscape(projectUUID), url.PathEscape(envNameOrUUID))
	err := s.client.Get(ctx, path, &items)
	return items, err
}

func (s *SharedEnvService) CreateEnvironment(ctx context.Context, projectUUID, envNameOrUUID string, req models.SharedEnvCreateRequest) (*models.SharedEnvCreateResponse, error) {
	var resp models.SharedEnvCreateResponse
	path := fmt.Sprintf("projects/%s/environments/%s/envs", url.PathEscape(projectUUID), url.PathEscape(envNameOrUUID))
	err := s.client.Post(ctx, path, req, &resp)
	return &resp, err
}

func (s *SharedEnvService) UpdateEnvironment(ctx context.Context, projectUUID, envNameOrUUID string, id int, req models.SharedEnvUpdateRequest) (*models.SharedEnvironmentVariable, error) {
	var resp models.SharedEnvironmentVariable
	path := fmt.Sprintf("projects/%s/environments/%s/envs/%d", url.PathEscape(projectUUID), url.PathEscape(envNameOrUUID), id)
	err := s.client.Patch(ctx, path, req, &resp)
	return &resp, err
}

func (s *SharedEnvService) DeleteEnvironment(ctx context.Context, projectUUID, envNameOrUUID string, id int) error {
	path := fmt.Sprintf("projects/%s/environments/%s/envs/%d", url.PathEscape(projectUUID), url.PathEscape(envNameOrUUID), id)
	return s.client.Delete(ctx, path)
}

func (s *SharedEnvService) ListServer(ctx context.Context, serverUUID string) ([]models.SharedEnvironmentVariable, error) {
	var items []models.SharedEnvironmentVariable
	err := s.client.Get(ctx, "servers/"+url.PathEscape(serverUUID)+"/envs", &items)
	return items, err
}

func (s *SharedEnvService) CreateServer(ctx context.Context, serverUUID string, req models.SharedEnvCreateRequest) (*models.SharedEnvCreateResponse, error) {
	var resp models.SharedEnvCreateResponse
	err := s.client.Post(ctx, "servers/"+url.PathEscape(serverUUID)+"/envs", req, &resp)
	return &resp, err
}

func (s *SharedEnvService) UpdateServer(ctx context.Context, serverUUID string, id int, req models.SharedEnvUpdateRequest) (*models.SharedEnvironmentVariable, error) {
	var resp models.SharedEnvironmentVariable
	err := s.client.Patch(ctx, "servers/"+url.PathEscape(serverUUID)+"/envs/"+strconv.Itoa(id), req, &resp)
	return &resp, err
}

func (s *SharedEnvService) DeleteServer(ctx context.Context, serverUUID string, id int) error {
	return s.client.Delete(ctx, "servers/"+url.PathEscape(serverUUID)+"/envs/"+strconv.Itoa(id))
}

// --- Cloud-init ---

type CloudInitService struct{ client *api.Client }

func NewCloudInitService(client *api.Client) *CloudInitService {
	return &CloudInitService{client: client}
}

func (s *CloudInitService) List(ctx context.Context) ([]models.CloudInitScript, error) {
	var items []models.CloudInitScript
	err := s.client.Get(ctx, "cloud-init-scripts", &items)
	return items, err
}

func (s *CloudInitService) Get(ctx context.Context, uuid string) (*models.CloudInitScript, error) {
	var item models.CloudInitScript
	err := s.client.Get(ctx, "cloud-init-scripts/"+url.PathEscape(uuid), &item)
	return &item, err
}

func (s *CloudInitService) Create(ctx context.Context, req models.CloudInitScriptCreateRequest) (*models.UUID, error) {
	var resp models.UUID
	err := s.client.Post(ctx, "cloud-init-scripts", req, &resp)
	return &resp, err
}

func (s *CloudInitService) Update(ctx context.Context, uuid string, req models.CloudInitScriptUpdateRequest) (*models.UUID, error) {
	var resp models.UUID
	err := s.client.Patch(ctx, "cloud-init-scripts/"+url.PathEscape(uuid), req, &resp)
	return &resp, err
}

func (s *CloudInitService) Delete(ctx context.Context, uuid string) error {
	return s.client.Delete(ctx, "cloud-init-scripts/"+url.PathEscape(uuid))
}

// --- Server subsystems ---

type ServerSubsystemService struct{ client *api.Client }

func NewServerSubsystemService(client *api.Client) *ServerSubsystemService {
	return &ServerSubsystemService{client: client}
}

func (s *ServerSubsystemService) GetDockerCleanup(ctx context.Context, serverUUID string) (*models.DockerCleanupSettings, error) {
	var out models.DockerCleanupSettings
	err := s.client.Get(ctx, "servers/"+url.PathEscape(serverUUID)+"/docker-cleanup", &out)
	return &out, err
}

func (s *ServerSubsystemService) UpdateDockerCleanup(ctx context.Context, serverUUID string, req models.DockerCleanupUpdateRequest) (*models.DockerCleanupSettings, error) {
	var out models.DockerCleanupSettings
	err := s.client.Patch(ctx, "servers/"+url.PathEscape(serverUUID)+"/docker-cleanup", req, &out)
	return &out, err
}

func (s *ServerSubsystemService) RunDockerCleanup(ctx context.Context, serverUUID string) (map[string]any, error) {
	var out map[string]any
	err := s.client.Post(ctx, "servers/"+url.PathEscape(serverUUID)+"/docker-cleanup/run", map[string]any{}, &out)
	return out, err
}

func (s *ServerSubsystemService) ListDockerCleanupExecutions(ctx context.Context, serverUUID string) ([]map[string]any, error) {
	var out []map[string]any
	err := s.client.Get(ctx, "servers/"+url.PathEscape(serverUUID)+"/docker-cleanup/executions", &out)
	return out, err
}

func (s *ServerSubsystemService) GetLogDrains(ctx context.Context, serverUUID string) (map[string]any, error) {
	var out map[string]any
	err := s.client.Get(ctx, "servers/"+url.PathEscape(serverUUID)+"/log-drains", &out)
	return out, err
}

func (s *ServerSubsystemService) UpdateLogDrains(ctx context.Context, serverUUID string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.client.Patch(ctx, "servers/"+url.PathEscape(serverUUID)+"/log-drains", body, &out)
	return out, err
}

func (s *ServerSubsystemService) GetSentinel(ctx context.Context, serverUUID string) (map[string]any, error) {
	var out map[string]any
	err := s.client.Get(ctx, "servers/"+url.PathEscape(serverUUID)+"/sentinel", &out)
	return out, err
}

func (s *ServerSubsystemService) UpdateSentinel(ctx context.Context, serverUUID string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	err := s.client.Patch(ctx, "servers/"+url.PathEscape(serverUUID)+"/sentinel", body, &out)
	return out, err
}

func (s *ServerSubsystemService) GetCloudflareTunnel(ctx context.Context, serverUUID string) (map[string]any, error) {
	var out map[string]any
	err := s.client.Get(ctx, "servers/"+url.PathEscape(serverUUID)+"/cloudflare-tunnel", &out)
	return out, err
}

func (s *ServerSubsystemService) EnableCloudflareTunnel(ctx context.Context, serverUUID string) (map[string]any, error) {
	var out map[string]any
	err := s.client.Post(ctx, "servers/"+url.PathEscape(serverUUID)+"/cloudflare-tunnel/enable", map[string]any{}, &out)
	return out, err
}

func (s *ServerSubsystemService) DisableCloudflareTunnel(ctx context.Context, serverUUID string) (map[string]any, error) {
	var out map[string]any
	err := s.client.Post(ctx, "servers/"+url.PathEscape(serverUUID)+"/cloudflare-tunnel/disable", map[string]any{}, &out)
	return out, err
}

// UpdateCloudflareTunnel PATCHes the is_cloudflare_tunnel setting.
func (s *ServerSubsystemService) UpdateCloudflareTunnel(ctx context.Context, serverUUID string, enabled bool) (map[string]any, error) {
	var out map[string]any
	err := s.client.Patch(ctx, "servers/"+url.PathEscape(serverUUID)+"/cloudflare-tunnel", map[string]any{
		"is_cloudflare_tunnel": enabled,
	}, &out)
	return out, err
}

func (s *ServerSubsystemService) GetProxy(ctx context.Context, serverUUID string) (*models.ServerProxySettings, error) {
	var out models.ServerProxySettings
	err := s.client.Get(ctx, "servers/"+url.PathEscape(serverUUID)+"/proxy", &out)
	return &out, err
}

func (s *ServerSubsystemService) UpdateProxy(ctx context.Context, serverUUID string, req models.ServerProxyUpdateRequest) (*models.ServerProxySettings, error) {
	var out models.ServerProxySettings
	err := s.client.Patch(ctx, "servers/"+url.PathEscape(serverUUID)+"/proxy", req, &out)
	return &out, err
}

func (s *ServerSubsystemService) SaveProxyConfiguration(ctx context.Context, serverUUID string, configuration string) (map[string]any, error) {
	var out map[string]any
	err := s.client.Put(ctx, "servers/"+url.PathEscape(serverUUID)+"/proxy/configuration", models.ProxyConfigurationRequest{Configuration: configuration}, &out)
	return out, err
}

func (s *ServerSubsystemService) RestartProxy(ctx context.Context, serverUUID string) (map[string]any, error) {
	var out map[string]any
	err := s.client.Post(ctx, "servers/"+url.PathEscape(serverUUID)+"/proxy/restart", map[string]any{}, &out)
	return out, err
}

// --- Tag team-level CRUD extensions ---

func (s *TagService) Create(ctx context.Context, name string) (*models.Tag, error) {
	var tag models.Tag
	err := s.client.Post(ctx, "tags", models.TagCreateRequest{Name: name}, &tag)
	return &tag, err
}

func (s *TagService) Update(ctx context.Context, uuid, name string) (*models.Tag, error) {
	var tag models.Tag
	err := s.client.Patch(ctx, "tags/"+url.PathEscape(uuid), models.TagUpdateRequest{Name: name}, &tag)
	return &tag, err
}

func (s *TagService) Delete(ctx context.Context, uuid string) error {
	return s.client.Delete(ctx, "tags/"+url.PathEscape(uuid))
}

// --- Destination update ---

func (s *DestinationService) Update(ctx context.Context, uuid string, req models.DestinationUpdateRequest) (*models.Destination, error) {
	var dest models.Destination
	err := s.client.Patch(ctx, "destinations/"+url.PathEscape(uuid), req, &dest)
	return &dest, err
}

// --- Application lifecycle extensions ---

func (s *ApplicationService) Clone(ctx context.Context, uuid string, req models.ResourceCloneRequest) (*models.ResourceCloneResponse, error) {
	var resp models.ResourceCloneResponse
	err := s.client.Post(ctx, "applications/"+url.PathEscape(uuid)+"/clone", req, &resp)
	return &resp, err
}

func (s *ApplicationService) Rollback(ctx context.Context, uuid, commit string) (map[string]any, error) {
	var resp map[string]any
	err := s.client.Post(ctx, "applications/"+url.PathEscape(uuid)+"/rollback", models.RollbackRequest{Commit: commit}, &resp)
	return resp, err
}

func (s *ApplicationService) ListRollbackImages(ctx context.Context, uuid string) (map[string]any, error) {
	var resp map[string]any
	err := s.client.Get(ctx, "applications/"+url.PathEscape(uuid)+"/rollback-images", &resp)
	return resp, err
}

func (s *ApplicationService) ListDestinations(ctx context.Context, uuid string) ([]map[string]any, error) {
	var resp []map[string]any
	err := s.client.Get(ctx, "applications/"+url.PathEscape(uuid)+"/destinations", &resp)
	return resp, err
}

func (s *ApplicationService) AddDestination(ctx context.Context, uuid, destinationUUID string) (map[string]any, error) {
	var resp map[string]any
	err := s.client.Post(ctx, "applications/"+url.PathEscape(uuid)+"/destinations", models.AppDestinationAttachRequest{DestinationUUID: destinationUUID}, &resp)
	return resp, err
}

func (s *ApplicationService) RemoveDestination(ctx context.Context, uuid, destinationUUID string) error {
	return s.client.Delete(ctx, "applications/"+url.PathEscape(uuid)+"/destinations/"+url.PathEscape(destinationUUID))
}

func (s *ApplicationService) RunStorageBackup(ctx context.Context, appUUID, storageUUID string) (map[string]any, error) {
	var resp map[string]any
	path := fmt.Sprintf("applications/%s/storages/%s/backups/run", url.PathEscape(appUUID), url.PathEscape(storageUUID))
	err := s.client.Post(ctx, path, map[string]any{}, &resp)
	return resp, err
}

func (s *ApplicationService) SetStorageBackupSchedule(ctx context.Context, appUUID, storageUUID string, req models.VolumeBackupScheduleRequest) (map[string]any, error) {
	var resp map[string]any
	path := fmt.Sprintf("applications/%s/storages/%s/backups", url.PathEscape(appUUID), url.PathEscape(storageUUID))
	err := s.client.Put(ctx, path, req, &resp)
	return resp, err
}

func (s *ApplicationService) DeleteStorageBackupSchedule(ctx context.Context, appUUID, storageUUID string) error {
	path := fmt.Sprintf("applications/%s/storages/%s/backups", url.PathEscape(appUUID), url.PathEscape(storageUUID))
	return s.client.Delete(ctx, path)
}

func (s *DatabaseService) Clone(ctx context.Context, uuid string, req models.ResourceCloneRequest) (*models.ResourceCloneResponse, error) {
	var resp models.ResourceCloneResponse
	err := s.client.Post(ctx, "databases/"+url.PathEscape(uuid)+"/clone", req, &resp)
	return &resp, err
}

func (s *DatabaseService) RunStorageBackup(ctx context.Context, dbUUID, storageUUID string) (map[string]any, error) {
	var resp map[string]any
	path := fmt.Sprintf("databases/%s/storages/%s/backups/run", url.PathEscape(dbUUID), url.PathEscape(storageUUID))
	err := s.client.Post(ctx, path, map[string]any{}, &resp)
	return resp, err
}

func (s *DatabaseService) SetStorageBackupSchedule(ctx context.Context, dbUUID, storageUUID string, req models.VolumeBackupScheduleRequest) (map[string]any, error) {
	var resp map[string]any
	path := fmt.Sprintf("databases/%s/storages/%s/backups", url.PathEscape(dbUUID), url.PathEscape(storageUUID))
	err := s.client.Put(ctx, path, req, &resp)
	return resp, err
}

func (s *DatabaseService) DeleteStorageBackupSchedule(ctx context.Context, dbUUID, storageUUID string) error {
	path := fmt.Sprintf("databases/%s/storages/%s/backups", url.PathEscape(dbUUID), url.PathEscape(storageUUID))
	return s.client.Delete(ctx, path)
}

func (s *Service) Clone(ctx context.Context, uuid string, req models.ResourceCloneRequest) (*models.ResourceCloneResponse, error) {
	var resp models.ResourceCloneResponse
	err := s.client.Post(ctx, "services/"+url.PathEscape(uuid)+"/clone", req, &resp)
	return &resp, err
}

func (s *Service) RunStorageBackup(ctx context.Context, serviceUUID, storageUUID string) (map[string]any, error) {
	var resp map[string]any
	path := fmt.Sprintf("services/%s/storages/%s/backups/run", url.PathEscape(serviceUUID), url.PathEscape(storageUUID))
	err := s.client.Post(ctx, path, map[string]any{}, &resp)
	return resp, err
}

func (s *Service) SetStorageBackupSchedule(ctx context.Context, serviceUUID, storageUUID string, req models.VolumeBackupScheduleRequest) (map[string]any, error) {
	var resp map[string]any
	path := fmt.Sprintf("services/%s/storages/%s/backups", url.PathEscape(serviceUUID), url.PathEscape(storageUUID))
	err := s.client.Put(ctx, path, req, &resp)
	return resp, err
}

func (s *Service) DeleteStorageBackupSchedule(ctx context.Context, serviceUUID, storageUUID string) error {
	path := fmt.Sprintf("services/%s/storages/%s/backups", url.PathEscape(serviceUUID), url.PathEscape(storageUUID))
	return s.client.Delete(ctx, path)
}

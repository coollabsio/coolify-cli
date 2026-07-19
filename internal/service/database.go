package service

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// DatabaseService handles database-related operations
type DatabaseService struct {
	client *api.Client
}

// NewDatabaseService creates a new database service
func NewDatabaseService(client *api.Client) *DatabaseService {
	return &DatabaseService{client: client}
}

// List retrieves all databases
func (s *DatabaseService) List(ctx context.Context) ([]models.Database, error) {
	var databases []models.Database
	err := s.client.Get(ctx, "databases", &databases)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	// Infer database type if not provided by API
	for i := range databases {
		if databases[i].Type == "" {
			databases[i].Type = inferDatabaseType(&databases[i])
		}
	}

	return databases, nil
}

// Get retrieves a database by UUID
func (s *DatabaseService) Get(ctx context.Context, uuid string) (*models.Database, error) {
	var database models.Database
	err := s.client.Get(ctx, fmt.Sprintf("databases/%s", uuid), &database)
	if err != nil {
		return nil, fmt.Errorf("failed to get database %s: %w", uuid, err)
	}

	// Infer database type if not provided by API
	if database.Type == "" {
		database.Type = inferDatabaseType(&database)
	}

	return &database, nil
}

// Create creates a new database of the specified type
func (s *DatabaseService) Create(ctx context.Context, dbType string, req *models.DatabaseCreateRequest) (*models.Database, error) {
	var database models.Database
	err := s.client.Post(ctx, fmt.Sprintf("databases/%s", dbType), req, &database)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s database: %w", dbType, err)
	}
	return &database, nil
}

// Update updates a database
func (s *DatabaseService) Update(ctx context.Context, uuid string, req *models.DatabaseUpdateRequest) error {
	err := s.client.Patch(ctx, fmt.Sprintf("databases/%s", uuid), req, nil)
	if err != nil {
		return fmt.Errorf("failed to update database %s: %w", uuid, err)
	}
	return nil
}

// Delete deletes a database
func (s *DatabaseService) Delete(ctx context.Context, uuid string, deleteConfigurations, deleteVolumes, dockerCleanup, deleteConnectedNetworks bool) error {
	url := fmt.Sprintf("databases/%s?delete_configurations=%t&delete_volumes=%t&docker_cleanup=%t&delete_connected_networks=%t",
		uuid, deleteConfigurations, deleteVolumes, dockerCleanup, deleteConnectedNetworks)

	err := s.client.Delete(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to delete database %s: %w", uuid, err)
	}
	return nil
}

// Start starts a database
func (s *DatabaseService) Start(ctx context.Context, uuid string) (*models.DatabaseLifecycleResponse, error) {
	var response models.DatabaseLifecycleResponse
	err := s.client.Post(ctx, fmt.Sprintf("databases/%s/start", uuid), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to start database %s: %w", uuid, err)
	}
	return &response, nil
}

// Stop stops a database
func (s *DatabaseService) Stop(ctx context.Context, uuid string) (*models.DatabaseLifecycleResponse, error) {
	var response models.DatabaseLifecycleResponse
	err := s.client.Post(ctx, fmt.Sprintf("databases/%s/stop", uuid), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to stop database %s: %w", uuid, err)
	}
	return &response, nil
}

// Restart restarts a database
func (s *DatabaseService) Restart(ctx context.Context, uuid string) (*models.DatabaseLifecycleResponse, error) {
	var response models.DatabaseLifecycleResponse
	err := s.client.Post(ctx, fmt.Sprintf("databases/%s/restart", uuid), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to restart database %s: %w", uuid, err)
	}
	return &response, nil
}

// Logs retrieves container logs for a database.
func (s *DatabaseService) Logs(ctx context.Context, uuid string, lines int, showTimestamps bool) (*models.LogsResponse, error) {
	query := url.Values{}
	query.Set("lines", strconv.Itoa(lines))
	query.Set("show_timestamps", strconv.FormatBool(showTimestamps))

	var response models.LogsResponse
	err := s.client.Get(ctx, fmt.Sprintf("databases/%s/logs?%s", uuid, query.Encode()), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for database %s: %w", uuid, err)
	}
	return &response, nil
}

// Move moves a database to another environment.
func (s *DatabaseService) Move(ctx context.Context, uuid, environmentUUID string) (*models.MoveResourceResponse, error) {
	var response models.MoveResourceResponse
	request := &models.MoveResourceRequest{EnvironmentUUID: environmentUUID}
	err := s.client.Post(ctx, fmt.Sprintf("databases/%s/move", uuid), request, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to move database %s: %w", uuid, err)
	}
	return &response, nil
}

// ListBackups retrieves all backup configurations for a database
func (s *DatabaseService) ListBackups(ctx context.Context, uuid string) ([]models.DatabaseBackup, error) {
	var backups []models.DatabaseBackup
	err := s.client.Get(ctx, fmt.Sprintf("databases/%s/backups", uuid), &backups)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups for database %s: %w", uuid, err)
	}
	return backups, nil
}

// CreateBackup creates a new scheduled backup configuration
// Note: This endpoint will be available in a future version of Coolify
func (s *DatabaseService) CreateBackup(ctx context.Context, uuid string, req *models.DatabaseBackupCreateRequest) (*models.DatabaseBackup, error) {
	var backup models.DatabaseBackup
	err := s.client.Post(ctx, fmt.Sprintf("databases/%s/backups", uuid), req, &backup)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup for database %s: %w", uuid, err)
	}
	return &backup, nil
}

// UpdateBackup updates a backup configuration
func (s *DatabaseService) UpdateBackup(ctx context.Context, dbUUID, backupUUID string, req *models.DatabaseBackupUpdateRequest) error {
	err := s.client.Patch(ctx, fmt.Sprintf("databases/%s/backups/%s", dbUUID, backupUUID), req, nil)
	if err != nil {
		return fmt.Errorf("failed to update backup %s for database %s: %w", backupUUID, dbUUID, err)
	}
	return nil
}

// DeleteBackup deletes a backup configuration
func (s *DatabaseService) DeleteBackup(ctx context.Context, dbUUID, backupUUID string, deleteS3 bool) error {
	url := fmt.Sprintf("databases/%s/backups/%s?delete_s3=%t", dbUUID, backupUUID, deleteS3)
	err := s.client.Delete(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to delete backup %s for database %s: %w", backupUUID, dbUUID, err)
	}
	return nil
}

// ListBackupExecutions retrieves all executions for a backup configuration
func (s *DatabaseService) ListBackupExecutions(ctx context.Context, dbUUID, backupUUID string) ([]models.DatabaseBackupExecution, error) {
	var response models.DatabaseBackupExecutionsResponse
	err := s.client.Get(ctx, fmt.Sprintf("databases/%s/backups/%s/executions", dbUUID, backupUUID), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup executions for backup %s: %w", backupUUID, err)
	}
	return response.Executions, nil
}

// DeleteBackupExecution deletes a specific backup execution
func (s *DatabaseService) DeleteBackupExecution(ctx context.Context, dbUUID, backupUUID, executionUUID string, deleteS3 bool) error {
	url := fmt.Sprintf("databases/%s/backups/%s/executions/%s?delete_s3=%t", dbUUID, backupUUID, executionUUID, deleteS3)
	err := s.client.Delete(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to delete backup execution %s: %w", executionUUID, err)
	}
	return nil
}

// ListEnvs retrieves all environment variables for a database
func (s *DatabaseService) ListEnvs(ctx context.Context, uuid string) ([]models.DatabaseEnvironmentVariable, error) {
	var envs []models.DatabaseEnvironmentVariable
	err := s.client.Get(ctx, fmt.Sprintf("databases/%s/envs", uuid), &envs)
	if err != nil {
		return nil, fmt.Errorf("failed to list environment variables for database %s: %w", uuid, err)
	}
	return envs, nil
}

// GetEnv retrieves a single environment variable by UUID or key
func (s *DatabaseService) GetEnv(ctx context.Context, dbUUID, envIdentifier string) (*models.DatabaseEnvironmentVariable, error) {
	envs, err := s.ListEnvs(ctx, dbUUID)
	if err != nil {
		return nil, err
	}

	for _, env := range envs {
		if env.UUID == envIdentifier || env.Key == envIdentifier {
			return &env, nil
		}
	}

	return nil, fmt.Errorf("environment variable '%s' not found in database %s", envIdentifier, dbUUID)
}

// CreateEnv creates a new environment variable for a database
func (s *DatabaseService) CreateEnv(ctx context.Context, uuid string, req *models.DatabaseEnvironmentVariableCreateRequest) (*models.DatabaseEnvironmentVariable, error) {
	var env models.DatabaseEnvironmentVariable
	err := s.client.Post(ctx, fmt.Sprintf("databases/%s/envs", uuid), req, &env)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment variable for database %s: %w", uuid, err)
	}
	return &env, nil
}

// UpdateEnv updates an environment variable for a database
func (s *DatabaseService) UpdateEnv(ctx context.Context, dbUUID string, req *models.DatabaseEnvironmentVariableUpdateRequest) (*models.DatabaseEnvironmentVariable, error) {
	var env models.DatabaseEnvironmentVariable
	err := s.client.Patch(ctx, fmt.Sprintf("databases/%s/envs", dbUUID), req, &env)
	if err != nil {
		return nil, fmt.Errorf("failed to update environment variable for database %s: %w", dbUUID, err)
	}
	return &env, nil
}

// DeleteEnv deletes an environment variable from a database
func (s *DatabaseService) DeleteEnv(ctx context.Context, dbUUID, envUUID string) error {
	err := s.client.Delete(ctx, fmt.Sprintf("databases/%s/envs/%s", dbUUID, envUUID))
	if err != nil {
		return fmt.Errorf("failed to delete environment variable %s from database %s: %w", envUUID, dbUUID, err)
	}
	return nil
}

// BulkUpdateEnvs updates multiple environment variables for a database in a single request
func (s *DatabaseService) BulkUpdateEnvs(ctx context.Context, dbUUID string, req *models.DatabaseEnvBulkUpdateRequest) (models.DatabaseEnvBulkUpdateResponse, error) {
	var response models.DatabaseEnvBulkUpdateResponse
	err := s.client.Patch(ctx, fmt.Sprintf("databases/%s/envs/bulk", dbUUID), req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update environment variables for database %s: %w", dbUUID, err)
	}
	return response, nil
}

// ListStorages retrieves all storages for a database
func (s *DatabaseService) ListStorages(ctx context.Context, uuid string) ([]models.StorageListItem, error) {
	var resp models.StoragesResponse
	err := s.client.Get(ctx, fmt.Sprintf("databases/%s/storages", uuid), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list storages for database %s: %w", uuid, err)
	}
	return models.MergeStorages(resp), nil
}

// CreateStorage creates a new storage for a database
func (s *DatabaseService) CreateStorage(ctx context.Context, uuid string, req *models.StorageCreateRequest) error {
	err := s.client.Post(ctx, fmt.Sprintf("databases/%s/storages", uuid), req, nil)
	if err != nil {
		return fmt.Errorf("failed to create storage for database %s: %w", uuid, err)
	}
	return nil
}

// UpdateStorage updates a storage for a database
func (s *DatabaseService) UpdateStorage(ctx context.Context, uuid string, req *models.StorageUpdateRequest) error {
	err := s.client.Patch(ctx, fmt.Sprintf("databases/%s/storages", uuid), req, nil)
	if err != nil {
		return fmt.Errorf("failed to update storage for database %s: %w", uuid, err)
	}
	return nil
}

// DeleteStorage deletes a storage from a database
func (s *DatabaseService) DeleteStorage(ctx context.Context, dbUUID, storageUUID string) error {
	err := s.client.Delete(ctx, fmt.Sprintf("databases/%s/storages/%s", dbUUID, storageUUID))
	if err != nil {
		return fmt.Errorf("failed to delete storage %s from database %s: %w", storageUUID, dbUUID, err)
	}
	return nil
}

// inferDatabaseType determines the database type from available fields
func inferDatabaseType(db *models.Database) string {
	// Check for PostgreSQL
	if db.PostgresUser != nil || db.PostgresPassword != nil || db.PostgresDB != nil {
		return "postgresql"
	}

	// Check for MySQL
	if db.MysqlUser != nil || db.MysqlPassword != nil || db.MysqlDatabase != nil {
		return "mysql"
	}

	// Check for MariaDB
	if db.MariadbUser != nil || db.MariadbPassword != nil || db.MariadbDatabase != nil {
		return "mariadb"
	}

	// Check for MongoDB
	if db.MongoInitdbRootUsername != nil || db.MongoInitdbRootPassword != nil || db.MongoInitdbDatabase != nil {
		return "mongodb"
	}

	// Check for Redis
	if db.RedisPassword != nil || db.RedisConf != nil {
		return "redis"
	}

	// Check for KeyDB
	if db.KeydbPassword != nil || db.KeydbConf != nil {
		return "keydb"
	}

	// Check for Clickhouse
	if db.ClickhouseAdminUser != nil || db.ClickhouseAdminPassword != nil {
		return "clickhouse"
	}

	// Check for Dragonfly
	if db.DragonflyPassword != nil {
		return "dragonfly"
	}

	// Fallback: try to infer from image name
	if db.Image != nil {
		image := *db.Image
		if len(image) > 0 {
			// Extract base image name (e.g., "postgres:16-alpine" -> "postgres")
			for i := 0; i < len(image); i++ {
				if image[i] == ':' {
					return image[:i]
				}
			}
			return image
		}
	}

	return ""
}

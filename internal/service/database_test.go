package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestDatabaseService_List(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful list",
			serverResponse: `[
				{
					"id": 1,
					"uuid": "db-uuid-1",
					"name": "Production PostgreSQL",
					"description": "Main database",
					"status": "running",
					"type": "postgresql",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			]`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:           "empty list",
			serverResponse: `[]`,
			statusCode:     http.StatusOK,
			wantErr:        false,
			wantCount:      0,
		},
		{
			name:           "server error",
			serverResponse: `{"error":"internal server error"}`,
			statusCode:     http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			databases, err := dbService.List(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, databases, tt.wantCount)
		})
	}
}

func TestDatabaseService_Get(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		serverResponse string
		statusCode     int
		wantErr        bool
		wantName       string
	}{
		{
			name: "successful get",
			uuid: "db-uuid-1",
			serverResponse: `{
				"id": 1,
				"uuid": "db-uuid-1",
				"name": "Production PostgreSQL",
				"description": "Main database",
				"status": "running",
				"type": "postgresql",
				"postgres_db": "myapp",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantName:   "Production PostgreSQL",
		},
		{
			name:           "not found",
			uuid:           "nonexistent",
			serverResponse: `{"error":"not found"}`,
			statusCode:     http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.uuid, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			database, err := dbService.Get(context.Background(), tt.uuid)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantName, database.Name)
			assert.Equal(t, tt.uuid, database.UUID)
		})
	}
}

func TestDatabaseService_Create(t *testing.T) {
	tests := []struct {
		name           string
		dbType         string
		request        *models.DatabaseCreateRequest
		serverResponse string
		statusCode     int
		wantErr        bool
		wantUUID       string
	}{
		{
			name:   "create postgresql",
			dbType: "postgresql",
			request: &models.DatabaseCreateRequest{
				ServerUUID:  "server-uuid-1",
				ProjectUUID: "project-uuid-1",
			},
			serverResponse: `{
				"id": 1,
				"uuid": "db-uuid-new",
				"name": "New Database",
				"status": "starting",
				"type": "postgresql",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-01T00:00:00Z"
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantUUID:   "db-uuid-new",
		},
		{
			name:   "validation error",
			dbType: "mysql",
			request: &models.DatabaseCreateRequest{
				ServerUUID: "server-uuid-1",
			},
			serverResponse: `{"error":"project_uuid is required"}`,
			statusCode:     http.StatusUnprocessableEntity,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.dbType, r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				var req models.DatabaseCreateRequest
				_ = json.NewDecoder(r.Body).Decode(&req)
				assert.Equal(t, tt.request.ServerUUID, req.ServerUUID)

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			database, err := dbService.Create(context.Background(), tt.dbType, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUUID, database.UUID)
		})
	}
}

func TestDatabaseService_Update(t *testing.T) {
	tests := []struct {
		name       string
		uuid       string
		request    *models.DatabaseUpdateRequest
		statusCode int
		wantErr    bool
	}{
		{
			name: "successful update",
			uuid: "db-uuid-1",
			request: &models.DatabaseUpdateRequest{
				Name: stringPtr("Updated Name"),
			},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "not found",
			uuid: "nonexistent",
			request: &models.DatabaseUpdateRequest{
				Name: stringPtr("Updated Name"),
			},
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.uuid, r.URL.Path)
				assert.Equal(t, http.MethodPatch, r.Method)

				var req models.DatabaseUpdateRequest
				_ = json.NewDecoder(r.Body).Decode(&req)

				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusNotFound {
					_, _ = w.Write([]byte(`{"error":"not found"}`))
				}
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			err := dbService.Update(context.Background(), tt.uuid, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDatabaseService_Delete(t *testing.T) {
	tests := []struct {
		name                    string
		uuid                    string
		deleteConfigurations    bool
		deleteVolumes           bool
		dockerCleanup           bool
		deleteConnectedNetworks bool
		statusCode              int
		wantErr                 bool
		expectedQueryString     string
	}{
		{
			name:                    "successful delete with all cleanup",
			uuid:                    "db-uuid-1",
			deleteConfigurations:    true,
			deleteVolumes:           true,
			dockerCleanup:           true,
			deleteConnectedNetworks: true,
			statusCode:              http.StatusOK,
			wantErr:                 false,
			expectedQueryString:     "delete_configurations=true&delete_volumes=true&docker_cleanup=true&delete_connected_networks=true",
		},
		{
			name:                    "successful delete without cleanup",
			uuid:                    "db-uuid-2",
			deleteConfigurations:    false,
			deleteVolumes:           false,
			dockerCleanup:           false,
			deleteConnectedNetworks: false,
			statusCode:              http.StatusOK,
			wantErr:                 false,
			expectedQueryString:     "delete_configurations=false&delete_volumes=false&docker_cleanup=false&delete_connected_networks=false",
		},
		{
			name:                    "not found",
			uuid:                    "nonexistent",
			deleteConfigurations:    true,
			deleteVolumes:           true,
			dockerCleanup:           true,
			deleteConnectedNetworks: true,
			statusCode:              http.StatusNotFound,
			wantErr:                 true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.uuid, r.URL.Path)
				assert.Equal(t, http.MethodDelete, r.Method)
				if tt.expectedQueryString != "" {
					assert.Equal(t, tt.expectedQueryString, r.URL.RawQuery)
				}

				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_, _ = w.Write([]byte(`{"message":"Database deleted"}`))
				} else {
					_, _ = w.Write([]byte(`{"error":"not found"}`))
				}
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			err := dbService.Delete(context.Background(), tt.uuid, tt.deleteConfigurations, tt.deleteVolumes, tt.dockerCleanup, tt.deleteConnectedNetworks)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDatabaseService_Start(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		serverResponse string
		statusCode     int
		wantErr        bool
		wantMessage    string
	}{
		{
			name:           "successful start",
			uuid:           "db-uuid-1",
			serverResponse: `{"message":"Database started successfully"}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
			wantMessage:    "Database started successfully",
		},
		{
			name:           "not found",
			uuid:           "nonexistent",
			serverResponse: `{"error":"not found"}`,
			statusCode:     http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.uuid+"/start", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			response, err := dbService.Start(context.Background(), tt.uuid)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMessage, response.Message)
		})
	}
}

func TestDatabaseService_Stop(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		serverResponse string
		statusCode     int
		wantErr        bool
		wantMessage    string
	}{
		{
			name:           "successful stop",
			uuid:           "db-uuid-1",
			serverResponse: `{"message":"Database stopped successfully"}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
			wantMessage:    "Database stopped successfully",
		},
		{
			name:           "not found",
			uuid:           "nonexistent",
			serverResponse: `{"error":"not found"}`,
			statusCode:     http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.uuid+"/stop", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			response, err := dbService.Stop(context.Background(), tt.uuid)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMessage, response.Message)
		})
	}
}

func TestDatabaseService_Restart(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		serverResponse string
		statusCode     int
		wantErr        bool
		wantMessage    string
	}{
		{
			name:           "successful restart",
			uuid:           "db-uuid-1",
			serverResponse: `{"message":"Database restarted successfully"}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
			wantMessage:    "Database restarted successfully",
		},
		{
			name:           "not found",
			uuid:           "nonexistent",
			serverResponse: `{"error":"not found"}`,
			statusCode:     http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.uuid+"/restart", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			response, err := dbService.Restart(context.Background(), tt.uuid)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMessage, response.Message)
		})
	}
}

func TestDatabaseService_ListBackups(t *testing.T) {
	tests := []struct {
		name           string
		dbUUID         string
		serverResponse string
		statusCode     int
		wantErr        bool
		wantCount      int
	}{
		{
			name:   "successful list",
			dbUUID: "db-uuid-1",
			serverResponse: `[
				{
					"uuid": "backup-uuid-1",
					"enabled": true,
					"frequency": "0 2 * * *",
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			]`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:           "empty list",
			dbUUID:         "db-uuid-2",
			serverResponse: `[]`,
			statusCode:     http.StatusOK,
			wantErr:        false,
			wantCount:      0,
		},
		{
			name:           "not found",
			dbUUID:         "nonexistent",
			serverResponse: `{"error":"not found"}`,
			statusCode:     http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.dbUUID+"/backups", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			backups, err := dbService.ListBackups(context.Background(), tt.dbUUID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, backups, tt.wantCount)
		})
	}
}

func TestDatabaseService_UpdateBackup(t *testing.T) {
	tests := []struct {
		name       string
		dbUUID     string
		backupUUID string
		request    *models.DatabaseBackupUpdateRequest
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful update",
			dbUUID:     "db-uuid-1",
			backupUUID: "backup-uuid-1",
			request: &models.DatabaseBackupUpdateRequest{
				Enabled: boolPtr(true),
			},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "not found",
			dbUUID:     "db-uuid-1",
			backupUUID: "nonexistent",
			request: &models.DatabaseBackupUpdateRequest{
				Enabled: boolPtr(false),
			},
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.dbUUID+"/backups/"+tt.backupUUID, r.URL.Path)
				assert.Equal(t, http.MethodPatch, r.Method)
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusNotFound {
					_, _ = w.Write([]byte(`{"error":"not found"}`))
				}
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			err := dbService.UpdateBackup(context.Background(), tt.dbUUID, tt.backupUUID, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDatabaseService_DeleteBackup(t *testing.T) {
	tests := []struct {
		name       string
		dbUUID     string
		backupUUID string
		deleteS3   bool
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful delete with S3",
			dbUUID:     "db-uuid-1",
			backupUUID: "backup-uuid-1",
			deleteS3:   true,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "successful delete without S3",
			dbUUID:     "db-uuid-1",
			backupUUID: "backup-uuid-2",
			deleteS3:   false,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "not found",
			dbUUID:     "db-uuid-1",
			backupUUID: "nonexistent",
			deleteS3:   false,
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.dbUUID+"/backups/"+tt.backupUUID, r.URL.Path)
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.RawQuery, fmt.Sprintf("delete_s3=%t", tt.deleteS3))
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_, _ = w.Write([]byte(`{"message":"Backup deleted"}`))
				} else {
					_, _ = w.Write([]byte(`{"error":"not found"}`))
				}
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			err := dbService.DeleteBackup(context.Background(), tt.dbUUID, tt.backupUUID, tt.deleteS3)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDatabaseService_ListBackupExecutions(t *testing.T) {
	tests := []struct {
		name           string
		dbUUID         string
		backupUUID     string
		serverResponse string
		statusCode     int
		wantErr        bool
		wantCount      int
	}{
		{
			name:       "successful list",
			dbUUID:     "db-uuid-1",
			backupUUID: "backup-uuid-1",
			serverResponse: `{
				"executions": [
					{
						"uuid": "exec-uuid-1",
						"filename": "backup.sql",
						"size": 1024,
						"status": "success",
						"created_at": "2024-01-01T00:00:00Z"
					}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:       "empty list",
			dbUUID:     "db-uuid-2",
			backupUUID: "backup-uuid-2",
			serverResponse: `{
				"executions": []
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:           "not found",
			dbUUID:         "db-uuid-1",
			backupUUID:     "nonexistent",
			serverResponse: `{"error":"not found"}`,
			statusCode:     http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.dbUUID+"/backups/"+tt.backupUUID+"/executions", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			executions, err := dbService.ListBackupExecutions(context.Background(), tt.dbUUID, tt.backupUUID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, executions, tt.wantCount)
		})
	}
}

func TestDatabaseService_DeleteBackupExecution(t *testing.T) {
	tests := []struct {
		name          string
		dbUUID        string
		backupUUID    string
		executionUUID string
		deleteS3      bool
		statusCode    int
		wantErr       bool
	}{
		{
			name:          "successful delete with S3",
			dbUUID:        "db-uuid-1",
			backupUUID:    "backup-uuid-1",
			executionUUID: "exec-uuid-1",
			deleteS3:      true,
			statusCode:    http.StatusOK,
			wantErr:       false,
		},
		{
			name:          "successful delete without S3",
			dbUUID:        "db-uuid-1",
			backupUUID:    "backup-uuid-1",
			executionUUID: "exec-uuid-2",
			deleteS3:      false,
			statusCode:    http.StatusOK,
			wantErr:       false,
		},
		{
			name:          "not found",
			dbUUID:        "db-uuid-1",
			backupUUID:    "backup-uuid-1",
			executionUUID: "nonexistent",
			deleteS3:      false,
			statusCode:    http.StatusNotFound,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/databases/"+tt.dbUUID+"/backups/"+tt.backupUUID+"/executions/"+tt.executionUUID, r.URL.Path)
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.RawQuery, fmt.Sprintf("delete_s3=%t", tt.deleteS3))
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					_, _ = w.Write([]byte(`{"message":"Backup execution deleted"}`))
				} else {
					_, _ = w.Write([]byte(`{"error":"not found"}`))
				}
			}))
			defer server.Close()

			client := api.NewClient(server.URL, "test-token")
			dbService := NewDatabaseService(client)

			err := dbService.DeleteBackupExecution(context.Background(), tt.dbUUID, tt.backupUUID, tt.executionUUID, tt.deleteS3)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

// --- Environment Variable Tests ---

func TestDatabaseService_ListEnvs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"uuid": "env-1", "key": "DB_HOST", "value": "localhost", "is_literal": false},
			{"uuid": "env-2", "key": "DB_PORT", "value": "5432", "is_literal": true}
		]`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	envs, err := svc.ListEnvs(context.Background(), "db-uuid-123")
	require.NoError(t, err)
	assert.Len(t, envs, 2)
	assert.Equal(t, "DB_HOST", envs[0].Key)
	assert.Equal(t, "DB_PORT", envs[1].Key)
}

func TestDatabaseService_ListEnvs_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"database not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	envs, err := svc.ListEnvs(context.Background(), "db-uuid-123")
	require.Error(t, err)
	assert.Nil(t, envs)
	assert.Contains(t, err.Error(), "failed to list environment variables")
}

func TestDatabaseService_GetEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-uuid-123/envs", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"uuid": "env-1", "key": "DB_HOST", "value": "localhost"},
			{"uuid": "env-2", "key": "DB_PORT", "value": "5432"}
		]`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	// Find by UUID
	env, err := svc.GetEnv(context.Background(), "db-uuid-123", "env-2")
	require.NoError(t, err)
	assert.Equal(t, "DB_PORT", env.Key)

	// Find by key
	env, err = svc.GetEnv(context.Background(), "db-uuid-123", "DB_HOST")
	require.NoError(t, err)
	assert.Equal(t, "env-1", env.UUID)

	// Not found
	env, err = svc.GetEnv(context.Background(), "db-uuid-123", "NONEXISTENT")
	require.Error(t, err)
	assert.Nil(t, env)
	assert.Contains(t, err.Error(), "not found")
}

func TestDatabaseService_CreateEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.DatabaseEnvironmentVariableCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "NEW_VAR", req.Key)
		assert.Equal(t, "new_value", req.Value)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"uuid": "env-new",
			"key": "NEW_VAR",
			"value": "new_value",
			"comment": "test comment"
		}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	comment := "test comment"
	env, err := svc.CreateEnv(context.Background(), "db-uuid-123", &models.DatabaseEnvironmentVariableCreateRequest{
		Key:     "NEW_VAR",
		Value:   "new_value",
		Comment: &comment,
	})

	require.NoError(t, err)
	assert.Equal(t, "NEW_VAR", env.Key)
	assert.Equal(t, "new_value", env.Value)
	assert.Equal(t, "test comment", *env.Comment)
}

func TestDatabaseService_CreateEnv_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"validation failed"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	env, err := svc.CreateEnv(context.Background(), "db-uuid-123", &models.DatabaseEnvironmentVariableCreateRequest{
		Key:   "NEW_VAR",
		Value: "new_value",
	})

	require.Error(t, err)
	assert.Nil(t, env)
	assert.Contains(t, err.Error(), "failed to create environment variable")
}

func TestDatabaseService_UpdateEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-uuid-123/envs", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		comment := "updated comment"
		env := models.DatabaseEnvironmentVariable{
			UUID:    "env-1",
			Key:     "DB_HOST",
			Value:   "newhost",
			Comment: &comment,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(env)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	newKey := "DB_HOST"
	newValue := "newhost"
	newComment := "updated comment"
	req := &models.DatabaseEnvironmentVariableUpdateRequest{
		Key:     &newKey,
		Value:   &newValue,
		Comment: &newComment,
	}

	result, err := svc.UpdateEnv(context.Background(), "db-uuid-123", req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "DB_HOST", result.Key)
	assert.Equal(t, "newhost", result.Value)
	assert.Equal(t, "updated comment", *result.Comment)
}

func TestDatabaseService_UpdateEnv_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"environment variable not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	newKey := "DB_HOST"
	newValue := "newhost"
	result, err := svc.UpdateEnv(context.Background(), "db-uuid-123", &models.DatabaseEnvironmentVariableUpdateRequest{
		Key:   &newKey,
		Value: &newValue,
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update environment variable")
}

func TestDatabaseService_DeleteEnv(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-uuid-123/envs/env-1", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	err := svc.DeleteEnv(context.Background(), "db-uuid-123", "env-1")
	require.NoError(t, err)
}

func TestDatabaseService_DeleteEnv_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	err := svc.DeleteEnv(context.Background(), "db-uuid-123", "env-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete environment variable")
}

func TestDatabaseService_BulkUpdateEnvs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-uuid-123/envs/bulk", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"uuid": "env-1", "key": "DB_HOST", "value": "localhost"},
			{"uuid": "env-2", "key": "DB_PORT", "value": "5432"}
		]`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	req := &models.DatabaseEnvBulkUpdateRequest{
		Data: []models.DatabaseEnvironmentVariableCreateRequest{
			{Key: "DB_HOST", Value: "localhost"},
			{Key: "DB_PORT", Value: "5432"},
		},
	}

	result, err := svc.BulkUpdateEnvs(context.Background(), "db-uuid-123", req)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestDatabaseService_ListStorages(t *testing.T) {
	hostPath := "/var/data"
	content := "key: value"
	resp := models.StoragesResponse{
		PersistentStorages: []models.PersistentStorage{
			{
				ID:                     1,
				UUID:                   "ps-uuid-1",
				Name:                   "data-volume",
				MountPath:              "/data",
				HostPath:               &hostPath,
				IsPreviewSuffixEnabled: true,
			},
		},
		FileStorages: []models.FileStorage{
			{
				ID:                     2,
				UUID:                   "fs-uuid-1",
				FsPath:                 "/app/config.yml",
				MountPath:              "/app/config.yml",
				Content:                &content,
				IsPreviewSuffixEnabled: false,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-uuid-123/storages", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	result, err := svc.ListStorages(context.Background(), "db-uuid-123")
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "persistent", result[0].Type)
	assert.Equal(t, "data-volume", result[0].Name)
	assert.Equal(t, true, result[0].IsPreviewSuffixEnabled)
	assert.Equal(t, "file", result[1].Type)
	assert.Equal(t, "/app/config.yml", result[1].Name)
}

func TestDatabaseService_CreateStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-uuid-123/storages", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.StorageCreateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "persistent", req.Type)
		assert.Equal(t, "/data", req.MountPath)

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	name := "my-volume"
	req := &models.StorageCreateRequest{
		Type:      "persistent",
		MountPath: "/data",
		Name:      &name,
	}

	err := svc.CreateStorage(context.Background(), "db-uuid-123", req)
	require.NoError(t, err)
}

func TestDatabaseService_UpdateStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-uuid-123/storages", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var req models.StorageUpdateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "persistent", req.Type)
		assert.NotNil(t, req.UUID)
		assert.Equal(t, "storage-uuid-1", *req.UUID)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	storageUUID := "storage-uuid-1"
	name := "new-name"
	req := &models.StorageUpdateRequest{
		UUID: &storageUUID,
		Type: "persistent",
		Name: &name,
	}

	err := svc.UpdateStorage(context.Background(), "db-uuid-123", req)
	require.NoError(t, err)
}

func TestDatabaseService_DeleteStorage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/databases/db-uuid-123/storages/storage-uuid-1", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"Storage deleted."}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	svc := NewDatabaseService(client)

	err := svc.DeleteStorage(context.Background(), "db-uuid-123", "storage-uuid-1")
	require.NoError(t, err)
}

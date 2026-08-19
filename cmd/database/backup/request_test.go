package backup

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCreateBackupRequestUsesRegisteredFlags(t *testing.T) {
	cmd := NewCreateCommand()
	require.NoError(t, cmd.Flags().Set("databases-to-backup", "appdb"))
	require.NoError(t, cmd.Flags().Set("retention-max-storage-locally", "2.5"))
	require.NoError(t, cmd.Flags().Set("retention-max-storage-s3", "3.5"))
	require.NoError(t, cmd.Flags().Set("disable-local-backup", "true"))

	req := buildCreateBackupRequest(cmd)

	require.NotNil(t, req.DatabasesToBackup)
	assert.Equal(t, "appdb", *req.DatabasesToBackup)
	require.NotNil(t, req.DatabaseBackupRetentionMaxStorageLocally)
	assert.Equal(t, 2.5, *req.DatabaseBackupRetentionMaxStorageLocally)
	require.NotNil(t, req.DatabaseBackupRetentionMaxStorageS3)
	assert.Equal(t, 3.5, *req.DatabaseBackupRetentionMaxStorageS3)
	require.NotNil(t, req.DisableLocalBackup)
	assert.True(t, *req.DisableLocalBackup)

	payload, err := json.Marshal(req)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"databases_to_backup":"appdb",
		"database_backup_retention_max_storage_locally":2.5,
		"database_backup_retention_max_storage_s3":3.5,
		"disable_local_backup":true
	}`, string(payload))
}

func TestBuildUpdateBackupRequestAcceptsDecimalStorage(t *testing.T) {
	cmd := NewUpdateCommand()
	require.NoError(t, cmd.Flags().Set("retention-max-storage-locally", "4.5"))
	require.NoError(t, cmd.Flags().Set("retention-max-storage-s3", "5.5"))

	req, hasChanges := buildUpdateBackupRequest(cmd)

	assert.True(t, hasChanges)
	require.NotNil(t, req.DatabaseBackupRetentionMaxStorageLocally)
	assert.Equal(t, 4.5, *req.DatabaseBackupRetentionMaxStorageLocally)
	require.NotNil(t, req.DatabaseBackupRetentionMaxStorageS3)
	assert.Equal(t, 5.5, *req.DatabaseBackupRetentionMaxStorageS3)
}

func TestBuildUpdateBackupRequestReportsNoChanges(t *testing.T) {
	_, hasChanges := buildUpdateBackupRequest(NewUpdateCommand())

	assert.False(t, hasChanges)
}

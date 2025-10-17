package models

// Database represents a standalone Coolify database
type Database struct {
	ID          int     `json:"-" table:"-"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Image       *string `json:"image,omitempty"`
	Status      string  `json:"status"`
	Type        string  `json:"type"` // postgresql, mysql, mongodb, redis, etc.

	// Network configuration
	IsPublic   *bool `json:"is_public,omitempty"`
	PublicPort *int  `json:"public_port,omitempty"`

	// Resource limits (hidden from CLI output)
	LimitsMemory            *string `json:"limits_memory,omitempty" table:"-"`
	LimitsMemorySwap        *string `json:"limits_memory_swap,omitempty" table:"-"`
	LimitsMemorySwappiness  *int    `json:"limits_memory_swappiness,omitempty" table:"-"`
	LimitsMemoryReservation *string `json:"limits_memory_reservation,omitempty" table:"-"`
	LimitsCpus              *string `json:"limits_cpus,omitempty" table:"-"`
	LimitsCpuset            *string `json:"limits_cpuset,omitempty" table:"-"`
	LimitsCpuShares         *int    `json:"limits_cpu_shares,omitempty" table:"-"`

	// PostgreSQL specific
	PostgresUser           *string `json:"postgres_user,omitempty" table:"-"`
	PostgresPassword       *string `json:"postgres_password,omitempty" table:"-"`
	PostgresDb             *string `json:"postgres_db,omitempty" table:"-"`
	PostgresInitdbArgs     *string `json:"postgres_initdb_args,omitempty" table:"-"`
	PostgresHostAuthMethod *string `json:"postgres_host_auth_method,omitempty" table:"-"`
	PostgresConf           *string `json:"postgres_conf,omitempty" table:"-"`

	// MySQL specific
	MysqlRootPassword *string `json:"mysql_root_password,omitempty" table:"-"`
	MysqlPassword     *string `json:"mysql_password,omitempty" table:"-"`
	MysqlUser         *string `json:"mysql_user,omitempty" table:"-"`
	MysqlDatabase     *string `json:"mysql_database,omitempty" table:"-"`
	MysqlConf         *string `json:"mysql_conf,omitempty" table:"-"`

	// MariaDB specific
	MariadbRootPassword *string `json:"mariadb_root_password,omitempty" table:"-"`
	MariadbPassword     *string `json:"mariadb_password,omitempty" table:"-"`
	MariadbUser         *string `json:"mariadb_user,omitempty" table:"-"`
	MariadbDatabase     *string `json:"mariadb_database,omitempty" table:"-"`
	MariadbConf         *string `json:"mariadb_conf,omitempty" table:"-"`

	// MongoDB specific
	MongoInitdbRootUsername *string `json:"mongo_initdb_root_username,omitempty" table:"-"`
	MongoInitdbRootPassword *string `json:"mongo_initdb_root_password,omitempty" table:"-"`
	MongoInitdbDatabase     *string `json:"mongo_initdb_database,omitempty" table:"-"`
	MongoConf               *string `json:"mongo_conf,omitempty" table:"-"`

	// Redis specific
	RedisPassword *string `json:"redis_password,omitempty" table:"-"`
	RedisConf     *string `json:"redis_conf,omitempty" table:"-"`

	// KeyDB specific
	KeydbPassword *string `json:"keydb_password,omitempty" table:"-"`
	KeydbConf     *string `json:"keydb_conf,omitempty" table:"-"`

	// Clickhouse specific
	ClickhouseAdminUser     *string `json:"clickhouse_admin_user,omitempty" table:"-"`
	ClickhouseAdminPassword *string `json:"clickhouse_admin_password,omitempty" table:"-"`

	// Dragonfly specific
	DragonflyPassword *string `json:"dragonfly_password,omitempty" table:"-"`

	// Relationship IDs - internal database IDs (hidden from output)
	ServerID      *int `json:"-" table:"-"`
	EnvironmentID *int `json:"-" table:"-"`
	ProjectID     *int `json:"-" table:"-"`

	// Metadata
	CreatedAt string `json:"-" table:"-"`
	UpdatedAt string `json:"-" table:"-"`
}

// DatabaseCreateRequest represents the base request to create a database
type DatabaseCreateRequest struct {
	ServerUUID      string  `json:"server_uuid"`
	ProjectUUID     string  `json:"project_uuid"`
	EnvironmentName *string `json:"environment_name,omitempty"`
	EnvironmentUUID *string `json:"environment_uuid,omitempty"`
	DestinationUUID *string `json:"destination_uuid,omitempty"`
	InstantDeploy   *bool   `json:"instant_deploy,omitempty"`

	// Common fields
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Image       *string `json:"image,omitempty"`
	IsPublic    *bool   `json:"is_public,omitempty"`
	PublicPort  *int    `json:"public_port,omitempty"`

	// Resource limits (hidden from CLI output)
	LimitsMemory            *string `json:"limits_memory,omitempty" table:"-"`
	LimitsMemorySwap        *string `json:"limits_memory_swap,omitempty" table:"-"`
	LimitsMemorySwappiness  *int    `json:"limits_memory_swappiness,omitempty" table:"-"`
	LimitsMemoryReservation *string `json:"limits_memory_reservation,omitempty" table:"-"`
	LimitsCpus              *string `json:"limits_cpus,omitempty" table:"-"`
	LimitsCpuset            *string `json:"limits_cpuset,omitempty" table:"-"`
	LimitsCpuShares         *int    `json:"limits_cpu_shares,omitempty" table:"-"`

	// PostgreSQL specific
	PostgresUser           *string `json:"postgres_user,omitempty"`
	PostgresPassword       *string `json:"postgres_password,omitempty"`
	PostgresDb             *string `json:"postgres_db,omitempty" table:"-"`
	PostgresInitdbArgs     *string `json:"postgres_initdb_args,omitempty"`
	PostgresHostAuthMethod *string `json:"postgres_host_auth_method,omitempty"`
	PostgresConf           *string `json:"postgres_conf,omitempty"`

	// MySQL specific
	MysqlRootPassword *string `json:"mysql_root_password,omitempty"`
	MysqlPassword     *string `json:"mysql_password,omitempty"`
	MysqlUser         *string `json:"mysql_user,omitempty"`
	MysqlDatabase     *string `json:"mysql_database,omitempty" table:"-"`
	MysqlConf         *string `json:"mysql_conf,omitempty"`

	// MariaDB specific
	MariadbRootPassword *string `json:"mariadb_root_password,omitempty"`
	MariadbPassword     *string `json:"mariadb_password,omitempty"`
	MariadbUser         *string `json:"mariadb_user,omitempty"`
	MariadbDatabase     *string `json:"mariadb_database,omitempty" table:"-"`
	MariadbConf         *string `json:"mariadb_conf,omitempty"`

	// MongoDB specific
	MongoInitdbRootUsername *string `json:"mongo_initdb_root_username,omitempty"`
	MongoInitdbRootPassword *string `json:"mongo_initdb_root_password,omitempty"`
	MongoInitdbDatabase     *string `json:"mongo_initdb_database,omitempty" table:"-"`
	MongoConf               *string `json:"mongo_conf,omitempty"`

	// Redis specific
	RedisPassword *string `json:"redis_password,omitempty"`
	RedisConf     *string `json:"redis_conf,omitempty"`

	// KeyDB specific
	KeydbPassword *string `json:"keydb_password,omitempty"`
	KeydbConf     *string `json:"keydb_conf,omitempty"`

	// Clickhouse specific
	ClickhouseAdminUser     *string `json:"clickhouse_admin_user,omitempty"`
	ClickhouseAdminPassword *string `json:"clickhouse_admin_password,omitempty"`

	// Dragonfly specific
	DragonflyPassword *string `json:"dragonfly_password,omitempty"`
}

// DatabaseUpdateRequest represents the request to update a database
// Only common configuration fields that make sense to update after creation
type DatabaseUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Image       *string `json:"image,omitempty"`
	Description *string `json:"description,omitempty"`
	IsPublic    *bool   `json:"is_public,omitempty"`
	PublicPort  *int    `json:"public_port,omitempty"`

	// Resource limits
	LimitsMemory *string `json:"limits_memory,omitempty"`
	LimitsCpus   *string `json:"limits_cpus,omitempty"`
}

// DatabaseLifecycleResponse represents the response from lifecycle operations
type DatabaseLifecycleResponse struct {
	Message string `json:"message"`
}

// DatabaseBackup represents a scheduled database backup configuration
type DatabaseBackup struct {
	ID                                       int     `json:"-" table:"-"`
	UUID                                     string  `json:"uuid"`
	Description                              *string `json:"description,omitempty"`
	Enabled                                  *bool   `json:"enabled,omitempty"`
	Frequency                                *string `json:"frequency,omitempty"`
	SaveS3                                   *bool   `json:"save_s3,omitempty"`
	S3StorageID                              *int    `json:"-" table:"-"`
	DatabasesToBackup                        *string `json:"databases_to_backup,omitempty"`
	DumpAll                                  *bool   `json:"dump_all,omitempty"`
	DatabaseBackupRetentionAmountLocally     *int    `json:"database_backup_retention_amount_locally,omitempty"`
	DatabaseBackupRetentionDaysLocally       *int    `json:"database_backup_retention_days_locally,omitempty"`
	DatabaseBackupRetentionMaxStorageLocally *string `json:"database_backup_retention_max_storage_locally,omitempty"`
	DatabaseBackupRetentionAmountS3          *int    `json:"database_backup_retention_amount_s3,omitempty"`
	DatabaseBackupRetentionDaysS3            *int    `json:"database_backup_retention_days_s3,omitempty"`
	DatabaseBackupRetentionMaxStorageS3      *string `json:"database_backup_retention_max_storage_s3,omitempty"`
	DatabaseType                             *string `json:"database_type,omitempty" table:"-"`
	DatabaseID                               *int    `json:"-" table:"-"`
	TeamID                                   *int    `json:"-" table:"-"`
	Timeout                                  *int    `json:"timeout,omitempty"`
	DisableLocalBackup                       *bool   `json:"disable_local_backup,omitempty"`
	CreatedAt                                string  `json:"-" table:"-"`
	UpdatedAt                                string  `json:"-" table:"-"`
}

// DatabaseBackupCreateRequest represents the request to create a backup configuration
type DatabaseBackupCreateRequest struct {
	Frequency                                *string `json:"frequency,omitempty"`
	Enabled                                  *bool   `json:"enabled,omitempty"`
	SaveS3                                   *bool   `json:"save_s3,omitempty"`
	S3StorageUUID                            *string `json:"s3_storage_uuid,omitempty"`
	DatabasesToBackup                        *string `json:"databases_to_backup,omitempty"`
	DumpAll                                  *bool   `json:"dump_all,omitempty"`
	DatabaseBackupRetentionAmountLocally     *int    `json:"database_backup_retention_amount_locally,omitempty"`
	DatabaseBackupRetentionDaysLocally       *int    `json:"database_backup_retention_days_locally,omitempty"`
	DatabaseBackupRetentionMaxStorageLocally *string `json:"database_backup_retention_max_storage_locally,omitempty"`
	DatabaseBackupRetentionAmountS3          *int    `json:"database_backup_retention_amount_s3,omitempty"`
	DatabaseBackupRetentionDaysS3            *int    `json:"database_backup_retention_days_s3,omitempty"`
	DatabaseBackupRetentionMaxStorageS3      *string `json:"database_backup_retention_max_storage_s3,omitempty"`
	Timeout                                  *int    `json:"timeout,omitempty"`
	DisableLocalBackup                       *bool   `json:"disable_local_backup,omitempty"`
}

// DatabaseBackupUpdateRequest represents the request to update a backup configuration
type DatabaseBackupUpdateRequest struct {
	SaveS3                                   *bool   `json:"save_s3,omitempty"`
	S3StorageUUID                            *string `json:"s3_storage_uuid,omitempty"`
	BackupNow                                *bool   `json:"backup_now,omitempty"`
	Enabled                                  *bool   `json:"enabled,omitempty"`
	DatabasesToBackup                        *string `json:"databases_to_backup,omitempty"`
	DumpAll                                  *bool   `json:"dump_all,omitempty"`
	Frequency                                *string `json:"frequency,omitempty"`
	DatabaseBackupRetentionAmountLocally     *int    `json:"database_backup_retention_amount_locally,omitempty"`
	DatabaseBackupRetentionDaysLocally       *int    `json:"database_backup_retention_days_locally,omitempty"`
	DatabaseBackupRetentionMaxStorageLocally *int    `json:"database_backup_retention_max_storage_locally,omitempty"`
	DatabaseBackupRetentionAmountS3          *int    `json:"database_backup_retention_amount_s3,omitempty"`
	DatabaseBackupRetentionDaysS3            *int    `json:"database_backup_retention_days_s3,omitempty"`
	DatabaseBackupRetentionMaxStorageS3      *int    `json:"database_backup_retention_max_storage_s3,omitempty"`
}

// DatabaseBackupExecution represents a single backup execution
type DatabaseBackupExecution struct {
	UUID      string  `json:"uuid"`
	Filename  *string `json:"filename,omitempty"`
	Size      *int    `json:"size,omitempty"`
	Status    *string `json:"status,omitempty"`
	Message   *string `json:"message,omitempty"`
	CreatedAt string  `json:"-" table:"-"`
}

// DatabaseBackupExecutionsResponse represents the response containing backup executions
type DatabaseBackupExecutionsResponse struct {
	Executions []DatabaseBackupExecution `json:"executions"`
}

// DatabaseBackupResponse represents a generic backup operation response
type DatabaseBackupResponse struct {
	Message string `json:"message"`
}

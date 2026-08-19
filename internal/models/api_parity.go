package models

// S3Storage represents a team S3/compatible storage destination.
type S3Storage struct {
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Endpoint    string  `json:"endpoint"`
	Bucket      string  `json:"bucket"`
	Region      string  `json:"region"`
	Key         *string `json:"key,omitempty" sensitive:"true"`
	Secret      *string `json:"secret,omitempty" sensitive:"true"`
	IsUsable    *bool   `json:"is_usable,omitempty"`
	TeamID      int     `json:"team_id" table:"-"`
	CreatedAt   string  `json:"created_at" table:"-"`
	UpdatedAt   string  `json:"updated_at" table:"-"`
}

type S3StorageCreateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Endpoint    string  `json:"endpoint"`
	Bucket      string  `json:"bucket"`
	Region      string  `json:"region"`
	Key         string  `json:"key"`
	Secret      string  `json:"secret"`
	IsUsable    *bool   `json:"is_usable,omitempty"`
}

type S3StorageUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Endpoint    *string `json:"endpoint,omitempty"`
	Bucket      *string `json:"bucket,omitempty"`
	Region      *string `json:"region,omitempty"`
	Key         *string `json:"key,omitempty"`
	Secret      *string `json:"secret,omitempty"`
	IsUsable    *bool   `json:"is_usable,omitempty"`
}

type S3StorageValidation struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// SharedEnvironmentVariable is a hierarchical shared env (team/project/environment/server).
type SharedEnvironmentVariable struct {
	ID          int     `json:"id"`
	Key         string  `json:"key"`
	Value       *string `json:"value,omitempty" sensitive:"true"`
	IsLiteral   bool    `json:"is_literal"`
	IsMultiline bool    `json:"is_multiline"`
	IsShownOnce bool    `json:"is_shown_once"`
	Comment     *string `json:"comment,omitempty"`
	Type        string  `json:"type,omitempty" table:"-"`
	CreatedAt   string  `json:"created_at" table:"-"`
	UpdatedAt   string  `json:"updated_at" table:"-"`
}

type SharedEnvCreateRequest struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	IsLiteral   *bool   `json:"is_literal,omitempty"`
	IsMultiline *bool   `json:"is_multiline,omitempty"`
	IsShownOnce *bool   `json:"is_shown_once,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

type SharedEnvUpdateRequest struct {
	Key         *string `json:"key,omitempty"`
	Value       *string `json:"value,omitempty"`
	IsLiteral   *bool   `json:"is_literal,omitempty"`
	IsMultiline *bool   `json:"is_multiline,omitempty"`
	IsShownOnce *bool   `json:"is_shown_once,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

type SharedEnvCreateResponse struct {
	ID int `json:"id"`
}

// CloudInitScript is a team cloud-init template.
type CloudInitScript struct {
	UUID      string  `json:"uuid"`
	Name      string  `json:"name"`
	Script    *string `json:"script,omitempty" sensitive:"true"`
	CreatedAt string  `json:"created_at" table:"-"`
	UpdatedAt string  `json:"updated_at" table:"-"`
}

type CloudInitScriptCreateRequest struct {
	Name   string `json:"name"`
	Script string `json:"script"`
}

type CloudInitScriptUpdateRequest struct {
	Name   *string `json:"name,omitempty"`
	Script *string `json:"script,omitempty"`
}

// Server subsystem settings (JSON map-friendly structs for known fields).

type DockerCleanupSettings struct {
	DockerCleanupFrequency           string `json:"docker_cleanup_frequency"`
	DockerCleanupThreshold           int    `json:"docker_cleanup_threshold"`
	ForceDockerCleanup               bool   `json:"force_docker_cleanup"`
	DeleteUnusedVolumes              bool   `json:"delete_unused_volumes"`
	DeleteUnusedNetworks             bool   `json:"delete_unused_networks"`
	DisableApplicationImageRetention bool   `json:"disable_application_image_retention"`
}

type DockerCleanupUpdateRequest struct {
	DockerCleanupFrequency           *string `json:"docker_cleanup_frequency,omitempty"`
	DockerCleanupThreshold           *int    `json:"docker_cleanup_threshold,omitempty"`
	ForceDockerCleanup               *bool   `json:"force_docker_cleanup,omitempty"`
	DeleteUnusedVolumes              *bool   `json:"delete_unused_volumes,omitempty"`
	DeleteUnusedNetworks             *bool   `json:"delete_unused_networks,omitempty"`
	DisableApplicationImageRetention *bool   `json:"disable_application_image_retention,omitempty"`
}

type ServerProxySettings struct {
	ProxyType           string  `json:"proxy_type"`
	Status              string  `json:"status"`
	RedirectEnabled     bool    `json:"redirect_enabled"`
	RedirectURL         *string `json:"redirect_url,omitempty"`
	GenerateExactLabels bool    `json:"generate_exact_labels"`
	Configuration       *string `json:"configuration,omitempty" table:"-"`
}

type ServerProxyUpdateRequest struct {
	RedirectEnabled     *bool   `json:"redirect_enabled,omitempty"`
	RedirectURL         *string `json:"redirect_url,omitempty"`
	GenerateExactLabels *bool   `json:"generate_exact_labels,omitempty"`
	ProxyType           *string `json:"proxy_type,omitempty"`
}

type ProxyConfigurationRequest struct {
	Configuration string `json:"configuration"`
}

type DestinationUpdateRequest struct {
	Name string `json:"name"`
}

type TagCreateRequest struct {
	Name string `json:"name"`
}

type TagUpdateRequest struct {
	Name string `json:"name"`
}

type EnvironmentUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type ResourceCloneRequest struct {
	DestinationUUID string  `json:"destination_uuid"`
	Name            *string `json:"name,omitempty"`
	CloneVolumes    *bool   `json:"clone_volumes,omitempty"`
}

type ResourceCloneResponse struct {
	UUID string `json:"uuid"`
}

type RollbackRequest struct {
	Commit string `json:"commit"`
}

type AppDestinationAttachRequest struct {
	DestinationUUID string `json:"destination_uuid"`
}

// VolumeBackupScheduleRequest is the body for PUT .../storages/{uuid}/backups.
type VolumeBackupScheduleRequest struct {
	Frequency                  string   `json:"frequency"`
	Enabled                    *bool    `json:"enabled,omitempty"`
	SaveS3                     *bool    `json:"save_s3,omitempty"`
	DisableLocalBackup         *bool    `json:"disable_local_backup,omitempty"`
	StopDuringBackup           *bool    `json:"stop_during_backup,omitempty"`
	S3StorageUUID              *string  `json:"s3_storage_uuid,omitempty"`
	RetentionAmountLocally     *int     `json:"retention_amount_locally,omitempty"`
	RetentionDaysLocally       *int     `json:"retention_days_locally,omitempty"`
	RetentionMaxStorageLocally *float64 `json:"retention_max_storage_locally,omitempty"`
	RetentionAmountS3          *int     `json:"retention_amount_s3,omitempty"`
	RetentionDaysS3            *int     `json:"retention_days_s3,omitempty"`
	RetentionMaxStorageS3      *float64 `json:"retention_max_storage_s3,omitempty"`
	Timeout                    *int     `json:"timeout,omitempty"`
}

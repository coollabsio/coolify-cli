package models

// Service represents a Coolify one-click service
type Service struct {
	ID          int     `json:"-" table:"-"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`

	// Relationship IDs - internal database IDs (hidden from output)
	ServerID      *int `json:"-" table:"-"`
	EnvironmentID *int `json:"-" table:"-"`
	ProjectID     *int `json:"-" table:"-"`

	// Docker configuration (hidden from table output)
	DockerCompose    *string `json:"docker_compose,omitempty" table:"-"`
	DockerComposeRaw *string `json:"docker_compose_raw,omitempty" table:"-"`

	// Additional metadata
	CreatedAt string `json:"-" table:"-"`
	UpdatedAt string `json:"-" table:"-"`

	// Nested resources
	Applications []ServiceApplication `json:"applications,omitempty"`
	Databases    []ServiceDatabase    `json:"databases,omitempty"`
}

// ServiceApplication represents an application within a service
type ServiceApplication struct {
	ID                   int     `json:"-" table:"-"`
	UUID                 string  `json:"uuid"`
	Name                 string  `json:"name"`
	HumanName            *string `json:"human_name,omitempty"`
	Description          *string `json:"description,omitempty"`
	Fqdn                 *string `json:"fqdn,omitempty"`
	Ports                *string `json:"ports,omitempty" table:"-"`
	Exposes              *string `json:"exposes,omitempty" table:"-"`
	Status               *string `json:"status,omitempty"`
	ServiceID            *int    `json:"service_id,omitempty" table:"-"`
	ExcludeFromStatus    *bool   `json:"exclude_from_status,omitempty" table:"-"`
	RequiredFqdn         *bool   `json:"required_fqdn,omitempty" table:"-"`
	Image                *string `json:"image,omitempty" table:"-"`
	IsLogDrainEnabled    *bool   `json:"is_log_drain_enabled,omitempty" table:"-"`
	IsIncludeTimestamps  *bool   `json:"is_include_timestamps,omitempty" table:"-"`
	IsGzipEnabled        *bool   `json:"is_gzip_enabled,omitempty" table:"-"`
	IsStripprefixEnabled *bool   `json:"is_stripprefix_enabled,omitempty" table:"-"`
	LastOnlineAt         *string `json:"last_online_at,omitempty" table:"-"`
	IsMigrated           *bool   `json:"is_migrated,omitempty" table:"-"`
	CreatedAt            string  `json:"created_at,omitempty" table:"-"`
	UpdatedAt            string  `json:"updated_at,omitempty" table:"-"`
	DeletedAt            *string `json:"deleted_at,omitempty" table:"-"`
}

// ServiceApplicationUpdateRequest represents mutable service application fields.
type ServiceApplicationUpdateRequest struct {
	URL                  *string `json:"url,omitempty"`
	HumanName            *string `json:"human_name,omitempty"`
	Description          *string `json:"description,omitempty"`
	Image                *string `json:"image,omitempty"`
	ExcludeFromStatus    *bool   `json:"exclude_from_status,omitempty"`
	IsLogDrainEnabled    *bool   `json:"is_log_drain_enabled,omitempty"`
	IsGzipEnabled        *bool   `json:"is_gzip_enabled,omitempty"`
	IsStripprefixEnabled *bool   `json:"is_stripprefix_enabled,omitempty"`
}

// ServiceDatabase represents a database within a service
type ServiceDatabase struct {
	ID                  int     `json:"-" table:"-"`
	UUID                string  `json:"uuid"`
	Name                string  `json:"name"`
	HumanName           *string `json:"human_name,omitempty"`
	Description         *string `json:"description,omitempty"`
	Type                *string `json:"type,omitempty"`
	Status              *string `json:"status,omitempty"`
	Image               *string `json:"image,omitempty" table:"-"`
	Ports               *string `json:"ports,omitempty" table:"-"`
	Exposes             *string `json:"exposes,omitempty" table:"-"`
	ExcludeFromStatus   *bool   `json:"exclude_from_status,omitempty" table:"-"`
	IsLogDrainEnabled   *bool   `json:"is_log_drain_enabled,omitempty" table:"-"`
	IsIncludeTimestamps *bool   `json:"is_include_timestamps,omitempty" table:"-"`
	IsPublic            *bool   `json:"is_public,omitempty"`
	PublicPort          *int    `json:"public_port,omitempty"`
	PublicPortTimeout   *int    `json:"public_port_timeout,omitempty" table:"-"`
	LastOnlineAt        *string `json:"last_online_at,omitempty" table:"-"`
	CustomType          *string `json:"custom_type,omitempty" table:"-"`
	CreatedAt           string  `json:"created_at,omitempty" table:"-"`
	UpdatedAt           string  `json:"updated_at,omitempty" table:"-"`
	DeletedAt           *string `json:"deleted_at,omitempty" table:"-"`
}

// ServiceDatabaseUpdateRequest represents mutable service database fields.
type ServiceDatabaseUpdateRequest struct {
	HumanName         *string `json:"human_name,omitempty"`
	Description       *string `json:"description,omitempty"`
	Image             *string `json:"image,omitempty"`
	ExcludeFromStatus *bool   `json:"exclude_from_status,omitempty"`
	IsLogDrainEnabled *bool   `json:"is_log_drain_enabled,omitempty"`
	IsPublic          *bool   `json:"is_public,omitempty"`
	PublicPort        *int    `json:"public_port,omitempty"`
	PublicPortTimeout *int    `json:"public_port_timeout,omitempty"`
}

// ServiceCreateRequest represents the request to create a service
type ServiceCreateRequest struct {
	Type            string   `json:"type"`
	Name            *string  `json:"name,omitempty"`
	Description     *string  `json:"description,omitempty"`
	ServerUUID      string   `json:"server_uuid"`
	ProjectUUID     string   `json:"project_uuid"`
	EnvironmentName string   `json:"environment_name,omitempty"`
	EnvironmentUUID *string  `json:"environment,omitempty"`
	InstantDeploy   *bool    `json:"instant_deploy,omitempty"`
	DockerCompose   *string  `json:"docker_compose,omitempty"`
	Destination     *string  `json:"destination,omitempty"`
	Tags            []string `json:"tags,omitempty"`
}

// ServiceUpdateRequest represents the request to update a service
type ServiceUpdateRequest struct {
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	DockerCompose *string `json:"docker_compose,omitempty"`
}

// ServiceLifecycleResponse represents the response from lifecycle operations
type ServiceLifecycleResponse struct {
	Message string `json:"message"`
}

// ServiceEnvironmentVariable represents an environment variable for a service
// Services don't have preview deployments, so IsPreview is excluded from output
type ServiceEnvironmentVariable struct {
	ID             int     `json:"-" table:"-"`
	UUID           string  `json:"uuid"`
	Key            string  `json:"key"`
	Value          string  `json:"value" sensitive:"true"`
	IsBuildTime    bool    `json:"is_buildtime"`
	IsLiteralValue bool    `json:"is_literal"`
	IsShownOnce    bool    `json:"is_shown_once"`
	IsRuntime      bool    `json:"is_runtime"`
	IsShared       bool    `json:"is_shared"`
	Comment        *string `json:"comment,omitempty"`
	RealValue      *string `json:"real_value,omitempty" sensitive:"true"`
	ServiceID      *int    `json:"-" table:"-"`
	CreatedAt      string  `json:"-" table:"-"`
	UpdatedAt      string  `json:"-" table:"-"`
}

// ServiceEnvironmentVariableCreateRequest represents the request to create a service environment variable
type ServiceEnvironmentVariableCreateRequest struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	IsBuildTime *bool   `json:"is_build_time,omitempty"`
	IsLiteral   *bool   `json:"is_literal,omitempty"`
	IsMultiline *bool   `json:"is_multiline,omitempty"`
	IsRuntime   *bool   `json:"is_runtime,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

// ServiceEnvironmentVariableUpdateRequest represents the request to update a service environment variable
type ServiceEnvironmentVariableUpdateRequest struct {
	Key         *string `json:"key,omitempty"`
	Value       *string `json:"value,omitempty"`
	IsBuildTime *bool   `json:"is_build_time,omitempty"`
	IsLiteral   *bool   `json:"is_literal,omitempty"`
	IsMultiline *bool   `json:"is_multiline,omitempty"`
	IsRuntime   *bool   `json:"is_runtime,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

// ServiceEnvBulkUpdateRequest represents the request to bulk update service environment variables
type ServiceEnvBulkUpdateRequest struct {
	Data []ServiceEnvironmentVariableCreateRequest `json:"data"`
}

// ServiceEnvBulkUpdateResponse represents the response from service bulk update
type ServiceEnvBulkUpdateResponse []ServiceEnvironmentVariable

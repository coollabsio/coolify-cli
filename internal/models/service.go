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
	ID     int     `json:"-" table:"-"`
	UUID   string  `json:"uuid"`
	Name   string  `json:"name"`
	Status *string `json:"status,omitempty"`
	Fqdn   *string `json:"fqdn,omitempty"`
}

// ServiceDatabase represents a database within a service
type ServiceDatabase struct {
	ID     int     `json:"-" table:"-"`
	UUID   string  `json:"uuid"`
	Name   string  `json:"name"`
	Type   *string `json:"type,omitempty"`
	Status *string `json:"status,omitempty"`
}

// ServiceCreateRequest represents the request to create a service
type ServiceCreateRequest struct {
	Type            string  `json:"type"`
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	ServerUUID      string  `json:"server_uuid"`
	ProjectUUID     string  `json:"project_uuid"`
	EnvironmentName string  `json:"environment_name"`
	InstantDeploy   *bool   `json:"instant_deploy,omitempty"`
	DockerCompose   *string `json:"docker_compose,omitempty"`
	Destination     *string `json:"destination,omitempty"`
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
	RealValue      *string `json:"real_value,omitempty" sensitive:"true"`
	ServiceID      *int    `json:"-" table:"-"`
	CreatedAt      string  `json:"-" table:"-"`
	UpdatedAt      string  `json:"-" table:"-"`
}

// ServiceEnvironmentVariableCreateRequest represents the request to create a service environment variable
type ServiceEnvironmentVariableCreateRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsBuildTime *bool  `json:"is_build_time,omitempty"`
	IsLiteral   *bool  `json:"is_literal,omitempty"`
	IsMultiline *bool  `json:"is_multiline,omitempty"`
	IsRuntime   *bool  `json:"is_runtime,omitempty"`
}

// ServiceEnvironmentVariableUpdateRequest represents the request to update a service environment variable
type ServiceEnvironmentVariableUpdateRequest struct {
	UUID        string  `json:"uuid"`
	Key         *string `json:"key,omitempty"`
	Value       *string `json:"value,omitempty"`
	IsBuildTime *bool   `json:"is_build_time,omitempty"`
	IsLiteral   *bool   `json:"is_literal,omitempty"`
	IsMultiline *bool   `json:"is_multiline,omitempty"`
	IsRuntime   *bool   `json:"is_runtime,omitempty"`
}

// ServiceEnvBulkUpdateRequest represents the request to bulk update service environment variables
type ServiceEnvBulkUpdateRequest struct {
	Data []ServiceEnvironmentVariableCreateRequest `json:"data"`
}

// ServiceEnvBulkUpdateResponse represents the response from service bulk update
type ServiceEnvBulkUpdateResponse struct {
	Message string `json:"message,omitempty"`
}

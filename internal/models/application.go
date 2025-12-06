package models

// Application represents a Coolify application
type Application struct {
	ID          int     `json:"-" table:"-"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
	GitBranch   *string `json:"git_branch,omitempty"`
	FQDN        *string `json:"fqdn,omitempty"`
	CreatedAt   string  `json:"-" table:"-"`
	UpdatedAt   string  `json:"-" table:"-"`
}

// ApplicationListItem represents a simplified application for list view
type ApplicationListItem struct {
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
	GitBranch   *string `json:"git_branch,omitempty"`
	FQDN        *string `json:"fqdn,omitempty"`
}

// ApplicationUpdateRequest represents the request to update an application
// All fields are optional - only provided fields will be updated
type ApplicationUpdateRequest struct {
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	GitBranch        *string `json:"git_branch,omitempty"`
	GitRepository    *string `json:"git_repository,omitempty"`
	GitCommitSHA     *string `json:"git_commit_sha,omitempty"`
	Domains          *string `json:"domains,omitempty"`
	BuildCommand     *string `json:"build_command,omitempty"`
	StartCommand     *string `json:"start_command,omitempty"`
	InstallCommand   *string `json:"install_command,omitempty"`
	BaseDirectory    *string `json:"base_directory,omitempty"`
	PublishDirectory *string `json:"publish_directory,omitempty"`
	BuildPack        *string `json:"build_pack,omitempty"`
	PortsExposes     *string `json:"ports_exposes,omitempty"`
	PortsMappings    *string `json:"ports_mappings,omitempty"`

	// Docker configuration
	Dockerfile              *string `json:"dockerfile,omitempty"`
	DockerRegistryImageName *string `json:"docker_registry_image_name,omitempty"`
	DockerRegistryImageTag  *string `json:"docker_registry_image_tag,omitempty"`
	CustomDockerRunOptions  *string `json:"custom_docker_run_options,omitempty"`
	CustomLabels            *string `json:"custom_labels,omitempty"`

	// Health checks
	HealthCheckEnabled      *bool   `json:"health_check_enabled,omitempty"`
	HealthCheckPath         *string `json:"health_check_path,omitempty"`
	HealthCheckPort         *string `json:"health_check_port,omitempty"`
	HealthCheckHost         *string `json:"health_check_host,omitempty"`
	HealthCheckMethod       *string `json:"health_check_method,omitempty"`
	HealthCheckScheme       *string `json:"health_check_scheme,omitempty"`
	HealthCheckReturnCode   *int    `json:"health_check_return_code,omitempty"`
	HealthCheckResponseText *string `json:"health_check_response_text,omitempty"`
	HealthCheckInterval     *int    `json:"health_check_interval,omitempty"`
	HealthCheckTimeout      *int    `json:"health_check_timeout,omitempty"`
	HealthCheckRetries      *int    `json:"health_check_retries,omitempty"`
	HealthCheckStartPeriod  *int    `json:"health_check_start_period,omitempty"`

	// Resource limits
	LimitsCPUs              *string `json:"limits_cpus,omitempty"`
	LimitsCPUShares         *int    `json:"limits_cpu_shares,omitempty"`
	LimitsCPUSet            *string `json:"limits_cpuset,omitempty"`
	LimitsMemory            *string `json:"limits_memory,omitempty"`
	LimitsMemoryReservation *string `json:"limits_memory_reservation,omitempty"`
	LimitsMemorySwap        *string `json:"limits_memory_swap,omitempty"`
	LimitsMemorySwappiness  *int    `json:"limits_memory_swappiness,omitempty"`

	// Deployment hooks
	PreDeploymentCommand           *string `json:"pre_deployment_command,omitempty"`
	PreDeploymentCommandContainer  *string `json:"pre_deployment_command_container,omitempty"`
	PostDeploymentCommand          *string `json:"post_deployment_command,omitempty"`
	PostDeploymentCommandContainer *string `json:"post_deployment_command_container,omitempty"`

	// Misc
	Redirect   *string `json:"redirect,omitempty"`
	WatchPaths *string `json:"watch_paths,omitempty"`
	IsStatic   *bool   `json:"is_static,omitempty"`
}

// ApplicationLifecycleResponse represents the response from lifecycle operations
type ApplicationLifecycleResponse struct {
	Message        string  `json:"message"`
	DeploymentUUID *string `json:"deployment_uuid,omitempty"`
}

// ApplicationLogsResponse represents the response from logs endpoint
type ApplicationLogsResponse struct {
	Logs string `json:"logs"`
}

// EnvironmentVariable represents an environment variable for an application
type EnvironmentVariable struct {
	ID             int     `json:"-" table:"-"`
	UUID           string  `json:"uuid"`
	Key            string  `json:"key"`
	Value          string  `json:"value" sensitive:"true"`
	IsBuildTime    bool    `json:"is_buildtime"`
	IsPreview      bool    `json:"is_preview"`
	IsLiteralValue bool    `json:"is_literal"`
	IsShownOnce    bool    `json:"is_shown_once"`
	IsRuntime      bool    `json:"is_runtime"`
	IsShared       bool    `json:"is_shared"`
	RealValue      *string `json:"real_value,omitempty" sensitive:"true"`
	ApplicationID  *int    `json:"-" table:"-"`
	CreatedAt      string  `json:"-" table:"-"`
	UpdatedAt      string  `json:"-" table:"-"`
}

// EnvironmentVariableCreateRequest represents the request to create an environment variable
type EnvironmentVariableCreateRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsBuildTime *bool  `json:"is_build_time,omitempty"`
	IsPreview   *bool  `json:"is_preview,omitempty"`
	IsLiteral   *bool  `json:"is_literal,omitempty"`
	IsMultiline *bool  `json:"is_multiline,omitempty"`
	IsRuntime   *bool  `json:"is_runtime,omitempty"`
}

// EnvironmentVariableUpdateRequest represents the request to update an environment variable
type EnvironmentVariableUpdateRequest struct {
	UUID        string  `json:"uuid"`
	Key         *string `json:"key,omitempty"`
	Value       *string `json:"value,omitempty"`
	IsBuildTime *bool   `json:"is_build_time,omitempty"`
	IsPreview   *bool   `json:"is_preview,omitempty"`
	IsLiteral   *bool   `json:"is_literal,omitempty"`
	IsMultiline *bool   `json:"is_multiline,omitempty"`
	IsRuntime   *bool   `json:"is_runtime,omitempty"`
}

// ApplicationCreatePublicRequest for POST /applications/public
// Creates an application from a public git repository
type ApplicationCreatePublicRequest struct {
	// Required fields
	ProjectUUID   string `json:"project_uuid"`
	ServerUUID    string `json:"server_uuid"`
	GitRepository string `json:"git_repository"`
	GitBranch     string `json:"git_branch"`
	BuildPack     string `json:"build_pack"` // nixpacks, static, dockerfile, dockercompose
	PortsExposes  string `json:"ports_exposes"`

	// Environment (one of these is required)
	EnvironmentName *string `json:"environment_name,omitempty"`
	EnvironmentUUID *string `json:"environment,omitempty"`

	// Optional fields
	Name                   *string `json:"name,omitempty"`
	Description            *string `json:"description,omitempty"`
	Domains                *string `json:"domains,omitempty"`
	InstantDeploy          *bool   `json:"instant_deploy,omitempty"`
	GitCommitSHA           *string `json:"git_commit_sha,omitempty"`
	DestinationUUID        *string `json:"destination_uuid,omitempty"`
	BuildCommand           *string `json:"build_command,omitempty"`
	StartCommand           *string `json:"start_command,omitempty"`
	InstallCommand         *string `json:"install_command,omitempty"`
	BaseDirectory          *string `json:"base_directory,omitempty"`
	PublishDirectory       *string `json:"publish_directory,omitempty"`
	PortsMappings          *string `json:"ports_mappings,omitempty"`
	CustomDockerRunOptions *string `json:"custom_docker_run_options,omitempty"`
	CustomLabels           *string `json:"custom_labels,omitempty"`

	// Health checks
	HealthCheckEnabled *bool   `json:"health_check_enabled,omitempty"`
	HealthCheckPath    *string `json:"health_check_path,omitempty"`
	HealthCheckPort    *string `json:"health_check_port,omitempty"`
	HealthCheckMethod  *string `json:"health_check_method,omitempty"`

	// Resource limits
	LimitsCPUs   *string `json:"limits_cpus,omitempty"`
	LimitsMemory *string `json:"limits_memory,omitempty"`
}

// ApplicationCreateGitHubAppRequest for POST /applications/private-github-app
// Creates an application from a private repository using GitHub App authentication
type ApplicationCreateGitHubAppRequest struct {
	// Required fields
	ProjectUUID   string `json:"project_uuid"`
	ServerUUID    string `json:"server_uuid"`
	GitHubAppUUID string `json:"github_app_uuid"`
	GitRepository string `json:"git_repository"`
	GitBranch     string `json:"git_branch"`
	BuildPack     string `json:"build_pack"`
	PortsExposes  string `json:"ports_exposes"`

	// Environment (one of these is required)
	EnvironmentName *string `json:"environment_name,omitempty"`
	EnvironmentUUID *string `json:"environment,omitempty"`

	// Optional fields (same as public)
	Name                   *string `json:"name,omitempty"`
	Description            *string `json:"description,omitempty"`
	Domains                *string `json:"domains,omitempty"`
	InstantDeploy          *bool   `json:"instant_deploy,omitempty"`
	GitCommitSHA           *string `json:"git_commit_sha,omitempty"`
	DestinationUUID        *string `json:"destination_uuid,omitempty"`
	BuildCommand           *string `json:"build_command,omitempty"`
	StartCommand           *string `json:"start_command,omitempty"`
	InstallCommand         *string `json:"install_command,omitempty"`
	BaseDirectory          *string `json:"base_directory,omitempty"`
	PublishDirectory       *string `json:"publish_directory,omitempty"`
	PortsMappings          *string `json:"ports_mappings,omitempty"`
	CustomDockerRunOptions *string `json:"custom_docker_run_options,omitempty"`
	CustomLabels           *string `json:"custom_labels,omitempty"`
	HealthCheckEnabled     *bool   `json:"health_check_enabled,omitempty"`
	HealthCheckPath        *string `json:"health_check_path,omitempty"`
	HealthCheckPort        *string `json:"health_check_port,omitempty"`
	HealthCheckMethod      *string `json:"health_check_method,omitempty"`
	LimitsCPUs             *string `json:"limits_cpus,omitempty"`
	LimitsMemory           *string `json:"limits_memory,omitempty"`
}

// ApplicationCreateDeployKeyRequest for POST /applications/private-deploy-key
// Creates an application from a private repository using SSH deploy key
type ApplicationCreateDeployKeyRequest struct {
	// Required fields
	ProjectUUID    string `json:"project_uuid"`
	ServerUUID     string `json:"server_uuid"`
	PrivateKeyUUID string `json:"private_key_uuid"`
	GitRepository  string `json:"git_repository"`
	GitBranch      string `json:"git_branch"`
	BuildPack      string `json:"build_pack"`
	PortsExposes   string `json:"ports_exposes"`

	// Environment (one of these is required)
	EnvironmentName *string `json:"environment_name,omitempty"`
	EnvironmentUUID *string `json:"environment,omitempty"`

	// Optional fields (same as public)
	Name                   *string `json:"name,omitempty"`
	Description            *string `json:"description,omitempty"`
	Domains                *string `json:"domains,omitempty"`
	InstantDeploy          *bool   `json:"instant_deploy,omitempty"`
	GitCommitSHA           *string `json:"git_commit_sha,omitempty"`
	DestinationUUID        *string `json:"destination_uuid,omitempty"`
	BuildCommand           *string `json:"build_command,omitempty"`
	StartCommand           *string `json:"start_command,omitempty"`
	InstallCommand         *string `json:"install_command,omitempty"`
	BaseDirectory          *string `json:"base_directory,omitempty"`
	PublishDirectory       *string `json:"publish_directory,omitempty"`
	PortsMappings          *string `json:"ports_mappings,omitempty"`
	CustomDockerRunOptions *string `json:"custom_docker_run_options,omitempty"`
	CustomLabels           *string `json:"custom_labels,omitempty"`
	HealthCheckEnabled     *bool   `json:"health_check_enabled,omitempty"`
	HealthCheckPath        *string `json:"health_check_path,omitempty"`
	HealthCheckPort        *string `json:"health_check_port,omitempty"`
	HealthCheckMethod      *string `json:"health_check_method,omitempty"`
	LimitsCPUs             *string `json:"limits_cpus,omitempty"`
	LimitsMemory           *string `json:"limits_memory,omitempty"`
}

// ApplicationCreateDockerfileRequest for POST /applications/dockerfile
// Creates an application from a custom Dockerfile
type ApplicationCreateDockerfileRequest struct {
	// Required fields
	ProjectUUID string `json:"project_uuid"`
	ServerUUID  string `json:"server_uuid"`
	Dockerfile  string `json:"dockerfile"`

	// Environment (one of these is required)
	EnvironmentName *string `json:"environment_name,omitempty"`
	EnvironmentUUID *string `json:"environment,omitempty"`

	// Optional fields
	Name                   *string `json:"name,omitempty"`
	Description            *string `json:"description,omitempty"`
	Domains                *string `json:"domains,omitempty"`
	InstantDeploy          *bool   `json:"instant_deploy,omitempty"`
	DestinationUUID        *string `json:"destination_uuid,omitempty"`
	PortsExposes           *string `json:"ports_exposes,omitempty"`
	PortsMappings          *string `json:"ports_mappings,omitempty"`
	CustomDockerRunOptions *string `json:"custom_docker_run_options,omitempty"`
	CustomLabels           *string `json:"custom_labels,omitempty"`
	HealthCheckEnabled     *bool   `json:"health_check_enabled,omitempty"`
	HealthCheckPath        *string `json:"health_check_path,omitempty"`
	HealthCheckPort        *string `json:"health_check_port,omitempty"`
	HealthCheckMethod      *string `json:"health_check_method,omitempty"`
	LimitsCPUs             *string `json:"limits_cpus,omitempty"`
	LimitsMemory           *string `json:"limits_memory,omitempty"`
}

// ApplicationCreateDockerImageRequest for POST /applications/dockerimage
// Creates an application from a pre-built Docker image
type ApplicationCreateDockerImageRequest struct {
	// Required fields
	ProjectUUID             string `json:"project_uuid"`
	ServerUUID              string `json:"server_uuid"`
	DockerRegistryImageName string `json:"docker_registry_image_name"`
	PortsExposes            string `json:"ports_exposes"`

	// Environment (one of these is required)
	EnvironmentName *string `json:"environment_name,omitempty"`
	EnvironmentUUID *string `json:"environment,omitempty"`

	// Optional fields
	Name                   *string `json:"name,omitempty"`
	Description            *string `json:"description,omitempty"`
	Domains                *string `json:"domains,omitempty"`
	InstantDeploy          *bool   `json:"instant_deploy,omitempty"`
	DestinationUUID        *string `json:"destination_uuid,omitempty"`
	DockerRegistryImageTag *string `json:"docker_registry_image_tag,omitempty"`
	PortsMappings          *string `json:"ports_mappings,omitempty"`
	CustomDockerRunOptions *string `json:"custom_docker_run_options,omitempty"`
	CustomLabels           *string `json:"custom_labels,omitempty"`
	HealthCheckEnabled     *bool   `json:"health_check_enabled,omitempty"`
	HealthCheckPath        *string `json:"health_check_path,omitempty"`
	HealthCheckPort        *string `json:"health_check_port,omitempty"`
	HealthCheckMethod      *string `json:"health_check_method,omitempty"`
	LimitsCPUs             *string `json:"limits_cpus,omitempty"`
	LimitsMemory           *string `json:"limits_memory,omitempty"`
}

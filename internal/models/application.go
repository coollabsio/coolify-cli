package models

// ApplicationSettings represents the settings for a Coolify application
type ApplicationSettings struct {
	ID                          int    `json:"-" table:"-"`
	ApplicationID               int    `json:"-" table:"-"`
	IsStatic                    *bool  `json:"is_static,omitempty"`
	IsBuildServerEnabled        *bool  `json:"is_build_server_enabled,omitempty"`
	IsPreserveRepositoryEnabled *bool  `json:"is_preserve_repository_enabled,omitempty"`
	IsAutoDeployEnabled         *bool  `json:"is_auto_deploy_enabled,omitempty"`
	IsForceHTTPSEnabled         *bool  `json:"is_force_https_enabled,omitempty"`
	IsDebugEnabled              *bool  `json:"is_debug_enabled,omitempty"`
	IsPreviewDeploymentsEnabled *bool  `json:"is_preview_deployments_enabled,omitempty"`
	IsGitSubmodulesEnabled      *bool  `json:"is_git_submodules_enabled,omitempty"`
	IsGitLFSEnabled             *bool  `json:"is_git_lfs_enabled,omitempty"`
	CreatedAt                   string `json:"-" table:"-"`
	UpdatedAt                   string `json:"-" table:"-"`
}

// Application represents a Coolify application
type Application struct {
	// Table-visible fields
	UUID          string  `json:"uuid"`
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	Status        string  `json:"status"`
	FQDN          *string `json:"fqdn,omitempty"`
	GitRepository *string `json:"git_repository,omitempty"`
	GitBranch     *string `json:"git_branch,omitempty"`
	BuildPack     *string `json:"build_pack,omitempty"`
	PortsExposes  *string `json:"ports_exposes,omitempty"`

	// Git extended (JSON-only)
	GitCommitSHA *string `json:"git_commit_sha,omitempty" table:"-"`
	GitFullURL   *string `json:"git_full_url,omitempty" table:"-"`

	// Build configuration (JSON-only)
	InstallCommand   *string `json:"install_command,omitempty" table:"-"`
	BuildCommand     *string `json:"build_command,omitempty" table:"-"`
	StartCommand     *string `json:"start_command,omitempty" table:"-"`
	BaseDirectory    *string `json:"base_directory,omitempty" table:"-"`
	PublishDirectory *string `json:"publish_directory,omitempty" table:"-"`
	StaticImage      *string `json:"static_image,omitempty" table:"-"`

	// Docker configuration (JSON-only)
	Dockerfile               *string `json:"dockerfile,omitempty" table:"-"`
	DockerfileLocation       *string `json:"dockerfile_location,omitempty" table:"-"`
	DockerRegistryImageName  *string `json:"docker_registry_image_name,omitempty" table:"-"`
	DockerRegistryImageTag   *string `json:"docker_registry_image_tag,omitempty" table:"-"`
	DockerCompose            *string `json:"docker_compose,omitempty" table:"-"`
	DockerComposeRaw         *string `json:"docker_compose_raw,omitempty" table:"-"`
	DockerComposeLocation    *string `json:"docker_compose_location,omitempty" table:"-"`
	CustomDockerRunOptions   *string `json:"custom_docker_run_options,omitempty" table:"-"`
	CustomLabels             *string `json:"custom_labels,omitempty" table:"-"`
	CustomNginxConfiguration *string `json:"custom_nginx_configuration,omitempty" table:"-"`

	// Networking (JSON-only)
	PortsMappings      *string `json:"ports_mappings,omitempty" table:"-"`
	Domains            *string `json:"domains,omitempty" table:"-"`
	Redirect           *string `json:"redirect,omitempty" table:"-"`
	PreviewURLTemplate *string `json:"preview_url_template,omitempty" table:"-"`

	// Health checks (JSON-only)
	HealthCheckEnabled      *bool   `json:"health_check_enabled,omitempty" table:"-"`
	HealthCheckPath         *string `json:"health_check_path,omitempty" table:"-"`
	HealthCheckPort         *string `json:"health_check_port,omitempty" table:"-"`
	HealthCheckHost         *string `json:"health_check_host,omitempty" table:"-"`
	HealthCheckMethod       *string `json:"health_check_method,omitempty" table:"-"`
	HealthCheckScheme       *string `json:"health_check_scheme,omitempty" table:"-"`
	HealthCheckReturnCode   *int    `json:"health_check_return_code,omitempty" table:"-"`
	HealthCheckResponseText *string `json:"health_check_response_text,omitempty" table:"-"`
	HealthCheckInterval     *int    `json:"health_check_interval,omitempty" table:"-"`
	HealthCheckTimeout      *int    `json:"health_check_timeout,omitempty" table:"-"`
	HealthCheckRetries      *int    `json:"health_check_retries,omitempty" table:"-"`
	HealthCheckStartPeriod  *int    `json:"health_check_start_period,omitempty" table:"-"`

	// Resource limits (JSON-only)
	LimitsCPUs              *string `json:"limits_cpus,omitempty" table:"-"`
	LimitsCPUShares         *int    `json:"limits_cpu_shares,omitempty" table:"-"`
	LimitsCPUSet            *string `json:"limits_cpuset,omitempty" table:"-"`
	LimitsMemory            *string `json:"limits_memory,omitempty" table:"-"`
	LimitsMemoryReservation *string `json:"limits_memory_reservation,omitempty" table:"-"`
	LimitsMemorySwap        *string `json:"limits_memory_swap,omitempty" table:"-"`
	LimitsMemorySwappiness  *int    `json:"limits_memory_swappiness,omitempty" table:"-"`

	// Deployment hooks (JSON-only)
	PreDeploymentCommand           *string `json:"pre_deployment_command,omitempty" table:"-"`
	PreDeploymentCommandContainer  *string `json:"pre_deployment_command_container,omitempty" table:"-"`
	PostDeploymentCommand          *string `json:"post_deployment_command,omitempty" table:"-"`
	PostDeploymentCommandContainer *string `json:"post_deployment_command_container,omitempty" table:"-"`

	// Webhook secrets (JSON-only)
	ManualWebhookSecretGitHub    *string `json:"manual_webhook_secret_github,omitempty" table:"-" sensitive:"true"`
	ManualWebhookSecretGitLab    *string `json:"manual_webhook_secret_gitlab,omitempty" table:"-" sensitive:"true"`
	ManualWebhookSecretBitbucket *string `json:"manual_webhook_secret_bitbucket,omitempty" table:"-" sensitive:"true"`
	ManualWebhookSecretGitea     *string `json:"manual_webhook_secret_gitea,omitempty" table:"-" sensitive:"true"`

	// Misc (JSON-only)
	WatchPaths    *string `json:"watch_paths,omitempty" table:"-"`
	SwarmReplicas *int    `json:"swarm_replicas,omitempty" table:"-"`
	ConfigHash    *string `json:"config_hash,omitempty" table:"-"`

	// Nested settings (JSON-only)
	Settings *ApplicationSettings `json:"settings,omitempty" table:"-"`

	// Hidden fields (not in JSON or table output)
	ID            int    `json:"-" table:"-"`
	EnvironmentID *int   `json:"-" table:"-"`
	DestinationID *int   `json:"-" table:"-"`
	SourceID      *int   `json:"-" table:"-"`
	PrivateKeyID  *int   `json:"-" table:"-"`
	CreatedAt     string `json:"-" table:"-"`
	UpdatedAt     string `json:"-" table:"-"`
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
	DockerfileTargetBuild   *string `json:"dockerfile_target_build,omitempty"`
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
	Comment        *string `json:"comment,omitempty"`
	RealValue      *string `json:"real_value,omitempty" sensitive:"true"`
	ApplicationID  *int    `json:"-" table:"-"`
	CreatedAt      string  `json:"-" table:"-"`
	UpdatedAt      string  `json:"-" table:"-"`
}

// EnvironmentVariableCreateRequest represents the request to create an environment variable
type EnvironmentVariableCreateRequest struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	IsBuildTime *bool   `json:"is_build_time,omitempty"`
	IsPreview   *bool   `json:"is_preview,omitempty"`
	IsLiteral   *bool   `json:"is_literal,omitempty"`
	IsMultiline *bool   `json:"is_multiline,omitempty"`
	IsRuntime   *bool   `json:"is_runtime,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

// EnvironmentVariableUpdateRequest represents the request to update an environment variable
type EnvironmentVariableUpdateRequest struct {
	Key         *string `json:"key,omitempty"`
	Value       *string `json:"value,omitempty"`
	IsBuildTime *bool   `json:"is_build_time,omitempty"`
	IsPreview   *bool   `json:"is_preview,omitempty"`
	IsLiteral   *bool   `json:"is_literal,omitempty"`
	IsMultiline *bool   `json:"is_multiline,omitempty"`
	IsRuntime   *bool   `json:"is_runtime,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

// StoragesResponse represents the API response for listing storages
type StoragesResponse struct {
	PersistentStorages []PersistentStorage `json:"persistent_storages"`
	FileStorages       []FileStorage       `json:"file_storages"`
}

// PersistentStorage represents a persistent volume for an application
type PersistentStorage struct {
	ID                     int     `json:"id" table:"-"`
	UUID                   string  `json:"uuid"`
	Name                   string  `json:"name"`
	MountPath              string  `json:"mount_path"`
	HostPath               *string `json:"host_path,omitempty"`
	IsPreviewSuffixEnabled bool    `json:"is_preview_suffix_enabled"`
	IsReadOnly             bool    `json:"is_readonly"`
	ResourceType           string  `json:"resource_type" table:"-"`
	ResourceID             int     `json:"resource_id" table:"-"`
}

// FileStorage represents a file storage for an application
type FileStorage struct {
	ID                     int     `json:"id"`
	UUID                   string  `json:"uuid"`
	FsPath                 string  `json:"fs_path"`
	MountPath              string  `json:"mount_path"`
	Content                *string `json:"content,omitempty"`
	IsDirectory            bool    `json:"is_directory"`
	IsBasedOnGit           bool    `json:"is_based_on_git"`
	IsPreviewSuffixEnabled bool    `json:"is_preview_suffix_enabled"`
	Chown                  *string `json:"chown,omitempty"`
	Chmod                  *string `json:"chmod,omitempty"`
	ResourceType           string  `json:"resource_type" table:"-"`
	ResourceID             int     `json:"resource_id" table:"-"`
}

// StorageListItem is a unified view of both storage types for table output
type StorageListItem struct {
	ID                     int    `json:"id" table:"-"`
	UUID                   string `json:"uuid"`
	Type                   string `json:"type"`
	Name                   string `json:"name"`
	MountPath              string `json:"mount_path"`
	HostPath               string `json:"host_path,omitempty"`
	IsPreviewSuffixEnabled bool   `json:"is_preview_suffix_enabled"`
	Content                string `json:"content,omitempty" table:"-"`
}

// MergeStorages converts a StoragesResponse into a unified list of StorageListItem
func MergeStorages(resp StoragesResponse) []StorageListItem {
	var items []StorageListItem
	for _, ps := range resp.PersistentStorages {
		hostPath := ""
		if ps.HostPath != nil {
			hostPath = *ps.HostPath
		}
		items = append(items, StorageListItem{
			ID:                     ps.ID,
			UUID:                   ps.UUID,
			Type:                   "persistent",
			Name:                   ps.Name,
			MountPath:              ps.MountPath,
			HostPath:               hostPath,
			IsPreviewSuffixEnabled: ps.IsPreviewSuffixEnabled,
		})
	}
	for _, fs := range resp.FileStorages {
		content := ""
		if fs.Content != nil {
			content = *fs.Content
		}
		items = append(items, StorageListItem{
			ID:                     fs.ID,
			UUID:                   fs.UUID,
			Type:                   "file",
			Name:                   fs.FsPath,
			MountPath:              fs.MountPath,
			IsPreviewSuffixEnabled: fs.IsPreviewSuffixEnabled,
			Content:                content,
		})
	}
	return items
}

// StorageCreateRequest represents the request to create a storage for applications and databases
type StorageCreateRequest struct {
	Type        string  `json:"type"`       // "persistent" or "file"
	MountPath   string  `json:"mount_path"` // required
	Name        *string `json:"name,omitempty"`
	HostPath    *string `json:"host_path,omitempty"`
	Content     *string `json:"content,omitempty"`
	IsDirectory *bool   `json:"is_directory,omitempty"`
	FsPath      *string `json:"fs_path,omitempty"`
}

// ServiceStorageCreateRequest represents the request to create a storage for services
// Services require resource_uuid to identify which sub-resource the storage belongs to
type ServiceStorageCreateRequest struct {
	Type         string  `json:"type"`          // "persistent" or "file"
	MountPath    string  `json:"mount_path"`    // required
	ResourceUUID string  `json:"resource_uuid"` // required for services
	Name         *string `json:"name,omitempty"`
	HostPath     *string `json:"host_path,omitempty"`
	Content      *string `json:"content,omitempty"`
	IsDirectory  *bool   `json:"is_directory,omitempty"`
	FsPath       *string `json:"fs_path,omitempty"`
}

// StorageUpdateRequest represents the request to update a storage
type StorageUpdateRequest struct {
	// Required fields
	Type string `json:"type"` // "persistent" or "file"

	// Identifier (uuid preferred, id deprecated)
	UUID *string `json:"uuid,omitempty"`
	ID   *int    `json:"id,omitempty"`

	// Optional fields
	IsPreviewSuffixEnabled *bool   `json:"is_preview_suffix_enabled,omitempty"`
	Name                   *string `json:"name,omitempty"`
	MountPath              *string `json:"mount_path,omitempty"`
	HostPath               *string `json:"host_path,omitempty"`
	Content                *string `json:"content,omitempty"`
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
	DockerfileTargetBuild  *string `json:"dockerfile_target_build,omitempty"`

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
	DockerfileTargetBuild  *string `json:"dockerfile_target_build,omitempty"`
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
	DockerfileTargetBuild  *string `json:"dockerfile_target_build,omitempty"`
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
	DockerfileTargetBuild  *string `json:"dockerfile_target_build,omitempty"`
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
	DockerfileTargetBuild  *string `json:"dockerfile_target_build,omitempty"`
	HealthCheckEnabled     *bool   `json:"health_check_enabled,omitempty"`
	HealthCheckPath        *string `json:"health_check_path,omitempty"`
	HealthCheckPort        *string `json:"health_check_port,omitempty"`
	HealthCheckMethod      *string `json:"health_check_method,omitempty"`
	LimitsCPUs             *string `json:"limits_cpus,omitempty"`
	LimitsMemory           *string `json:"limits_memory,omitempty"`
}

package models

// Server represents a Coolify server
type Server struct {
	ID       int      `json:"-" table:"-"`
	UUID     string   `json:"uuid"`
	Name     string   `json:"name"`
	IP       string   `json:"ip" sensitive:"true"`
	User     string   `json:"user" sensitive:"true"`
	Port     int      `json:"port" sensitive:"true"`
	Settings Settings `json:"settings" table:"-"`
}

// Settings for server
type Settings struct {
	IsReachable bool `json:"is_reachable"`
	IsUsable    bool `json:"is_usable"`
}

// ServerCreateRequest for creating servers
type ServerCreateRequest struct {
	Name            string `json:"name"`
	IP              string `json:"ip"`
	Port            int    `json:"port"`
	User            string `json:"user"`
	PrivateKeyUUID  string `json:"private_key_uuid"`
	InstantValidate bool   `json:"instant_validate"`
}

// ServerUpdateRequest for patching servers. Only non-nil fields are sent.
type ServerUpdateRequest struct {
	Name                                   *string `json:"name,omitempty"`
	Description                            *string `json:"description,omitempty"`
	IP                                     *string `json:"ip,omitempty"`
	Port                                   *int    `json:"port,omitempty"`
	User                                   *string `json:"user,omitempty"`
	PrivateKeyUUID                         *string `json:"private_key_uuid,omitempty"`
	IsBuildServer                          *bool   `json:"is_build_server,omitempty"`
	InstantValidate                        *bool   `json:"instant_validate,omitempty"`
	ProxyType                              *string `json:"proxy_type,omitempty"`
	ConcurrentBuilds                       *int    `json:"concurrent_builds,omitempty"`
	DynamicTimeout                         *int    `json:"dynamic_timeout,omitempty"`
	DeploymentQueueLimit                   *int    `json:"deployment_queue_limit,omitempty"`
	ServerDiskUsageNotificationThreshold   *int    `json:"server_disk_usage_notification_threshold,omitempty"`
	ServerDiskUsageCheckFrequency          *string `json:"server_disk_usage_check_frequency,omitempty"`
	ConnectionTimeout                      *int    `json:"connection_timeout,omitempty"`
	IsTerminalEnabled                      *bool   `json:"is_terminal_enabled,omitempty"`
}

// ServerValidationRequest configures server validation behavior.
type ServerValidationRequest struct {
	Install bool `json:"install"`
}

// Domain represents a domain configuration
type Domain struct {
	IP      string   `json:"ip"`
	Domains []string `json:"domains"`
}

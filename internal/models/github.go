package models

// GitHubApp represents a GitHub App integration
type GitHubApp struct {
	ID             int     `json:"id" table:"-"`
	UUID           string  `json:"uuid"`
	Name           string  `json:"name"`
	Organization   *string `json:"organization,omitempty"`
	APIURL         string  `json:"api_url"`
	HTMLURL        string  `json:"html_url"`
	CustomUser     string  `json:"custom_user"`
	CustomPort     int     `json:"custom_port"`
	AppID          int     `json:"app_id" table:"-"`
	InstallationID int     `json:"installation_id" table:"-"`
	ClientID       string  `json:"client_id" table:"-"`
	PrivateKeyID   int     `json:"private_key_id" table:"-"`
	IsSystemWide   bool    `json:"is_system_wide" table:"-"`
	TeamID         int     `json:"team_id" table:"-"`
}

// GitHubAppCreateRequest represents a request to create a GitHub App
type GitHubAppCreateRequest struct {
	Name           string  `json:"name"`
	Organization   *string `json:"organization,omitempty"`
	APIURL         string  `json:"api_url"`
	HTMLURL        string  `json:"html_url"`
	CustomUser     *string `json:"custom_user,omitempty"`
	CustomPort     *int    `json:"custom_port,omitempty"`
	AppID          int     `json:"app_id"`
	InstallationID int     `json:"installation_id"`
	ClientID       string  `json:"client_id"`
	ClientSecret   string  `json:"client_secret" sensitive:"true"`
	WebhookSecret  *string `json:"webhook_secret,omitempty" sensitive:"true"`
	PrivateKeyUUID string  `json:"private_key_uuid"`
	IsSystemWide   *bool   `json:"is_system_wide,omitempty"`
}

// GitHubAppUpdateRequest represents a request to update a GitHub App
type GitHubAppUpdateRequest struct {
	Name           *string `json:"name,omitempty"`
	Organization   *string `json:"organization,omitempty"`
	APIURL         *string `json:"api_url,omitempty"`
	HTMLURL        *string `json:"html_url,omitempty"`
	CustomUser     *string `json:"custom_user,omitempty"`
	CustomPort     *int    `json:"custom_port,omitempty"`
	AppID          *int    `json:"app_id,omitempty"`
	InstallationID *int    `json:"installation_id,omitempty"`
	ClientID       *string `json:"client_id,omitempty"`
	ClientSecret   *string `json:"client_secret,omitempty" sensitive:"true"`
	WebhookSecret  *string `json:"webhook_secret,omitempty" sensitive:"true"`
	PrivateKeyUUID *string `json:"private_key_uuid,omitempty"`
	IsSystemWide   *bool   `json:"is_system_wide,omitempty"`
}

// GitHubRepository represents a repository from GitHub
type GitHubRepository struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	HTMLURL  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
}

// GitHubBranch represents a branch from GitHub
type GitHubBranch struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
}

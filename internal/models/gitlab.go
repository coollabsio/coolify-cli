package models

// GitLabApp represents a GitLab App (OAuth source) integration
type GitLabApp struct {
	ID           int     `json:"id"`
	UUID         string  `json:"uuid"`
	Name         string  `json:"name"`
	APIURL       string  `json:"api_url"`
	HTMLURL      string  `json:"html_url"`
	CustomUser   string  `json:"custom_user"`
	CustomPort   int     `json:"custom_port"`
	ClientID     *string `json:"client_id,omitempty"`
	GroupName    *string `json:"group_name,omitempty"`
	RedirectURI  *string `json:"redirect_uri,omitempty"`
	IsSystemWide bool    `json:"is_system_wide" table:"-"`
	IsPublic     bool    `json:"is_public" table:"-"`
	TeamID       int     `json:"team_id" table:"-"`
}

// GitLabAppCreateRequest represents a request to create a GitLab App
type GitLabAppCreateRequest struct {
	Name         string  `json:"name"`
	HTMLURL      string  `json:"html_url"`
	APIURL       *string `json:"api_url,omitempty"`
	CustomUser   *string `json:"custom_user,omitempty"`
	CustomPort   *int    `json:"custom_port,omitempty"`
	GroupName    *string `json:"group_name,omitempty"`
	ClientID     *string `json:"client_id,omitempty"`
	ClientSecret *string `json:"client_secret,omitempty" sensitive:"true"`
	WebhookToken *string `json:"webhook_token,omitempty" sensitive:"true"`
	RedirectURI  *string `json:"redirect_uri,omitempty"`
	IsSystemWide *bool   `json:"is_system_wide,omitempty"`
}

// GitLabAppUpdateRequest represents a request to update a GitLab App
type GitLabAppUpdateRequest struct {
	Name         *string `json:"name,omitempty"`
	HTMLURL      *string `json:"html_url,omitempty"`
	APIURL       *string `json:"api_url,omitempty"`
	CustomUser   *string `json:"custom_user,omitempty"`
	CustomPort   *int    `json:"custom_port,omitempty"`
	GroupName    *string `json:"group_name,omitempty"`
	ClientID     *string `json:"client_id,omitempty"`
	ClientSecret *string `json:"client_secret,omitempty" sensitive:"true"`
	WebhookToken *string `json:"webhook_token,omitempty" sensitive:"true"`
	RedirectURI  *string `json:"redirect_uri,omitempty"`
	IsSystemWide *bool   `json:"is_system_wide,omitempty"`
}

// GitLabAppUpdateResponse is the API response for PATCH /gitlab-apps/{id}
type GitLabAppUpdateResponse struct {
	Message string    `json:"message"`
	Data    GitLabApp `json:"data"`
}

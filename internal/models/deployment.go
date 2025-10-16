package models

// Deployment represents a deployment operation
type Deployment struct {
	ID              int     `json:"id" table:"-"`
	UUID            string  `json:"deployment_uuid"`
	ApplicationID   *string `json:"application_id,omitempty" table:"-"`
	ApplicationName *string `json:"application_name,omitempty"`
	ServerName      *string `json:"server_name,omitempty"`
	Status          string  `json:"status"`
	Commit          *string `json:"commit,omitempty"`
	CommitMessage   *string `json:"commit_message,omitempty" table:"-"`
	// Additional fields from API that we want to ignore
	DeploymentURL *string `json:"deployment_url,omitempty" table:"-"`
	FinishedAt    *string `json:"finished_at,omitempty" table:"-"`
	Logs          *string `json:"logs,omitempty" table:"-"`
	CreatedAt     *string `json:"created_at,omitempty" table:"-"`
	UpdatedAt     *string `json:"updated_at,omitempty" table:"-"`
}

// DeployResponse wraps deployment trigger responses
type DeployResponse struct {
	Message        string `json:"message"`
	DeploymentUUID string `json:"deployment_uuid,omitempty"`
}

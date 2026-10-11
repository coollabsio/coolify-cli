package models

// Resource represents any deployable resource
type Resource struct {
	ID     int    `json:"-" table:"-"`
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// ResourceStatus constants
const (
	ResourceStatusRunning = "running"
	ResourceStatusStopped = "stopped"
	ResourceStatusError   = "error"
)

// ServerResource is a resource deployed on a server (GET /servers/{uuid}/resources)
type ServerResource struct {
	ID        int    `json:"id" table:"-"`
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at" table:"-"`
	UpdatedAt string `json:"updated_at" table:"-"`
}

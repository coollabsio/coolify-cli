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

// Resources wraps a list of resources
type Resources struct {
	Resources []Resource `json:"resources"`
}

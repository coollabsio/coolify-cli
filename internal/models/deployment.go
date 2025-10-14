package models

// Deployment represents a deployment operation
type Deployment struct {
	Message        string `json:"message"`
	ResourceUUID   string `json:"resource_uuid"`
	DeploymentUUID string `json:"deployment_uuid"`
}

// DeployResponse wraps deployment responses
type DeployResponse struct {
	Deployments []Deployment `json:"deployments"`
}

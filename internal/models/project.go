package models

// Project represents a Coolify project
type Project struct {
	UUID         string        `json:"uuid"`
	Name         string        `json:"name"`
	Description  *string       `json:"description,omitempty"`
	Environments []Environment `json:"environments,omitempty"`
}

// Environment within a project
type Environment struct {
	ID           int           `json:"id"`
	UUID         string        `json:"uuid"`
	Name         string        `json:"name"`
	Description  *string       `json:"description,omitempty"`
	Applications []Application `json:"applications,omitempty"`
	CreatedAt    string        `json:"created_at,omitempty"`
	UpdatedAt    string        `json:"updated_at,omitempty"`
}

// Application within an environment
type Application struct {
	ID          int     `json:"id"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
}

// ProjectCreateRequest for creating projects
type ProjectCreateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

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
	ID           int                    `json:"-" table:"-"`
	UUID         string                 `json:"uuid"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description,omitempty"`
	Applications []ApplicationInProject `json:"applications,omitempty"`
	CreatedAt    string                 `json:"-" table:"-"`
	UpdatedAt    string                 `json:"-" table:"-"`
}

// ApplicationInProject represents a simplified application within an environment
type ApplicationInProject struct {
	ID          int     `json:"-" table:"-"`
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

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

// ProjectUpdateRequest for patching a project
type ProjectUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// EnvironmentCreateRequest for creating an environment in a project
type EnvironmentCreateRequest struct {
	Name string `json:"name"`
}

// EnvironmentDetails is an environment with its resources
// (GET /projects/{uuid}/{environment_name_or_uuid}).
type EnvironmentDetails struct {
	ID           int                   `json:"id" table:"-"`
	UUID         string                `json:"uuid"`
	Name         string                `json:"name"`
	Description  *string               `json:"description,omitempty"`
	ProjectID    int                   `json:"project_id" table:"-"`
	CreatedAt    string                `json:"created_at" table:"-"`
	UpdatedAt    string                `json:"updated_at" table:"-"`
	Applications []EnvironmentResource `json:"applications" table:"-"`
	Postgresqls  []EnvironmentResource `json:"postgresqls" table:"-"`
	Redis        []EnvironmentResource `json:"redis" table:"-"`
	Mongodbs     []EnvironmentResource `json:"mongodbs" table:"-"`
	Mysqls       []EnvironmentResource `json:"mysqls" table:"-"`
	Mariadbs     []EnvironmentResource `json:"mariadbs" table:"-"`
	Services     []EnvironmentResource `json:"services" table:"-"`
}

// EnvironmentResource is the summary of a resource inside an environment.
type EnvironmentResource struct {
	ID          int     `json:"id" table:"-"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
}

// EnvironmentResourceRow is a flattened, typed resource row for table output.
type EnvironmentResourceRow struct {
	Type   string `json:"type"`
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ResourceRows flattens every resource of the environment into typed rows.
func (e *EnvironmentDetails) ResourceRows() []EnvironmentResourceRow {
	groups := []struct {
		kind  string
		items []EnvironmentResource
	}{
		{"application", e.Applications},
		{"postgresql", e.Postgresqls},
		{"redis", e.Redis},
		{"mongodb", e.Mongodbs},
		{"mysql", e.Mysqls},
		{"mariadb", e.Mariadbs},
		{"service", e.Services},
	}
	rows := []EnvironmentResourceRow{}
	for _, group := range groups {
		for _, item := range group.items {
			rows = append(rows, EnvironmentResourceRow{Type: group.kind, UUID: item.UUID, Name: item.Name, Status: item.Status})
		}
	}
	return rows
}

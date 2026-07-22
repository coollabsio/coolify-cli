package models

// ScheduledTask represents a Coolify application or service scheduled task.
type ScheduledTask struct {
	ID        int     `json:"id,omitempty" table:"-"`
	UUID      string  `json:"uuid"`
	Enabled   bool    `json:"enabled"`
	Name      string  `json:"name"`
	Command   string  `json:"command"`
	Frequency string  `json:"frequency"`
	Container *string `json:"container,omitempty"`
	Timeout   int     `json:"timeout"`
	CreatedAt string  `json:"created_at" table:"-"`
	UpdatedAt string  `json:"updated_at" table:"-"`
}

// ScheduledTaskExecution represents a single run of a scheduled task.
type ScheduledTaskExecution struct {
	UUID       string   `json:"uuid"`
	Status     string   `json:"status"`
	Message    *string  `json:"message,omitempty"`
	RetryCount int      `json:"retry_count"`
	Duration   *float64 `json:"duration,omitempty"`
	StartedAt  *string  `json:"started_at,omitempty"`
	FinishedAt *string  `json:"finished_at,omitempty"`
	CreatedAt  string   `json:"created_at" table:"-"`
	UpdatedAt  string   `json:"updated_at" table:"-"`
}

// ScheduledTaskCreateRequest is the body for creating a scheduled task.
type ScheduledTaskCreateRequest struct {
	Name      string  `json:"name"`
	Command   string  `json:"command"`
	Frequency string  `json:"frequency"`
	Container *string `json:"container,omitempty"`
	Timeout   *int    `json:"timeout,omitempty"`
	Enabled   *bool   `json:"enabled,omitempty"`
}

// ScheduledTaskUpdateRequest is the body for updating a scheduled task.
// All fields are optional; the API requires at least one field.
type ScheduledTaskUpdateRequest struct {
	Name      *string `json:"name,omitempty"`
	Command   *string `json:"command,omitempty"`
	Frequency *string `json:"frequency,omitempty"`
	Container *string `json:"container,omitempty"`
	Timeout   *int    `json:"timeout,omitempty"`
	Enabled   *bool   `json:"enabled,omitempty"`
}

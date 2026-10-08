package models

// AuditEvent is an authenticated team activity record.
type AuditEvent struct {
	ID             int            `json:"id" table:"-"`
	Event          string         `json:"event"`
	Source         string         `json:"source"`
	Action         string         `json:"action"`
	ActorName      string         `json:"actor_name" table:"actor"`
	ActorEmail     string         `json:"actor_email" table:"actor_email"`
	ActorTokenName string         `json:"actor_token_name" table:"api_token"`
	ResourceType   string         `json:"resource_type"`
	ResourceUUID   string         `json:"resource_uuid"`
	ResourceName   string         `json:"resource_name"`
	Description    string         `json:"description"`
	Metadata       map[string]any `json:"metadata" table:"-"`
	IPAddress      string         `json:"ip_address" table:"-"`
	UserAgent      string         `json:"user_agent" table:"-"`
	CreatedAt      string         `json:"created_at" table:"created_at"`
}

// AuditEventsPage is a page returned by the audit-events API.
type AuditEventsPage struct {
	CurrentPage int          `json:"current_page"`
	Data        []AuditEvent `json:"data"`
	LastPage    int          `json:"last_page"`
	PerPage     int          `json:"per_page"`
	Total       int          `json:"total"`
}

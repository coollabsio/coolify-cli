package models

// AuditEventSources lists the accepted values of the audit-events "source" filter.
var AuditEventSources = []string{"all", "ui", "api", "mcp", "webhook", "system", "scheduler"}

// AuditEvent is one entry of the team audit log (GET /audit-events).
// Sensitive fields are only returned for tokens with read:sensitive or root.
type AuditEvent struct {
	ID           int     `json:"id" table:"-"`
	TeamID       *int    `json:"team_id,omitempty" table:"-"`
	CreatedAt    string  `json:"created_at"`
	Event        string  `json:"event" table:"-"`
	Source       string  `json:"source"`
	Action       string  `json:"action"`
	Level        string  `json:"level"`
	ActorType    *string `json:"actor_type,omitempty" table:"-"`
	ActorID      *int    `json:"actor_id,omitempty" table:"-"`
	ActorName    *string `json:"actor_name,omitempty"`
	ResourceType *string `json:"resource_type,omitempty"`
	ResourceUUID *string `json:"resource_uuid,omitempty" table:"-"`
	ResourceName *string `json:"resource_name,omitempty"`
	Description  *string `json:"description,omitempty"`

	ActorEmail     *string `json:"actor_email,omitempty" table:"-" sensitive:"true"`
	ActorTokenID   *int    `json:"actor_token_id,omitempty" table:"-"`
	ActorTokenName *string `json:"actor_token_name,omitempty" table:"-"`
	Metadata       any     `json:"metadata,omitempty" table:"-"`
	Changes        any     `json:"changes,omitempty" table:"-"`
	IPAddress      *string `json:"ip_address,omitempty" table:"-" sensitive:"true"`
	UserAgent      *string `json:"user_agent,omitempty" table:"-"`
}

// AuditEventPage is a page of audit events, newest first.
type AuditEventPage struct {
	CurrentPage int          `json:"current_page"`
	Data        []AuditEvent `json:"data"`
	PerPage     int          `json:"per_page"`
	LastPage    int          `json:"last_page"`
	Total       int          `json:"total"`
}

// AuditEventListOptions holds the query filters of GET /audit-events.
// Zero values are omitted from the query string.
type AuditEventListOptions struct {
	PerPage int
	Page    int
	Search  string
	Action  string
	Source  string
}

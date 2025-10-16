package models

// Team represents a Coolify team
type Team struct {
	ID          int     `json:"-" table:"-"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	CreatedAt   string  `json:"-" table:"-"`
	UpdatedAt   string  `json:"-" table:"-"`
}

// TeamMember represents a member of a team
type TeamMember struct {
	ID        int     `json:"-" table:"-"`
	UUID      string  `json:"uuid" table:"-"`
	Name      string  `json:"name"`
	Email     string  `json:"email" sensitive:"true"`
	Role      *string `json:"role,omitempty" table:"-"`
	CreatedAt string  `json:"-" table:"-"`
	UpdatedAt string  `json:"-" table:"-"`
}

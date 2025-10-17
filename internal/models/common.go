package models

// Response wraps common API response fields
type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	UUID    string `json:"uuid,omitempty"`
}

// UUID is a common UUID field
type UUID struct {
	UUID string `json:"uuid"`
}

// Timestamps for created/updated times
type Timestamps struct {
	CreatedAt string `json:"-" table:"-"`
	UpdatedAt string `json:"-" table:"-"`
}

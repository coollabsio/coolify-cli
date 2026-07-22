package models

// PrivateKey represents an SSH private key
type PrivateKey struct {
	ID          int     `json:"-" table:"-"`
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	PublicKey   string  `json:"public_key" sensitive:"true"`
	PrivateKey  string  `json:"private_key" sensitive:"true"`
}

// PrivateKeyCreateRequest for creating keys
type PrivateKeyCreateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	PrivateKey  string  `json:"private_key"`
}

// PrivateKeyUpdateRequest for updating keys
type PrivateKeyUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	PrivateKey  string  `json:"private_key"`
}

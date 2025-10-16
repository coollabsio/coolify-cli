package models

// PrivateKey represents an SSH private key
type PrivateKey struct {
	ID         int    `json:"-" table:"-"`
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	PublicKey  string `json:"public_key" sensitive:"true"`
	PrivateKey string `json:"private_key" sensitive:"true"`
}

// PrivateKeyCreateRequest for creating keys
type PrivateKeyCreateRequest struct {
	Name       string `json:"name"`
	PrivateKey string `json:"private_key"`
}

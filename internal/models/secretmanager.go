package models

type SecretManagerMetadata struct {
	BaseURL   string `json:"base_url,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type SecretManagerTokenCreateRequest struct {
	Provider string                `json:"provider"`
	Name     string                `json:"name"`
	Token    string                `json:"token" sensitive:"true" table:"-"`
	Metadata SecretManagerMetadata `json:"metadata,omitempty"`
}

type ApplicationSecretManagerRequest struct {
	IntegrationTokenUUID string            `json:"integration_token_uuid"`
	Settings             map[string]string `json:"settings,omitempty"`
}

type ApplicationSecretManager struct {
	IntegrationTokenUUID string            `json:"integration_token_uuid"`
	Provider             string            `json:"provider"`
	Settings             map[string]string `json:"settings,omitempty" table:"-"`
}

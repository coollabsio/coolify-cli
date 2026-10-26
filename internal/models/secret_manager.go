package models

// IntegrationTokenProviders lists the secret manager providers accepted by
// POST /security/integration-tokens.
var IntegrationTokenProviders = []string{"doppler", "infisical", "vault"}

// IntegrationTokenCreateRequest is the body for POST /security/integration-tokens.
type IntegrationTokenCreateRequest struct {
	Provider string         `json:"provider"`
	Name     string         `json:"name"`
	Token    string         `json:"token"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ApplicationSecretManagerRequest is the body for PATCH /applications/{uuid}/secret-manager.
type ApplicationSecretManagerRequest struct {
	IntegrationTokenUUID string         `json:"integration_token_uuid"`
	Settings             map[string]any `json:"settings,omitempty"`
}

// ApplicationSecretManager is the secret manager configuration of an application.
type ApplicationSecretManager struct {
	IntegrationTokenUUID string         `json:"integration_token_uuid"`
	Provider             string         `json:"provider"`
	Settings             map[string]any `json:"settings"`
}

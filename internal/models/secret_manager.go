package models

// IntegrationTokenProviders lists the providers accepted by POST /security/integration-tokens.
var IntegrationTokenProviders = []string{"cloudflare", "doppler", "infisical", "vault"}

// IntegrationTokenCreateRequest is the body for POST /security/integration-tokens.
type IntegrationTokenCreateRequest struct {
	Provider     string         `json:"provider"`
	Name         string         `json:"name"`
	Token        string         `json:"token"`
	Capabilities []string       `json:"capabilities,omitempty"`
	AutomaticDNS *bool          `json:"automatic_dns,omitempty"`
	ProxiedDNS   *bool          `json:"proxied_dns,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// IntegrationTokenUpdateRequest is the body for PATCH /security/integration-tokens/{uuid}.
// Omitted fields keep their value.
type IntegrationTokenUpdateRequest struct {
	Name         *string        `json:"name,omitempty"`
	Token        string         `json:"token,omitempty"`
	Capabilities []string       `json:"capabilities,omitempty"`
	AutomaticDNS *bool          `json:"automatic_dns,omitempty"`
	ProxiedDNS   *bool          `json:"proxied_dns,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// IntegrationToken is an integration token as returned by the API. The secret is never returned.
type IntegrationToken struct {
	UUID          string   `json:"uuid"`
	Name          string   `json:"name"`
	Provider      string   `json:"provider"`
	Capabilities  []string `json:"capabilities"`
	AutomaticDNS  *bool    `json:"automatic_dns,omitempty"`
	ProxiedDNS    *bool    `json:"proxied_dns,omitempty"`
	DNSZonesCount *int     `json:"dns_zones_count,omitempty"`
	ZonesSyncedAt *string  `json:"zones_synced_at,omitempty" table:"-"`
	CreatedAt     string   `json:"created_at" table:"-"`
	UpdatedAt     string   `json:"updated_at" table:"-"`
}

// ZoneSyncResult is the response of POST /security/integration-tokens/{uuid}/sync-zones.
type ZoneSyncResult struct {
	ZonesCount int `json:"zones_count"`
}

// ApplicationSecretManagerRequest is the body for PATCH /{resource_type}/{uuid}/secret-manager.
type ApplicationSecretManagerRequest struct {
	IntegrationTokenUUID string         `json:"integration_token_uuid"`
	Settings             map[string]any `json:"settings,omitempty"`
}

// ApplicationSecretManager is the response of PATCH /{resource_type}/{uuid}/secret-manager.
type ApplicationSecretManager struct {
	IntegrationTokenUUID string         `json:"integration_token_uuid"`
	Provider             string         `json:"provider"`
	Settings             map[string]any `json:"settings"`
}

// SecretManagerTokenOption is a secret manager token that a resource can link.
type SecretManagerTokenOption struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}

// SecretManagerSource is the response of GET /{resource_type}/{uuid}/secret-manager.
type SecretManagerSource struct {
	Linked               bool                       `json:"linked"`
	IntegrationTokenUUID *string                    `json:"integration_token_uuid"`
	Provider             *string                    `json:"provider"`
	Settings             map[string]any             `json:"settings"`
	AvailableTokens      []SecretManagerTokenOption `json:"available_tokens"`
}

// SecretManagerKeys is the response of GET /{resource_type}/{uuid}/secret-manager/keys.
type SecretManagerKeys struct {
	Keys []string `json:"keys"`
}

// SecretManagerReference is the response of POST /{resource_type}/{uuid}/secret-manager/references.
type SecretManagerReference struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SecretManagerImport is the response of POST /{resource_type}/{uuid}/secret-manager/import.
type SecretManagerImport struct {
	Imported []string `json:"imported"`
	Skipped  []string `json:"skipped"`
}

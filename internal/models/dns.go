package models

// DNSZone is a DNS zone that a Cloudflare token of the team can access (GET /dns/zones).
type DNSZone struct {
	Zone                 string  `json:"zone"`
	AccountName          *string `json:"account_name"`
	IntegrationTokenUUID string  `json:"integration_token_uuid"`
	IntegrationTokenName string  `json:"integration_token_name"`
	ManagedRecordsCount  int     `json:"managed_records_count"`
}

// ManagedDNSRecord is a DNS record that Coolify manages at a DNS provider.
type ManagedDNSRecord struct {
	UUID                 string `json:"uuid"`
	Hostname             string `json:"hostname"`
	Type                 string `json:"type"`
	Address              string `json:"address"`
	Owned                bool   `json:"owned"`
	IntegrationTokenUUID string `json:"integration_token_uuid"`
}

// DNSProvider is a connected DNS provider that manages a hostname.
type DNSProvider struct {
	IntegrationTokenUUID string `json:"integration_token_uuid"`
	IntegrationTokenName string `json:"integration_token_name"`
	Zone                 string `json:"zone"`
	AutomaticDNS         bool   `json:"automatic_dns"`
	ProxiedDNS           bool   `json:"proxied_dns"`
}

// DNSHostnameStatus is the DNS state of one domain of a resource.
type DNSHostnameStatus struct {
	Hostname  string             `json:"hostname"`
	Providers []DNSProvider      `json:"providers"`
	Records   []ManagedDNSRecord `json:"records"`
}

// ResourceDNSStatus is the response of GET /{applications|services}/{uuid}/dns.
type ResourceDNSStatus struct {
	DefaultAddress  *string             `json:"default_address"`
	PublicAddresses []string            `json:"public_addresses"`
	Hostnames       []DNSHostnameStatus `json:"hostnames"`
}

// DNSRecordCreateRequest is the body for POST /{applications|services}/{uuid}/dns.
type DNSRecordCreateRequest struct {
	Hostname             string `json:"hostname"`
	IntegrationTokenUUID string `json:"integration_token_uuid,omitempty"`
	Address              string `json:"address,omitempty"`
}

// DNSRecordReplaceRequest is the body for POST /{applications|services}/{uuid}/dns/replace.
type DNSRecordReplaceRequest struct {
	Hostname             string `json:"hostname"`
	IntegrationTokenUUID string `json:"integration_token_uuid,omitempty"`
	RecordID             string `json:"record_id"`
	Current              string `json:"current"`
	Address              string `json:"address,omitempty"`
}

// DNSRecordResponse wraps the record of a create or replace response.
type DNSRecordResponse struct {
	Record ManagedDNSRecord `json:"record"`
}

// DNSConflict is the conflicting record of a 409 create response.
type DNSConflict struct {
	RecordID string `json:"record_id"`
	Current  string `json:"current"`
	Proposed string `json:"proposed"`
}

// DNSReleaseResult is the response of DELETE /{applications|services}/{uuid}/dns/{hostname}.
type DNSReleaseResult struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

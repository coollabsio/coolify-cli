package models

// Tag represents a team-scoped resource tag.
type Tag struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at" table:"-"`
	UpdatedAt string `json:"updated_at" table:"-"`
}

// Destination represents a Docker network attached to a server.
type Destination struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	Network    string `json:"network"`
	Type       string `json:"type"`
	ServerUUID string `json:"server_uuid"`
	CreatedAt  string `json:"created_at" table:"-"`
	UpdatedAt  string `json:"updated_at" table:"-"`
}

type DestinationCreateRequest struct {
	Name    string `json:"name,omitempty"`
	Network string `json:"network"`
	Type    string `json:"type,omitempty"`
}

type CloudProvider string

const (
	CloudProviderHetzner      CloudProvider = "hetzner"
	CloudProviderDigitalOcean CloudProvider = "digitalocean"
	CloudProviderVultr        CloudProvider = "vultr"
)

func (p CloudProvider) Valid() bool {
	return p == CloudProviderHetzner || p == CloudProviderDigitalOcean || p == CloudProviderVultr
}

type CloudToken struct {
	UUID         string        `json:"uuid"`
	Name         string        `json:"name"`
	Provider     CloudProvider `json:"provider"`
	Token        *string       `json:"token,omitempty" sensitive:"true"`
	TeamID       int           `json:"team_id" table:"-"`
	ServersCount int           `json:"servers_count"`
	CreatedAt    string        `json:"created_at" table:"-"`
	UpdatedAt    string        `json:"updated_at" table:"-"`
}

type CloudTokenCreateRequest struct {
	Provider CloudProvider `json:"provider"`
	Token    string        `json:"token" sensitive:"true" table:"-"`
	Name     string        `json:"name"`
}

type CloudTokenUpdateRequest struct {
	Name string `json:"name"`
}

type CloudTokenValidation struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

type HetznerLocation struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Country     string  `json:"country"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type HetznerServerType struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Cores       int            `json:"cores"`
	Memory      float64        `json:"memory"`
	Disk        int            `json:"disk"`
	Prices      []HetznerPrice `json:"prices" table:"-"`
}

type HetznerPrice struct {
	Location     string          `json:"location"`
	PriceHourly  HetznerPriceNet `json:"price_hourly"`
	PriceMonthly HetznerPriceNet `json:"price_monthly"`
}

type HetznerPriceNet struct {
	Net   string `json:"net"`
	Gross string `json:"gross"`
}

type HetznerImage struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	OSFlavor     string `json:"os_flavor"`
	OSVersion    string `json:"os_version"`
	Architecture string `json:"architecture"`
}

type ProviderSSHKey struct {
	ID          any    `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key" sensitive:"true"`
}

type HetznerFirewall struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type HetznerNetwork struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IPRange string `json:"ip_range"`
}

type HetznerServerCreateRequest struct {
	CloudProviderTokenUUID string `json:"cloud_provider_token_uuid"`
	Location               string `json:"location"`
	ServerType             string `json:"server_type"`
	Image                  int    `json:"image"`
	Name                   string `json:"name,omitempty"`
	PrivateKeyUUID         string `json:"private_key_uuid"`
	EnableIPv4             bool   `json:"enable_ipv4"`
	EnableIPv6             bool   `json:"enable_ipv6"`
	EnableBackups          bool   `json:"enable_backups"`
	HetznerSSHKeyIDs       []int  `json:"hetzner_ssh_key_ids,omitempty"`
	HetznerFirewallIDs     []int  `json:"hetzner_firewall_ids,omitempty"`
	HetznerNetworkIDs      []int  `json:"hetzner_network_ids,omitempty"`
	CloudInitScript        string `json:"cloud_init_script,omitempty"`
	InstantValidate        bool   `json:"instant_validate"`
}

type HetznerServerCreateResponse struct {
	UUID            string `json:"uuid"`
	HetznerServerID int    `json:"hetzner_server_id"`
	IP              string `json:"ip" sensitive:"true"`
}

// DigitalOcean and Vultr option responses intentionally preserve provider fields.
type ProviderOption map[string]any

type DigitalOceanServerCreateRequest struct {
	CloudProviderTokenUUID string `json:"cloud_provider_token_uuid"`
	Region                 string `json:"region"`
	Size                   string `json:"size"`
	Image                  any    `json:"image"`
	Name                   string `json:"name,omitempty"`
	PrivateKeyUUID         string `json:"private_key_uuid"`
	EnableIPv6             bool   `json:"enable_ipv6"`
	Monitoring             bool   `json:"monitoring"`
	DigitalOceanSSHKeyIDs  []int  `json:"digitalocean_ssh_key_ids,omitempty"`
	CloudInitScript        string `json:"cloud_init_script,omitempty"`
	InstantValidate        bool   `json:"instant_validate"`
}

type DigitalOceanServerCreateResponse struct {
	UUID                  string `json:"uuid"`
	DigitalOceanDropletID int    `json:"digitalocean_droplet_id"`
	IP                    string `json:"ip" sensitive:"true"`
}

type VultrServerCreateRequest struct {
	CloudProviderTokenUUID string   `json:"cloud_provider_token_uuid"`
	Region                 string   `json:"region"`
	Plan                   string   `json:"plan"`
	OSID                   int      `json:"os_id"`
	Name                   string   `json:"name,omitempty"`
	PrivateKeyUUID         string   `json:"private_key_uuid"`
	EnableIPv6             bool     `json:"enable_ipv6"`
	DisablePublicIPv4      bool     `json:"disable_public_ipv4"`
	VultrSSHKeyIDs         []string `json:"vultr_ssh_key_ids,omitempty"`
	CloudInitScript        string   `json:"cloud_init_script,omitempty"`
	InstantValidate        bool     `json:"instant_validate"`
}

type VultrServerCreateResponse struct {
	UUID            string `json:"uuid"`
	VultrInstanceID string `json:"vultr_instance_id"`
	IP              string `json:"ip" sensitive:"true"`
}

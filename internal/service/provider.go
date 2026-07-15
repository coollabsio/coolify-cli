package service

import (
	"context"
	"net/url"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func providerPath(path, tokenUUID string) string {
	query := url.Values{"cloud_provider_token_uuid": {tokenUUID}}
	return path + "?" + query.Encode()
}

type HetznerService struct{ client *api.Client }

func NewHetznerService(client *api.Client) *HetznerService { return &HetznerService{client: client} }
func (s *HetznerService) Locations(ctx context.Context, token string) ([]models.HetznerLocation, error) {
	var out []models.HetznerLocation
	err := s.client.Get(ctx, providerPath("hetzner/locations", token), &out)
	return out, err
}
func (s *HetznerService) ServerTypes(ctx context.Context, token string) ([]models.HetznerServerType, error) {
	var out []models.HetznerServerType
	err := s.client.Get(ctx, providerPath("hetzner/server-types", token), &out)
	return out, err
}
func (s *HetznerService) Images(ctx context.Context, token string) ([]models.HetznerImage, error) {
	var out []models.HetznerImage
	err := s.client.Get(ctx, providerPath("hetzner/images", token), &out)
	return out, err
}
func (s *HetznerService) SSHKeys(ctx context.Context, token string) ([]models.ProviderSSHKey, error) {
	var out []models.ProviderSSHKey
	err := s.client.Get(ctx, providerPath("hetzner/ssh-keys", token), &out)
	return out, err
}
func (s *HetznerService) Firewalls(ctx context.Context, token string) ([]models.HetznerFirewall, error) {
	var out []models.HetznerFirewall
	err := s.client.Get(ctx, providerPath("hetzner/firewalls", token), &out)
	return out, err
}
func (s *HetznerService) Networks(ctx context.Context, token string) ([]models.HetznerNetwork, error) {
	var out []models.HetznerNetwork
	err := s.client.Get(ctx, providerPath("hetzner/networks", token), &out)
	return out, err
}
func (s *HetznerService) Create(ctx context.Context, req models.HetznerServerCreateRequest) (*models.HetznerServerCreateResponse, error) {
	var out models.HetznerServerCreateResponse
	err := s.client.Post(ctx, "servers/hetzner", req, &out)
	return &out, err
}

type DigitalOceanService struct{ client *api.Client }

func NewDigitalOceanService(client *api.Client) *DigitalOceanService {
	return &DigitalOceanService{client: client}
}
func (s *DigitalOceanService) Regions(ctx context.Context, token string) ([]models.ProviderOption, error) {
	return s.options(ctx, "digitalocean/regions", token)
}
func (s *DigitalOceanService) Sizes(ctx context.Context, token string) ([]models.ProviderOption, error) {
	return s.options(ctx, "digitalocean/sizes", token)
}
func (s *DigitalOceanService) Images(ctx context.Context, token string) ([]models.ProviderOption, error) {
	return s.options(ctx, "digitalocean/images", token)
}
func (s *DigitalOceanService) SSHKeys(ctx context.Context, token string) ([]models.ProviderOption, error) {
	return s.options(ctx, "digitalocean/ssh-keys", token)
}
func (s *DigitalOceanService) options(ctx context.Context, path, token string) ([]models.ProviderOption, error) {
	var out []models.ProviderOption
	err := s.client.Get(ctx, providerPath(path, token), &out)
	return out, err
}
func (s *DigitalOceanService) Create(ctx context.Context, req models.DigitalOceanServerCreateRequest) (*models.DigitalOceanServerCreateResponse, error) {
	var out models.DigitalOceanServerCreateResponse
	err := s.client.Post(ctx, "servers/digitalocean", req, &out)
	return &out, err
}

type VultrService struct{ client *api.Client }

func NewVultrService(client *api.Client) *VultrService { return &VultrService{client: client} }
func (s *VultrService) Regions(ctx context.Context, token string) ([]models.ProviderOption, error) {
	return s.options(ctx, "vultr/regions", token)
}
func (s *VultrService) Plans(ctx context.Context, token string) ([]models.ProviderOption, error) {
	return s.options(ctx, "vultr/plans", token)
}
func (s *VultrService) OperatingSystems(ctx context.Context, token string) ([]models.ProviderOption, error) {
	return s.options(ctx, "vultr/os", token)
}
func (s *VultrService) SSHKeys(ctx context.Context, token string) ([]models.ProviderOption, error) {
	return s.options(ctx, "vultr/ssh-keys", token)
}
func (s *VultrService) options(ctx context.Context, path, token string) ([]models.ProviderOption, error) {
	var out []models.ProviderOption
	err := s.client.Get(ctx, providerPath(path, token), &out)
	return out, err
}
func (s *VultrService) Create(ctx context.Context, req models.VultrServerCreateRequest) (*models.VultrServerCreateResponse, error) {
	var out models.VultrServerCreateResponse
	err := s.client.Post(ctx, "servers/vultr", req, &out)
	return &out, err
}

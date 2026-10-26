package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

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

// ErrHostingerPaymentPending is returned when Hostinger accepted the order but
// is still processing payment (HTTP 202). No Coolify server exists yet.
var ErrHostingerPaymentPending = errors.New("hostinger payment is still processing")

type HostingerService struct{ client *api.Client }

func NewHostingerService(client *api.Client) *HostingerService {
	return &HostingerService{client: client}
}
func (s *HostingerService) DataCenters(ctx context.Context, token string) ([]models.HostingerDataCenter, error) {
	var out []models.HostingerDataCenter
	err := s.client.Get(ctx, providerPath("hostinger/data-centers", token), &out)
	return out, err
}
func (s *HostingerService) Catalog(ctx context.Context, token string) ([]models.HostingerCatalogItem, error) {
	var out []models.HostingerCatalogItem
	err := s.client.Get(ctx, providerPath("hostinger/catalog", token), &out)
	return out, err
}
func (s *HostingerService) Templates(ctx context.Context, token string) ([]models.HostingerTemplate, error) {
	var out []models.HostingerTemplate
	err := s.client.Get(ctx, providerPath("hostinger/templates", token), &out)
	return out, err
}
func (s *HostingerService) SSHKeys(ctx context.Context, token string) ([]models.HostingerSSHKey, error) {
	var out []models.HostingerSSHKey
	err := s.client.Get(ctx, providerPath("hostinger/ssh-keys", token), &out)
	return out, err
}
func (s *HostingerService) PostInstallScripts(ctx context.Context, token string) ([]models.HostingerPostInstallScript, error) {
	var out []models.HostingerPostInstallScript
	err := s.client.Get(ctx, providerPath("hostinger/post-install-scripts", token), &out)
	return out, err
}

// Create purchases a Hostinger VPS and registers it in Coolify. The request is
// sent exactly once: retrying a failed purchase could buy a second server.
// A 202 response returns the response together with ErrHostingerPaymentPending.
func (s *HostingerService) Create(ctx context.Context, req models.HostingerServerCreateRequest) (*models.HostingerServerCreateResponse, error) {
	var out models.HostingerServerCreateResponse
	if err := s.client.PostOnce(ctx, "servers/hostinger", req, &out); err != nil {
		return nil, describeHostingerCreateError(err)
	}
	if out.PaymentPending() {
		message := out.Message
		if message == "" {
			message = "no server was returned"
		}
		return &out, fmt.Errorf("%w: %s", ErrHostingerPaymentPending, message)
	}
	return &out, nil
}

// describeHostingerCreateError adds validation details, rate-limit timing,
// and possibly-purchased VM IDs from the API error body to the error.
func describeHostingerCreateError(err error) error {
	var apiErr *api.Error
	if !errors.As(err, &apiErr) {
		return err
	}
	var body struct {
		Errors                    map[string][]string `json:"errors"`
		HostingerVirtualMachineID *int                `json:"hostinger_virtual_machine_id"`
	}
	_ = json.Unmarshal(apiErr.Body, &body)

	var details []string
	switch {
	case apiErr.StatusCode == 422 && len(body.Errors) > 0:
		fields := make([]string, 0, len(body.Errors))
		for field := range body.Errors {
			fields = append(fields, field)
		}
		sort.Strings(fields)
		for _, field := range fields {
			details = append(details, fmt.Sprintf("%s: %s", field, strings.Join(body.Errors[field], " ")))
		}
	case apiErr.StatusCode == 429:
		if apiErr.RetryAfter != "" {
			details = append(details, fmt.Sprintf("retry after %s seconds", apiErr.RetryAfter))
		}
	case apiErr.StatusCode >= 500 && body.HostingerVirtualMachineID != nil:
		details = append(details, fmt.Sprintf("Hostinger virtual machine %d may already have been purchased; check your Hostinger account before retrying", *body.HostingerVirtualMachineID))
	}
	if len(details) == 0 {
		return err
	}
	return fmt.Errorf("%w (%s)", err, strings.Join(details, "; "))
}

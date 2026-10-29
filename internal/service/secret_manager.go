package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// dopplerTokenPattern matches Doppler service tokens and service account tokens.
var dopplerTokenPattern = regexp.MustCompile(`^dp\.(st|sa)\.`)

// SecretManagerService handles secret manager integration tokens and
// application secret manager configuration.
type SecretManagerService struct {
	client *api.Client
}

// NewSecretManagerService creates a new secret manager service
func NewSecretManagerService(client *api.Client) *SecretManagerService {
	return &SecretManagerService{client: client}
}

// ParseJSONObject parses a JSON object flag value. An empty value returns nil.
func ParseJSONObject(flag, value string) (map[string]any, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(value), &out); err != nil || out == nil {
		return nil, fmt.Errorf("--%s must be a JSON object, e.g. '{\"key\":\"value\"}'", flag)
	}
	return out, nil
}

func metadataString(metadata map[string]any, key string) string {
	value, _ := metadata[key].(string)
	return strings.TrimSpace(value)
}

// ValidateIntegrationTokenOptions checks everything except the token itself
// (provider, name and provider specific metadata), so a CLI can fail fast before
// prompting for the secret.
func ValidateIntegrationTokenOptions(req models.IntegrationTokenCreateRequest) error {
	if !slices.Contains(models.IntegrationTokenProviders, req.Provider) {
		return fmt.Errorf("--provider must be one of: %s", strings.Join(models.IntegrationTokenProviders, ", "))
	}
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("--name is required")
	}
	if len(req.Name) > 255 {
		return fmt.Errorf("--name must not be longer than 255 characters")
	}
	if err := ValidateCapabilities(req.Provider, req.Capabilities); err != nil {
		return err
	}
	if req.Provider != "cloudflare" && (req.AutomaticDNS != nil || req.ProxiedDNS != nil) {
		return fmt.Errorf("--automatic-dns and --proxied-dns are only for cloudflare tokens")
	}
	switch req.Provider {
	case "cloudflare":
		if len(req.Metadata) > 0 {
			return fmt.Errorf("--metadata is not used for cloudflare tokens")
		}
	case "infisical":
		if metadataString(req.Metadata, "base_url") == "" || metadataString(req.Metadata, "client_id") == "" {
			return fmt.Errorf(`infisical requires --metadata '{"base_url":"https://...","client_id":"..."}'`)
		}
	case "vault":
		if metadataString(req.Metadata, "base_url") == "" {
			return fmt.Errorf(`vault requires --metadata '{"base_url":"https://..."}' (optionally "namespace")`)
		}
	}
	return nil
}

// ValidateIntegrationTokenCreateRequest checks the request against the API's validation rules,
// so mistakes are reported with a useful message instead of a bare "Validation failed.".
func ValidateIntegrationTokenCreateRequest(req models.IntegrationTokenCreateRequest) error {
	if err := ValidateIntegrationTokenOptions(req); err != nil {
		return err
	}
	if req.Token == "" {
		return fmt.Errorf("the provider token must not be empty")
	}
	if req.Provider == "doppler" && !dopplerTokenPattern.MatchString(req.Token) {
		return fmt.Errorf("doppler tokens must start with dp.st. or dp.sa.")
	}
	return nil
}

// CreateIntegrationToken creates and validates a secret manager integration token
// (POST /security/integration-tokens). The token itself is never returned.
func (s *SecretManagerService) CreateIntegrationToken(ctx context.Context, req models.IntegrationTokenCreateRequest) (*models.UUID, error) {
	if err := ValidateIntegrationTokenCreateRequest(req); err != nil {
		return nil, err
	}
	var resp models.UUID
	if err := s.client.Post(ctx, "security/integration-tokens", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to create integration token: %w", err)
	}
	return &resp, nil
}

// SecretManagerResourceTypes lists the resource types that can have a secret manager source.
var SecretManagerResourceTypes = []string{"applications", "services", "databases"}

func secretManagerPath(resourceType, uuid string) (string, error) {
	if !slices.Contains(SecretManagerResourceTypes, resourceType) {
		return "", fmt.Errorf("unsupported resource type %q", resourceType)
	}
	return resourceType + "/" + url.PathEscape(uuid) + "/secret-manager", nil
}

// GetSource gets the secret manager source of a resource (GET /{resource_type}/{uuid}/secret-manager).
func (s *SecretManagerService) GetSource(ctx context.Context, resourceType, uuid string) (*models.SecretManagerSource, error) {
	path, err := secretManagerPath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var resp models.SecretManagerSource
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("failed to get secret manager source of %s: %w", uuid, err)
	}
	return &resp, nil
}

// Configure sets the secret manager source of a resource (PATCH /{resource_type}/{uuid}/secret-manager).
func (s *SecretManagerService) Configure(ctx context.Context, resourceType, uuid string, req models.ApplicationSecretManagerRequest) (*models.ApplicationSecretManager, error) {
	if strings.TrimSpace(req.IntegrationTokenUUID) == "" {
		return nil, fmt.Errorf("--integration-token is required")
	}
	path, err := secretManagerPath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var resp models.ApplicationSecretManager
	if err := s.client.Patch(ctx, path, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to configure secret manager for %s: %w", uuid, err)
	}
	return &resp, nil
}

// RemoveSource removes the secret manager source of a resource (DELETE /{resource_type}/{uuid}/secret-manager).
func (s *SecretManagerService) RemoveSource(ctx context.Context, resourceType, uuid string) error {
	path, err := secretManagerPath(resourceType, uuid)
	if err != nil {
		return err
	}
	if err := s.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("failed to remove secret manager source of %s: %w", uuid, err)
	}
	return nil
}

// Keys lists the key names of the remote secrets (GET /{resource_type}/{uuid}/secret-manager/keys).
func (s *SecretManagerService) Keys(ctx context.Context, resourceType, uuid string) (*models.SecretManagerKeys, error) {
	path, err := secretManagerPath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var resp models.SecretManagerKeys
	if err := s.client.Get(ctx, path+"/keys", &resp); err != nil {
		return nil, fmt.Errorf("failed to list secret manager keys of %s: %w", uuid, err)
	}
	return &resp, nil
}

// AddReference adds a {{vault.KEY}} variable (POST /{resource_type}/{uuid}/secret-manager/references).
func (s *SecretManagerService) AddReference(ctx context.Context, resourceType, uuid, key string) (*models.SecretManagerReference, error) {
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("--key is required")
	}
	path, err := secretManagerPath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var resp models.SecretManagerReference
	if err := s.client.Post(ctx, path+"/references", map[string]string{"key": key}, &resp); err != nil {
		return nil, fmt.Errorf("failed to add secret manager reference %s: %w", key, describeValidationErrors(err))
	}
	return &resp, nil
}

// Import adds a {{vault.KEY}} variable for every remote key without a variable
// (POST /{resource_type}/{uuid}/secret-manager/import).
func (s *SecretManagerService) Import(ctx context.Context, resourceType, uuid string) (*models.SecretManagerImport, error) {
	path, err := secretManagerPath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var resp models.SecretManagerImport
	if err := s.client.PostOnce(ctx, path+"/import", map[string]any{}, &resp); err != nil {
		return nil, fmt.Errorf("failed to import secret manager references of %s: %w", uuid, err)
	}
	return &resp, nil
}

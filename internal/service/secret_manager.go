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
	switch req.Provider {
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

// ConfigureApplication sets the secret manager source of an application
// (PATCH /applications/{uuid}/secret-manager).
func (s *SecretManagerService) ConfigureApplication(ctx context.Context, appUUID string, req models.ApplicationSecretManagerRequest) (*models.ApplicationSecretManager, error) {
	if strings.TrimSpace(req.IntegrationTokenUUID) == "" {
		return nil, fmt.Errorf("--integration-token is required")
	}
	var resp models.ApplicationSecretManager
	path := "applications/" + url.PathEscape(appUUID) + "/secret-manager"
	if err := s.client.Patch(ctx, path, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to configure secret manager for application %s: %w", appUUID, err)
	}
	return &resp, nil
}

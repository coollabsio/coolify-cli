package service

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/models"
)

// allowedCapabilities mirrors IntegrationToken::allowedCapabilities() of the API.
func allowedCapabilities(provider string) []string {
	if provider == "cloudflare" {
		return []string{"dns", "tunnel"}
	}
	return []string{"secrets"}
}

// ValidateCapabilities checks the capabilities of a token of the provider. Cloudflare Tunnel
// setup also writes DNS records, so tunnel needs dns.
func ValidateCapabilities(provider string, capabilities []string) error {
	allowed := allowedCapabilities(provider)
	for _, capability := range capabilities {
		if !slices.Contains(allowed, capability) {
			return fmt.Errorf("--capabilities for %s must be one of: %s", provider, strings.Join(allowed, ", "))
		}
	}
	if slices.Contains(capabilities, "tunnel") && !slices.Contains(capabilities, "dns") {
		return fmt.Errorf("the tunnel capability also needs the dns capability")
	}
	return nil
}

// ListIntegrationTokens lists the integration tokens of the team (GET /security/integration-tokens).
func (s *SecretManagerService) ListIntegrationTokens(ctx context.Context) ([]models.IntegrationToken, error) {
	var tokens []models.IntegrationToken
	if err := s.client.Get(ctx, "security/integration-tokens", &tokens); err != nil {
		return nil, fmt.Errorf("failed to list integration tokens: %w", err)
	}
	return tokens, nil
}

// GetIntegrationToken gets an integration token by UUID (GET /security/integration-tokens/{uuid}).
func (s *SecretManagerService) GetIntegrationToken(ctx context.Context, uuid string) (*models.IntegrationToken, error) {
	var token models.IntegrationToken
	if err := s.client.Get(ctx, tokenPath(uuid), &token); err != nil {
		return nil, fmt.Errorf("failed to get integration token %s: %w", uuid, err)
	}
	return &token, nil
}

// UpdateIntegrationToken updates an integration token (PATCH /security/integration-tokens/{uuid}).
func (s *SecretManagerService) UpdateIntegrationToken(ctx context.Context, uuid string, req models.IntegrationTokenUpdateRequest) (*models.IntegrationToken, error) {
	if req.Name != nil && (strings.TrimSpace(*req.Name) == "" || len(*req.Name) > 255) {
		return nil, fmt.Errorf("--name must be 1 to 255 characters")
	}
	var token models.IntegrationToken
	if err := s.client.Patch(ctx, tokenPath(uuid), req, &token); err != nil {
		return nil, fmt.Errorf("failed to update integration token %s: %w", uuid, err)
	}
	return &token, nil
}

// DeleteIntegrationToken deletes an integration token (DELETE /security/integration-tokens/{uuid}).
func (s *SecretManagerService) DeleteIntegrationToken(ctx context.Context, uuid string) error {
	if err := s.client.Delete(ctx, tokenPath(uuid)); err != nil {
		return fmt.Errorf("failed to delete integration token %s: %w", uuid, err)
	}
	return nil
}

// SyncZones reads the DNS zones of a Cloudflare token again (POST /security/integration-tokens/{uuid}/sync-zones).
func (s *SecretManagerService) SyncZones(ctx context.Context, uuid string) (*models.ZoneSyncResult, error) {
	var resp models.ZoneSyncResult
	if err := s.client.Post(ctx, tokenPath(uuid)+"/sync-zones", map[string]any{}, &resp); err != nil {
		return nil, fmt.Errorf("failed to sync zones of integration token %s: %w", uuid, err)
	}
	return &resp, nil
}

// ListDNSZones lists the DNS zones of the Cloudflare tokens of the team (GET /dns/zones).
func (s *SecretManagerService) ListDNSZones(ctx context.Context) ([]models.DNSZone, error) {
	var zones []models.DNSZone
	if err := s.client.Get(ctx, "dns/zones", &zones); err != nil {
		return nil, fmt.Errorf("failed to list DNS zones: %w", err)
	}
	return zones, nil
}

func tokenPath(uuid string) string {
	return "security/integration-tokens/" + url.PathEscape(uuid)
}

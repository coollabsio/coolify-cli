package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// PrivateKeyService handles private key-related operations
type PrivateKeyService struct {
	client *api.Client
}

// NewPrivateKeyService creates a new private key service
func NewPrivateKeyService(client *api.Client) *PrivateKeyService {
	return &PrivateKeyService{
		client: client,
	}
}

// List retrieves all private keys
func (s *PrivateKeyService) List(ctx context.Context) ([]models.PrivateKey, error) {
	var keys []models.PrivateKey
	err := s.client.Get(ctx, "security/keys", &keys)
	if err != nil {
		return nil, fmt.Errorf("failed to list private keys: %w", err)
	}
	return keys, nil
}

// Create creates a new private key
func (s *PrivateKeyService) Create(ctx context.Context, req models.PrivateKeyCreateRequest) (*models.PrivateKey, error) {
	var key models.PrivateKey
	err := s.client.Post(ctx, "security/keys", req, &key)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}
	return &key, nil
}

// Delete deletes a private key by UUID
func (s *PrivateKeyService) Delete(ctx context.Context, uuid string) error {
	err := s.client.Delete(ctx, fmt.Sprintf("security/keys/%s", uuid))
	if err != nil {
		return fmt.Errorf("failed to delete private key %s: %w", uuid, err)
	}
	return nil
}

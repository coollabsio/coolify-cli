package service

import (
	"context"
	"net/url"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

type SecretManagerService struct {
	client *api.Client
}

func NewSecretManagerService(client *api.Client) *SecretManagerService {
	return &SecretManagerService{client: client}
}

func (s *SecretManagerService) CreateToken(ctx context.Context, request models.SecretManagerTokenCreateRequest) (*models.UUID, error) {
	var response models.UUID
	err := s.client.Post(ctx, "security/integration-tokens", request, &response)
	return &response, err
}

func (s *SecretManagerService) ConfigureApplication(ctx context.Context, applicationUUID string, request models.ApplicationSecretManagerRequest) (*models.ApplicationSecretManager, error) {
	var response models.ApplicationSecretManager
	err := s.client.Patch(ctx, "applications/"+url.PathEscape(applicationUUID)+"/secret-manager", request, &response)
	return &response, err
}

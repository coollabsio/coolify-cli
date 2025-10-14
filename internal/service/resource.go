package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// ResourceService handles resource-related operations
type ResourceService struct {
	client *api.Client
}

// NewResourceService creates a new resource service
func NewResourceService(client *api.Client) *ResourceService {
	return &ResourceService{
		client: client,
	}
}

// List retrieves all resources
func (s *ResourceService) List(ctx context.Context) ([]models.Resource, error) {
	var resources []models.Resource
	err := s.client.Get(ctx, "resources", &resources)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}
	return resources, nil
}

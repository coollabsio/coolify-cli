package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// DomainService handles domain-related operations
type DomainService struct {
	client *api.Client
}

// NewDomainService creates a new domain service
func NewDomainService(client *api.Client) *DomainService {
	return &DomainService{
		client: client,
	}
}

// List retrieves all domains
func (s *DomainService) List(ctx context.Context) ([]models.Domain, error) {
	var domains []models.Domain
	err := s.client.Get(ctx, "domains", &domains)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}
	return domains, nil
}

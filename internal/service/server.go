package service

import (
	"context"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// ServerService handles server-related operations
type ServerService struct {
	client *api.Client
}

// NewServerService creates a new server service
func NewServerService(client *api.Client) *ServerService {
	return &ServerService{client: client}
}

// List returns all servers
func (s *ServerService) List(ctx context.Context) ([]models.Server, error) {
	var servers []models.Server
	err := s.client.Get(ctx, "servers", &servers)
	return servers, err
}

// Get returns a single server by UUID
func (s *ServerService) Get(ctx context.Context, uuid string) (*models.Server, error) {
	var server models.Server
	err := s.client.Get(ctx, "servers/"+uuid, &server)
	return &server, err
}

// GetResources returns resources for a server
func (s *ServerService) GetResources(ctx context.Context, uuid string) (*models.Resources, error) {
	var resources models.Resources
	err := s.client.Get(ctx, "servers/"+uuid+"?resources=true", &resources)
	return &resources, err
}

// Create creates a new server
func (s *ServerService) Create(ctx context.Context, req models.ServerCreateRequest) (*models.Response, error) {
	var response models.Response
	err := s.client.Post(ctx, "servers", req, &response)
	return &response, err
}

// Delete deletes a server by UUID
func (s *ServerService) Delete(ctx context.Context, uuid string) error {
	return s.client.Delete(ctx, "servers/"+uuid)
}

// Validate validates a server by UUID
func (s *ServerService) Validate(ctx context.Context, uuid string) (*models.Response, error) {
	var response models.Response
	err := s.client.Get(ctx, "servers/"+uuid+"/validate", &response)
	return &response, err
}

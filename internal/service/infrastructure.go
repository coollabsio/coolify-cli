package service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

type TagResourceType string

const (
	TagResourceApplications TagResourceType = "applications"
	TagResourceDatabases    TagResourceType = "databases"
	TagResourceServices     TagResourceType = "services"
)

type TagService struct{ client *api.Client }

func NewTagService(client *api.Client) *TagService { return &TagService{client: client} }

func (s *TagService) List(ctx context.Context) ([]models.Tag, error) {
	var tags []models.Tag
	err := s.client.Get(ctx, "tags", &tags)
	return tags, err
}

func (s *TagService) resourcePath(resourceType TagResourceType, uuid string) (string, error) {
	if resourceType != TagResourceApplications && resourceType != TagResourceDatabases && resourceType != TagResourceServices {
		return "", fmt.Errorf("unsupported tag resource type %q", resourceType)
	}
	return string(resourceType) + "/" + url.PathEscape(uuid) + "/tags", nil
}

func (s *TagService) ListForResource(ctx context.Context, resourceType TagResourceType, uuid string) ([]models.Tag, error) {
	path, err := s.resourcePath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var tags []models.Tag
	err = s.client.Get(ctx, path, &tags)
	return tags, err
}

func (s *TagService) CreateForResource(ctx context.Context, resourceType TagResourceType, uuid, name string) ([]models.Tag, error) {
	path, err := s.resourcePath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var tags []models.Tag
	err = s.client.Post(ctx, path, map[string]string{"tag_name": name}, &tags)
	return tags, err
}

func (s *TagService) DeleteForResource(ctx context.Context, resourceType TagResourceType, uuid, tagUUID string) error {
	path, err := s.resourcePath(resourceType, uuid)
	if err != nil {
		return err
	}
	return s.client.Delete(ctx, path+"/"+url.PathEscape(tagUUID))
}

type DestinationService struct{ client *api.Client }

func NewDestinationService(client *api.Client) *DestinationService {
	return &DestinationService{client: client}
}

func (s *DestinationService) List(ctx context.Context) ([]models.Destination, error) {
	var destinations []models.Destination
	err := s.client.Get(ctx, "destinations", &destinations)
	return destinations, err
}

func (s *DestinationService) ListByServer(ctx context.Context, serverUUID string) ([]models.Destination, error) {
	var destinations []models.Destination
	err := s.client.Get(ctx, "servers/"+url.PathEscape(serverUUID)+"/destinations", &destinations)
	return destinations, err
}

func (s *DestinationService) Get(ctx context.Context, uuid string) (*models.Destination, error) {
	var destination models.Destination
	err := s.client.Get(ctx, "destinations/"+url.PathEscape(uuid), &destination)
	return &destination, err
}

func (s *DestinationService) CreateForServer(ctx context.Context, serverUUID string, req models.DestinationCreateRequest) (*models.Destination, error) {
	var destination models.Destination
	err := s.client.Post(ctx, "servers/"+url.PathEscape(serverUUID)+"/destinations", req, &destination)
	return &destination, err
}

func (s *DestinationService) Delete(ctx context.Context, uuid string) error {
	return s.client.Delete(ctx, "destinations/"+url.PathEscape(uuid))
}

type CloudTokenService struct{ client *api.Client }

func NewCloudTokenService(client *api.Client) *CloudTokenService {
	return &CloudTokenService{client: client}
}

func (s *CloudTokenService) List(ctx context.Context) ([]models.CloudToken, error) {
	var tokens []models.CloudToken
	err := s.client.Get(ctx, "cloud-tokens", &tokens)
	return tokens, err
}

func (s *CloudTokenService) Get(ctx context.Context, uuid string) (*models.CloudToken, error) {
	var token models.CloudToken
	err := s.client.Get(ctx, "cloud-tokens/"+url.PathEscape(uuid), &token)
	return &token, err
}

func (s *CloudTokenService) Create(ctx context.Context, req models.CloudTokenCreateRequest) (*models.UUID, error) {
	var response models.UUID
	err := s.client.Post(ctx, "cloud-tokens", req, &response)
	return &response, err
}

func (s *CloudTokenService) Update(ctx context.Context, uuid string, req models.CloudTokenUpdateRequest) (*models.UUID, error) {
	var response models.UUID
	err := s.client.Patch(ctx, "cloud-tokens/"+url.PathEscape(uuid), req, &response)
	return &response, err
}

func (s *CloudTokenService) Validate(ctx context.Context, uuid string) (*models.CloudTokenValidation, error) {
	var response models.CloudTokenValidation
	err := s.client.Post(ctx, "cloud-tokens/"+url.PathEscape(uuid)+"/validate", nil, &response)
	return &response, err
}

func (s *CloudTokenService) Delete(ctx context.Context, uuid string) error {
	return s.client.Delete(ctx, "cloud-tokens/"+url.PathEscape(uuid))
}

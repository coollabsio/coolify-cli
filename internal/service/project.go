package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// ProjectService handles project-related operations
type ProjectService struct {
	client *api.Client
}

// NewProjectService creates a new project service
func NewProjectService(client *api.Client) *ProjectService {
	return &ProjectService{
		client: client,
	}
}

// List retrieves all projects
func (s *ProjectService) List(ctx context.Context) ([]models.Project, error) {
	var projects []models.Project
	err := s.client.Get(ctx, "projects", &projects)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}

// Get retrieves a specific project by UUID
func (s *ProjectService) Get(ctx context.Context, uuid string) (*models.Project, error) {
	var project models.Project
	err := s.client.Get(ctx, fmt.Sprintf("projects/%s", uuid), &project)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", uuid, err)
	}
	return &project, nil
}

// Create creates a new project
func (s *ProjectService) Create(ctx context.Context, req *models.ProjectCreateRequest) (*models.Project, error) {
	var project models.Project
	err := s.client.Post(ctx, "projects", req, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	return &project, nil
}

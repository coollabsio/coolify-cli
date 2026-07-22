package service

import (
	"context"
	"fmt"
	"net/url"

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
	err := s.client.Get(ctx, "projects/"+url.PathEscape(uuid), &project)
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

// Update patches a project by UUID
func (s *ProjectService) Update(ctx context.Context, uuid string, req models.ProjectUpdateRequest) (*models.Project, error) {
	var project models.Project
	err := s.client.Patch(ctx, "projects/"+url.PathEscape(uuid), req, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to update project %s: %w", uuid, err)
	}
	return &project, nil
}

// Delete deletes a project by UUID
func (s *ProjectService) Delete(ctx context.Context, uuid string) error {
	err := s.client.Delete(ctx, "projects/"+url.PathEscape(uuid))
	if err != nil {
		return fmt.Errorf("failed to delete project %s: %w", uuid, err)
	}
	return nil
}

// ListEnvironments lists environments in a project
func (s *ProjectService) ListEnvironments(ctx context.Context, projectUUID string) ([]models.Environment, error) {
	var environments []models.Environment
	path := fmt.Sprintf("projects/%s/environments", url.PathEscape(projectUUID))
	err := s.client.Get(ctx, path, &environments)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments for project %s: %w", projectUUID, err)
	}
	return environments, nil
}

// CreateEnvironment creates an environment in a project
func (s *ProjectService) CreateEnvironment(ctx context.Context, projectUUID string, req models.EnvironmentCreateRequest) (*models.UUID, error) {
	var resp models.UUID
	path := fmt.Sprintf("projects/%s/environments", url.PathEscape(projectUUID))
	err := s.client.Post(ctx, path, req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment in project %s: %w", projectUUID, err)
	}
	return &resp, nil
}

// UpdateEnvironment patches a project environment name/description
func (s *ProjectService) UpdateEnvironment(ctx context.Context, projectUUID, envNameOrUUID string, req models.EnvironmentUpdateRequest) (map[string]any, error) {
	var out map[string]any
	path := fmt.Sprintf("projects/%s/environments/%s", url.PathEscape(projectUUID), url.PathEscape(envNameOrUUID))
	err := s.client.Patch(ctx, path, req, &out)
	if err != nil {
		return nil, fmt.Errorf("failed to update environment %s in project %s: %w", envNameOrUUID, projectUUID, err)
	}
	return out, nil
}

// DeleteEnvironment deletes an environment by name or UUID
func (s *ProjectService) DeleteEnvironment(ctx context.Context, projectUUID, envNameOrUUID string) error {
	path := fmt.Sprintf("projects/%s/environments/%s", url.PathEscape(projectUUID), url.PathEscape(envNameOrUUID))
	err := s.client.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete environment %s in project %s: %w", envNameOrUUID, projectUUID, err)
	}
	return nil
}

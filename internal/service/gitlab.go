package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// GitLabAppService handles GitLab App (OAuth source) operations
type GitLabAppService struct {
	client *api.Client
}

// NewGitLabAppService creates a new GitLab App service
func NewGitLabAppService(client *api.Client) *GitLabAppService {
	return &GitLabAppService{
		client: client,
	}
}

// List retrieves all GitLab Apps for the current token team (plus system-wide)
func (s *GitLabAppService) List(ctx context.Context) ([]models.GitLabApp, error) {
	var apps []models.GitLabApp
	err := s.client.Get(ctx, "gitlab-apps", &apps)
	if err != nil {
		return nil, fmt.Errorf("failed to list GitLab Apps: %w", err)
	}
	return apps, nil
}

// Get retrieves a GitLab App by numeric ID or UUID (resolved via list)
func (s *GitLabAppService) Get(ctx context.Context, idOrUUID string) (*models.GitLabApp, error) {
	app, err := s.find(ctx, idOrUUID)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// Create creates a new GitLab App
func (s *GitLabAppService) Create(ctx context.Context, req *models.GitLabAppCreateRequest) (*models.GitLabApp, error) {
	var app models.GitLabApp
	err := s.client.Post(ctx, "gitlab-apps", req, &app)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab App: %w", err)
	}
	return &app, nil
}

// Update updates an existing GitLab App (by numeric ID or UUID)
func (s *GitLabAppService) Update(ctx context.Context, idOrUUID string, req *models.GitLabAppUpdateRequest) (*models.GitLabApp, error) {
	id, err := s.resolveID(ctx, idOrUUID)
	if err != nil {
		return nil, err
	}

	var resp models.GitLabAppUpdateResponse
	err = s.client.Patch(ctx, fmt.Sprintf("gitlab-apps/%d", id), req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to update GitLab App %s: %w", idOrUUID, err)
	}
	return &resp.Data, nil
}

// Delete deletes a GitLab App (by numeric ID or UUID)
func (s *GitLabAppService) Delete(ctx context.Context, idOrUUID string) error {
	id, err := s.resolveID(ctx, idOrUUID)
	if err != nil {
		return err
	}

	err = s.client.Delete(ctx, fmt.Sprintf("gitlab-apps/%d", id))
	if err != nil {
		return fmt.Errorf("failed to delete GitLab App %s: %w", idOrUUID, err)
	}
	return nil
}

func (s *GitLabAppService) resolveID(ctx context.Context, idOrUUID string) (int, error) {
	idOrUUID = strings.TrimSpace(idOrUUID)
	if id, err := strconv.Atoi(idOrUUID); err == nil {
		return id, nil
	}

	app, err := s.find(ctx, idOrUUID)
	if err != nil {
		return 0, err
	}
	return app.ID, nil
}

func (s *GitLabAppService) find(ctx context.Context, idOrUUID string) (*models.GitLabApp, error) {
	apps, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	idOrUUID = strings.TrimSpace(idOrUUID)
	for i := range apps {
		if apps[i].UUID == idOrUUID {
			return &apps[i], nil
		}
		if strconv.Itoa(apps[i].ID) == idOrUUID {
			return &apps[i], nil
		}
	}

	return nil, fmt.Errorf("GitLab App %q not found", idOrUUID)
}

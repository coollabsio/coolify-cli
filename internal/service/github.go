package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// GitHubAppService handles GitHub App-related operations
type GitHubAppService struct {
	client *api.Client
}

// NewGitHubAppService creates a new GitHub App service
func NewGitHubAppService(client *api.Client) *GitHubAppService {
	return &GitHubAppService{
		client: client,
	}
}

// List retrieves all GitHub Apps
// Note: This endpoint will be available in a future version of Coolify
func (s *GitHubAppService) List(ctx context.Context) ([]models.GitHubApp, error) {
	var apps []models.GitHubApp
	err := s.client.Get(ctx, "github-apps", &apps)
	if err != nil {
		return nil, fmt.Errorf("failed to list GitHub Apps: %w", err)
	}
	return apps, nil
}

// Get retrieves a specific GitHub App by UUID
// Note: This endpoint will be available in a future version of Coolify
func (s *GitHubAppService) Get(ctx context.Context, uuid string) (*models.GitHubApp, error) {
	var app models.GitHubApp
	err := s.client.Get(ctx, fmt.Sprintf("github-apps/%s", uuid), &app)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub App %s: %w", uuid, err)
	}
	return &app, nil
}

// Create creates a new GitHub App
func (s *GitHubAppService) Create(ctx context.Context, req *models.GitHubAppCreateRequest) (*models.GitHubApp, error) {
	var app models.GitHubApp
	err := s.client.Post(ctx, "github-apps", req, &app)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub App: %w", err)
	}
	return &app, nil
}

// Update updates an existing GitHub App
func (s *GitHubAppService) Update(ctx context.Context, uuid string, req *models.GitHubAppUpdateRequest) error {
	type response struct {
		Message string `json:"message"`
	}
	var resp response
	err := s.client.Patch(ctx, fmt.Sprintf("github-apps/%s", uuid), req, &resp)
	if err != nil {
		return fmt.Errorf("failed to update GitHub App %s: %w", uuid, err)
	}
	return nil
}

// Delete deletes a GitHub App
func (s *GitHubAppService) Delete(ctx context.Context, uuid string) error {
	err := s.client.Delete(ctx, fmt.Sprintf("github-apps/%s", uuid))
	if err != nil {
		return fmt.Errorf("failed to delete GitHub App %s: %w", uuid, err)
	}
	return nil
}

// ListRepositories lists all repositories accessible by a GitHub App
func (s *GitHubAppService) ListRepositories(ctx context.Context, appUUID string) ([]models.GitHubRepository, error) {
	type response struct {
		Repositories []models.GitHubRepository `json:"repositories"`
	}
	var resp response
	err := s.client.Get(ctx, fmt.Sprintf("github-apps/%s/repositories", appUUID), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories for GitHub App %s: %w", appUUID, err)
	}
	return resp.Repositories, nil
}

// ListBranches lists all branches for a repository
func (s *GitHubAppService) ListBranches(ctx context.Context, appUUID string, owner, repo string) ([]models.GitHubBranch, error) {
	type response struct {
		Branches []models.GitHubBranch `json:"branches"`
	}
	var resp response
	err := s.client.Get(ctx, fmt.Sprintf("github-apps/%s/repositories/%s/%s/branches", appUUID, owner, repo), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches for %s/%s: %w", owner, repo, err)
	}
	return resp.Branches, nil
}

package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

// List retrieves all GitHub Apps for the current token team (plus system-wide)
func (s *GitHubAppService) List(ctx context.Context) ([]models.GitHubApp, error) {
	var apps []models.GitHubApp
	err := s.client.Get(ctx, "github-apps", &apps)
	if err != nil {
		return nil, fmt.Errorf("failed to list GitHub Apps: %w", err)
	}
	return apps, nil
}

// Get retrieves a GitHub App by UUID or numeric ID.
// The API has no single-app endpoint, so the app is resolved from the list.
func (s *GitHubAppService) Get(ctx context.Context, uuidOrID string) (*models.GitHubApp, error) {
	return s.find(ctx, uuidOrID)
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

// Update updates an existing GitHub App (by UUID or numeric ID)
func (s *GitHubAppService) Update(ctx context.Context, uuidOrID string, req *models.GitHubAppUpdateRequest) error {
	id, err := s.resolveID(ctx, uuidOrID)
	if err != nil {
		return err
	}

	type response struct {
		Message string `json:"message"`
	}
	var resp response
	err = s.client.Patch(ctx, fmt.Sprintf("github-apps/%d", id), req, &resp)
	if err != nil {
		return fmt.Errorf("failed to update GitHub App %s: %w", uuidOrID, err)
	}
	return nil
}

// Delete deletes a GitHub App (by UUID or numeric ID)
func (s *GitHubAppService) Delete(ctx context.Context, uuidOrID string) error {
	id, err := s.resolveID(ctx, uuidOrID)
	if err != nil {
		return err
	}

	err = s.client.Delete(ctx, fmt.Sprintf("github-apps/%d", id))
	if err != nil {
		return fmt.Errorf("failed to delete GitHub App %s: %w", uuidOrID, err)
	}
	return nil
}

// ListRepositories lists all repositories accessible by a GitHub App (by UUID or numeric ID)
func (s *GitHubAppService) ListRepositories(ctx context.Context, uuidOrID string) ([]models.GitHubRepository, error) {
	id, err := s.resolveID(ctx, uuidOrID)
	if err != nil {
		return nil, err
	}

	type response struct {
		Repositories []models.GitHubRepository `json:"repositories"`
	}
	var resp response
	err = s.client.Get(ctx, fmt.Sprintf("github-apps/%d/repositories", id), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories for GitHub App %s: %w", uuidOrID, err)
	}
	return resp.Repositories, nil
}

// ListBranches lists all branches for a repository (app by UUID or numeric ID)
func (s *GitHubAppService) ListBranches(ctx context.Context, uuidOrID string, owner, repo string) ([]models.GitHubBranch, error) {
	id, err := s.resolveID(ctx, uuidOrID)
	if err != nil {
		return nil, err
	}

	type response struct {
		Branches []models.GitHubBranch `json:"branches"`
	}
	var resp response
	err = s.client.Get(ctx, fmt.Sprintf("github-apps/%d/repositories/%s/%s/branches", id, owner, repo), &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches for %s/%s: %w", owner, repo, err)
	}
	return resp.Branches, nil
}

// resolveID maps a UUID (or numeric ID) to the integer ID the API routes expect.
// Numeric input is used as-is without a list call.
func (s *GitHubAppService) resolveID(ctx context.Context, uuidOrID string) (int, error) {
	uuidOrID = strings.TrimSpace(uuidOrID)
	if id, err := strconv.Atoi(uuidOrID); err == nil {
		return id, nil
	}

	app, err := s.find(ctx, uuidOrID)
	if err != nil {
		return 0, err
	}
	return app.ID, nil
}

func (s *GitHubAppService) find(ctx context.Context, uuidOrID string) (*models.GitHubApp, error) {
	apps, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	uuidOrID = strings.TrimSpace(uuidOrID)
	for i := range apps {
		if apps[i].UUID == uuidOrID || strconv.Itoa(apps[i].ID) == uuidOrID {
			return &apps[i], nil
		}
	}

	return nil, fmt.Errorf("GitHub App %q not found", uuidOrID)
}

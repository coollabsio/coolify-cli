package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// TeamService handles team-related operations
type TeamService struct {
	client *api.Client
}

// NewTeamService creates a new team service
func NewTeamService(client *api.Client) *TeamService {
	return &TeamService{client: client}
}

// List retrieves all teams
func (s *TeamService) List(ctx context.Context) ([]models.Team, error) {
	var teams []models.Team
	err := s.client.Get(ctx, "teams", &teams)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	return teams, nil
}

// Get retrieves a team by ID
func (s *TeamService) Get(ctx context.Context, id string) (*models.Team, error) {
	var team models.Team
	err := s.client.Get(ctx, fmt.Sprintf("teams/%s", id), &team)
	if err != nil {
		return nil, fmt.Errorf("failed to get team %s: %w", id, err)
	}
	return &team, nil
}

// Current retrieves the team bound to the API token
func (s *TeamService) Current(ctx context.Context) (*models.Team, error) {
	var team models.Team
	err := s.client.Get(ctx, "team", &team)
	if err != nil {
		return nil, fmt.Errorf("failed to get token team: %w", err)
	}
	return &team, nil
}

// ListMembers retrieves members of a specific team
func (s *TeamService) ListMembers(ctx context.Context, teamID string) ([]models.TeamMember, error) {
	var members []models.TeamMember
	err := s.client.Get(ctx, fmt.Sprintf("teams/%s/members", teamID), &members)
	if err != nil {
		return nil, fmt.Errorf("failed to list members for team %s: %w", teamID, err)
	}
	return members, nil
}

// CurrentMembers retrieves members of the team bound to the API token
func (s *TeamService) CurrentMembers(ctx context.Context) ([]models.TeamMember, error) {
	var members []models.TeamMember
	err := s.client.Get(ctx, "team/members", &members)
	if err != nil {
		return nil, fmt.Errorf("failed to list token team members: %w", err)
	}
	return members, nil
}

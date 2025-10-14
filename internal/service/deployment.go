package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
)

// DeploymentService handles deployment-related operations
type DeploymentService struct {
	client *api.Client
}

// NewDeploymentService creates a new deployment service
func NewDeploymentService(client *api.Client) *DeploymentService {
	return &DeploymentService{
		client: client,
	}
}

// DeployResponse represents the response from a deploy operation
type DeployResponse struct {
	Message      string `json:"message"`
	DeploymentID string `json:"deployment_uuid,omitempty"`
}

// Deploy triggers a deployment for a resource
func (s *DeploymentService) Deploy(ctx context.Context, uuid string, force bool) (*DeployResponse, error) {
	endpoint := fmt.Sprintf("deploy?uuid=%s", uuid)
	if force {
		endpoint += "&force=true"
	}

	var response DeployResponse
	err := s.client.Get(ctx, endpoint, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy resource %s: %w", uuid, err)
	}
	return &response, nil
}

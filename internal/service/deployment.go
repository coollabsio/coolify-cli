package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
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

// DeploymentInfo represents a single deployment in the deploy response
type DeploymentInfo struct {
	Message        string `json:"message"`
	ResourceUUID   string `json:"resource_uuid"`
	DeploymentUUID string `json:"deployment_uuid"`
}

// DeployResponse represents the response from a deploy operation
type DeployResponse struct {
	Deployments []DeploymentInfo `json:"deployments"`
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

// List retrieves all deployments
func (s *DeploymentService) List(ctx context.Context) ([]models.Deployment, error) {
	var deployments []models.Deployment
	err := s.client.Get(ctx, "deployments", &deployments)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	return deployments, nil
}

// Get retrieves a deployment by UUID
func (s *DeploymentService) Get(ctx context.Context, uuid string) (*models.Deployment, error) {
	var deployment models.Deployment
	err := s.client.Get(ctx, fmt.Sprintf("deployments/%s", uuid), &deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s: %w", uuid, err)
	}
	return &deployment, nil
}

// CancelResponse represents the response from canceling a deployment
type CancelResponse struct {
	Message        string `json:"message"`
	DeploymentUUID string `json:"deployment_uuid"`
	Status         string `json:"status"`
}

// Cancel cancels an in-progress deployment
// Note: This endpoint will be available in a future version of Coolify
func (s *DeploymentService) Cancel(ctx context.Context, uuid string) (*CancelResponse, error) {
	var response CancelResponse
	err := s.client.Post(ctx, fmt.Sprintf("deployments/%s/cancel", uuid), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel deployment %s: %w", uuid, err)
	}
	return &response, nil
}

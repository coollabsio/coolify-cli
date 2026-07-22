package service

import (
	"context"
	"fmt"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// MCPService handles MCP server enable/disable operations
type MCPService struct {
	client *api.Client
}

// NewMCPService creates a new MCP service
func NewMCPService(client *api.Client) *MCPService {
	return &MCPService{client: client}
}

// Enable enables the MCP server (root team only)
func (s *MCPService) Enable(ctx context.Context) (*models.Response, error) {
	var resp models.Response
	err := s.client.Post(ctx, "mcp/enable", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to enable MCP server: %w", err)
	}
	return &resp, nil
}

// Disable disables the MCP server (root team only)
func (s *MCPService) Disable(ctx context.Context) (*models.Response, error) {
	var resp models.Response
	err := s.client.Post(ctx, "mcp/disable", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to disable MCP server: %w", err)
	}
	return &resp, nil
}

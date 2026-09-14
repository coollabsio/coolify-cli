package service

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// AuditEventListOptions filters audit events for the current token's team.
type AuditEventListOptions struct {
	Source  string
	Action  string
	Search  string
	PerPage int
	Page    int
}

// AuditEventService handles audit event queries.
type AuditEventService struct {
	client *api.Client
}

func NewAuditEventService(client *api.Client) *AuditEventService {
	return &AuditEventService{client: client}
}

func (s *AuditEventService) List(ctx context.Context, options AuditEventListOptions) (*models.AuditEventsPage, error) {
	query := url.Values{}
	if options.Source != "" {
		query.Set("source", options.Source)
	}
	if options.Action != "" {
		query.Set("action", options.Action)
	}
	if options.Search != "" {
		query.Set("search", options.Search)
	}
	if options.PerPage > 0 {
		query.Set("per_page", strconv.Itoa(options.PerPage))
	}
	if options.Page > 0 {
		query.Set("page", strconv.Itoa(options.Page))
	}

	endpoint := "audit-events"
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	var response models.AuditEventsPage
	if err := s.client.Get(ctx, endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to list audit events: %w", err)
	}

	return &response, nil
}

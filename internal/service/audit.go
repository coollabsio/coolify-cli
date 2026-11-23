package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// AuditService handles team audit log operations
type AuditService struct {
	client *api.Client
}

// NewAuditService creates a new audit service
func NewAuditService(client *api.Client) *AuditService {
	return &AuditService{client: client}
}

// ValidateAuditListOptions checks the filters against the API's validation rules.
func ValidateAuditListOptions(opts models.AuditEventListOptions) error {
	if opts.PerPage != 0 && (opts.PerPage < 1 || opts.PerPage > 100) {
		return fmt.Errorf("--per-page must be between 1 and 100")
	}
	if opts.Page < 0 {
		return fmt.Errorf("--page must be 1 or greater")
	}
	if len(opts.Search) > 255 {
		return fmt.Errorf("--search must not be longer than 255 characters")
	}
	if len(opts.Action) > 255 {
		return fmt.Errorf("--action must not be longer than 255 characters")
	}
	if opts.Source != "" && !slices.Contains(models.AuditEventSources, opts.Source) {
		return fmt.Errorf("--source must be one of: %s", strings.Join(models.AuditEventSources, ", "))
	}
	return nil
}

// List retrieves one page of the team's audit events (GET /audit-events)
func (s *AuditService) List(ctx context.Context, opts models.AuditEventListOptions) (*models.AuditEventPage, error) {
	if err := ValidateAuditListOptions(opts); err != nil {
		return nil, err
	}

	query := url.Values{}
	if opts.PerPage > 0 {
		query.Set("per_page", strconv.Itoa(opts.PerPage))
	}
	if opts.Page > 0 {
		query.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Search != "" {
		query.Set("search", opts.Search)
	}
	if opts.Action != "" {
		query.Set("action", opts.Action)
	}
	if opts.Source != "" {
		query.Set("source", opts.Source)
	}

	path := "audit-events"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var body json.RawMessage
	headers, err := s.client.GetWithHeaders(ctx, path, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit events: %w", err)
	}
	page, err := parseAuditEventPage(body, headers, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit events: %w", err)
	}
	return page, nil
}

// auditDefaultPerPage is the page size of GET /audit-events without per_page.
const auditDefaultPerPage = 25

// parseAuditEventPage reads both response shapes of GET /audit-events: a plain
// array with the X-Total-Count header, and the paginator object of older Coolify versions.
func parseAuditEventPage(body json.RawMessage, headers http.Header, opts models.AuditEventListOptions) (*models.AuditEventPage, error) {
	if !isJSONArray(body) {
		var page models.AuditEventPage
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		return &page, nil
	}

	page := models.AuditEventPage{CurrentPage: max(opts.Page, 1), PerPage: opts.PerPage}
	if page.PerPage <= 0 {
		page.PerPage = auditDefaultPerPage
	}
	if err := json.Unmarshal(body, &page.Data); err != nil {
		return nil, err
	}
	page.Total = len(page.Data)
	if total, err := strconv.Atoi(headers.Get("X-Total-Count")); err == nil {
		page.Total = total
	}
	page.LastPage = max(1, (page.Total+page.PerPage-1)/page.PerPage)
	return &page, nil
}

// isJSONArray reports whether a JSON document is an array.
func isJSONArray(body json.RawMessage) bool {
	trimmed := bytes.TrimSpace(body)
	return len(trimmed) > 0 && trimmed[0] == '['
}

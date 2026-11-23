package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// ProjectConfigService handles coolify.json (config as code) operations.
type ProjectConfigService struct {
	client *api.Client
}

// NewProjectConfigService creates a new project config service.
func NewProjectConfigService(client *api.Client) *ProjectConfigService {
	return &ProjectConfigService{client: client}
}

type configFileBody struct {
	File string `json:"file"`
}

// Export returns the coolify.json file of an environment, as the API sends it.
func (s *ProjectConfigService) Export(ctx context.Context, environmentUUID string) (json.RawMessage, error) {
	var file json.RawMessage
	if err := s.client.Get(ctx, fmt.Sprintf("environments/%s/config/export", url.PathEscape(environmentUUID)), &file); err != nil {
		return nil, fmt.Errorf("failed to export environment %s: %w", environmentUUID, err)
	}
	return file, nil
}

// Plan compares a coolify.json file with an environment. Nothing changes.
// A file that is not valid returns a plan that has only errors.
func (s *ProjectConfigService) Plan(ctx context.Context, environmentUUID, file string) (*models.ConfigPlan, error) {
	var plan models.ConfigPlan
	err := s.client.Post(ctx, fmt.Sprintf("environments/%s/config/plan", url.PathEscape(environmentUUID)), configFileBody{File: file}, &plan)
	if err == nil {
		return &plan, nil
	}
	var apiErr *api.Error
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnprocessableEntity {
		var body struct {
			Errors []models.ConfigError `json:"errors"`
		}
		// Laravel validation errors are an object, not a list: those stay an API error.
		if json.Unmarshal(apiErr.Body, &body) == nil && len(body.Errors) > 0 {
			return &models.ConfigPlan{Errors: body.Errors}, nil
		}
	}
	return nil, fmt.Errorf("failed to plan environment %s: %w", environmentUUID, withValidationDetails(err))
}

// CreateLink links an environment to coolify.json.
func (s *ProjectConfigService) CreateLink(ctx context.Context, req models.ConfigLinkCreateRequest) (*models.ConfigLink, error) {
	var link models.ConfigLink
	if err := s.client.PostOnce(ctx, "config/links", req, &link); err != nil {
		return nil, fmt.Errorf("failed to create config link: %w", withValidationDetails(err))
	}
	return &link, nil
}

// GetLink returns a config link.
func (s *ProjectConfigService) GetLink(ctx context.Context, linkUUID string) (*models.ConfigLink, error) {
	var link models.ConfigLink
	if err := s.client.Get(ctx, linkPath(linkUUID, ""), &link); err != nil {
		return nil, fmt.Errorf("failed to get config link %s: %w", linkUUID, err)
	}
	return &link, nil
}

// UpdateLink changes auto apply or the server of a config link.
func (s *ProjectConfigService) UpdateLink(ctx context.Context, linkUUID string, req models.ConfigLinkUpdateRequest) (*models.ConfigLink, error) {
	var link models.ConfigLink
	if err := s.client.Patch(ctx, linkPath(linkUUID, ""), req, &link); err != nil {
		return nil, fmt.Errorf("failed to update config link %s: %w", linkUUID, withValidationDetails(err))
	}
	return &link, nil
}

// DeleteLink disconnects an environment from coolify.json. The resources stay.
func (s *ProjectConfigService) DeleteLink(ctx context.Context, linkUUID string) error {
	if err := s.client.Delete(ctx, linkPath(linkUUID, "")); err != nil {
		return fmt.Errorf("failed to delete config link %s: %w", linkUUID, err)
	}
	return nil
}

// Revisions lists the revisions of a config link, newest first.
func (s *ProjectConfigService) Revisions(ctx context.Context, linkUUID string) ([]models.ConfigRevision, error) {
	var revisions []models.ConfigRevision
	if err := s.client.Get(ctx, linkPath(linkUUID, "/revisions"), &revisions); err != nil {
		return nil, fmt.Errorf("failed to list revisions of config link %s: %w", linkUUID, err)
	}
	return revisions, nil
}

// Apply plans and applies a file (api link) or the head of the branch (git
// link, file is nil). A failed revision is returned without an error, so the
// caller can show its reason.
func (s *ProjectConfigService) Apply(ctx context.Context, linkUUID string, file *string) (*models.ConfigRevision, error) {
	var body any
	if file != nil {
		body = configFileBody{File: *file}
	}
	var revision models.ConfigRevision
	err := s.client.PostOnce(ctx, linkPath(linkUUID, "/apply"), body, &revision)
	if err == nil {
		return &revision, nil
	}
	var apiErr *api.Error
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnprocessableEntity {
		if json.Unmarshal(apiErr.Body, &revision) == nil && revision.UUID != "" && revision.Status != "" {
			return &revision, nil
		}
	}
	return nil, fmt.Errorf("failed to apply config link %s: %w", linkUUID, withValidationDetails(err))
}

// Suspend pauses the sync of a config link.
func (s *ProjectConfigService) Suspend(ctx context.Context, linkUUID string) (*models.ConfigLink, error) {
	var link models.ConfigLink
	if err := s.client.PostOnce(ctx, linkPath(linkUUID, "/suspend"), nil, &link); err != nil {
		return nil, fmt.Errorf("failed to suspend config link %s: %w", linkUUID, err)
	}
	return &link, nil
}

// ResumePreview returns the plan that a resume would apply. Nothing changes.
func (s *ProjectConfigService) ResumePreview(ctx context.Context, linkUUID string) (*models.ConfigPlan, error) {
	var plan models.ConfigPlan
	if err := s.client.PostOnce(ctx, linkPath(linkUUID, "/resume"), nil, &plan); err != nil {
		return nil, fmt.Errorf("failed to preview resume of config link %s: %w", linkUUID, err)
	}
	return &plan, nil
}

// Resume resumes the sync of a config link and applies the latest file.
func (s *ProjectConfigService) Resume(ctx context.Context, linkUUID string) (*models.ConfigResumeResult, error) {
	var result models.ConfigResumeResult
	if err := s.client.PostOnce(ctx, linkPath(linkUUID, "/resume")+"?confirm=1", nil, &result); err != nil {
		return nil, fmt.Errorf("failed to resume config link %s: %w", linkUUID, err)
	}
	return &result, nil
}

// Approve approves a revision that waits for approval. deleteData lists the
// keys of removed resources whose data is deleted too.
func (s *ProjectConfigService) Approve(ctx context.Context, revisionUUID string, deleteData []string) (*models.ConfigRevision, error) {
	if deleteData == nil {
		deleteData = []string{}
	}
	body := struct {
		DeleteData []string `json:"delete_data"`
	}{DeleteData: deleteData}
	var revision models.ConfigRevision
	if err := s.client.PostOnce(ctx, revisionPath(revisionUUID, "/approve"), body, &revision); err != nil {
		return nil, fmt.Errorf("failed to approve revision %s: %w", revisionUUID, withValidationDetails(err))
	}
	return &revision, nil
}

// Release releases a stuck apply of a revision.
func (s *ProjectConfigService) Release(ctx context.Context, revisionUUID string) (*models.ConfigRevision, error) {
	var revision models.ConfigRevision
	if err := s.client.PostOnce(ctx, revisionPath(revisionUUID, "/release"), nil, &revision); err != nil {
		return nil, fmt.Errorf("failed to release revision %s: %w", revisionUUID, err)
	}
	return &revision, nil
}

// Retry plans the file of a revision again.
func (s *ProjectConfigService) Retry(ctx context.Context, revisionUUID string) (*models.ConfigRevision, error) {
	var revision models.ConfigRevision
	if err := s.client.PostOnce(ctx, revisionPath(revisionUUID, "/retry"), nil, &revision); err != nil {
		return nil, fmt.Errorf("failed to retry revision %s: %w", revisionUUID, err)
	}
	return &revision, nil
}

// withValidationDetails adds the field errors of a Laravel 422 response
// ({"errors": {"field": ["message"]}}) to the error message.
func withValidationDetails(err error) error {
	var apiErr *api.Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUnprocessableEntity {
		return err
	}
	var body struct {
		Errors map[string][]string `json:"errors"`
	}
	if json.Unmarshal(apiErr.Body, &body) != nil || len(body.Errors) == 0 {
		return err
	}
	fields := make([]string, 0, len(body.Errors))
	for field := range body.Errors {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	details := make([]string, 0, len(fields))
	for _, field := range fields {
		details = append(details, field+": "+strings.Join(body.Errors[field], " "))
	}
	return fmt.Errorf("%w (%s)", err, strings.Join(details, "; "))
}

func linkPath(linkUUID, suffix string) string {
	return "config/links/" + url.PathEscape(linkUUID) + suffix
}

func revisionPath(revisionUUID, suffix string) string {
	return "config/revisions/" + url.PathEscape(revisionUUID) + suffix
}

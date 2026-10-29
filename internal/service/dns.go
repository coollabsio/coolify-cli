package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// DNSResourceTypes lists the resource types whose domains can have managed DNS records.
var DNSResourceTypes = []string{"applications", "services"}

// DNSConflictError is returned when another DNS record already uses the hostname (HTTP 409).
// Pass RecordID and Current to Replace to overwrite it.
type DNSConflictError struct {
	Message  string
	Conflict models.DNSConflict
}

func (e *DNSConflictError) Error() string {
	return fmt.Sprintf("%s The record %s points to %s; Coolify would point it to %s.",
		e.Message, e.Conflict.RecordID, e.Conflict.Current, e.Conflict.Proposed)
}

// DNSService manages Cloudflare DNS records for the domains of applications and services.
type DNSService struct {
	client *api.Client
}

// NewDNSService creates a new DNS service
func NewDNSService(client *api.Client) *DNSService {
	return &DNSService{client: client}
}

func dnsPath(resourceType, uuid string) (string, error) {
	if !slices.Contains(DNSResourceTypes, resourceType) {
		return "", fmt.Errorf("unsupported resource type %q", resourceType)
	}
	return resourceType + "/" + url.PathEscape(uuid) + "/dns", nil
}

// Status lists the DNS providers and managed records of each domain (GET /{type}/{uuid}/dns).
func (s *DNSService) Status(ctx context.Context, resourceType, uuid string) (*models.ResourceDNSStatus, error) {
	path, err := dnsPath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var resp models.ResourceDNSStatus
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("failed to get DNS status of %s: %w", uuid, err)
	}
	return &resp, nil
}

// Create creates the DNS record of a domain (POST /{type}/{uuid}/dns). A conflict returns a *DNSConflictError.
func (s *DNSService) Create(ctx context.Context, resourceType, uuid string, req models.DNSRecordCreateRequest) (*models.ManagedDNSRecord, error) {
	if strings.TrimSpace(req.Hostname) == "" {
		return nil, fmt.Errorf("--hostname is required")
	}
	path, err := dnsPath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var resp models.DNSRecordResponse
	if err := s.client.PostOnce(ctx, path, req, &resp); err != nil {
		return nil, describeDNSError("failed to create DNS record for "+req.Hostname, err)
	}
	return &resp.Record, nil
}

// Replace points the conflicting record of a create at the server (POST /{type}/{uuid}/dns/replace).
func (s *DNSService) Replace(ctx context.Context, resourceType, uuid string, req models.DNSRecordReplaceRequest) (*models.ManagedDNSRecord, error) {
	switch {
	case strings.TrimSpace(req.Hostname) == "":
		return nil, fmt.Errorf("--hostname is required")
	case strings.TrimSpace(req.RecordID) == "":
		return nil, fmt.Errorf("--record-id is required")
	case strings.TrimSpace(req.Current) == "":
		return nil, fmt.Errorf("--current is required")
	}
	path, err := dnsPath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var resp models.DNSRecordResponse
	if err := s.client.PostOnce(ctx, path+"/replace", req, &resp); err != nil {
		return nil, describeDNSError("failed to replace DNS record for "+req.Hostname, err)
	}
	return &resp.Record, nil
}

// Release releases the managed record of a hostname the resource no longer uses
// (DELETE /{type}/{uuid}/dns/{hostname}). With deleteRecord false, the record stays at Cloudflare.
func (s *DNSService) Release(ctx context.Context, resourceType, uuid, hostname string, deleteRecord bool) (*models.DNSReleaseResult, error) {
	if strings.TrimSpace(hostname) == "" {
		return nil, fmt.Errorf("the hostname is required")
	}
	path, err := dnsPath(resourceType, uuid)
	if err != nil {
		return nil, err
	}
	var resp models.DNSReleaseResult
	query := fmt.Sprintf("/%s?delete_record=%t", url.PathEscape(strings.ToLower(hostname)), deleteRecord)
	if err := s.client.DeleteWithResult(ctx, path+query, &resp); err != nil {
		return nil, describeDNSError("failed to release DNS record of "+hostname, err)
	}
	return &resp, nil
}

func describeDNSError(prefix string, err error) error {
	var apiErr *api.Error
	if errors.As(err, &apiErr) && apiErr.StatusCode == 409 {
		var body struct {
			Message  string             `json:"message"`
			Conflict models.DNSConflict `json:"conflict"`
		}
		if json.Unmarshal(apiErr.Body, &body) == nil && body.Conflict.RecordID != "" {
			return &DNSConflictError{Message: body.Message, Conflict: body.Conflict}
		}
	}
	return fmt.Errorf("%s: %w", prefix, describeValidationErrors(err))
}

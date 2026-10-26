package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

// ServerRegistryService manages Docker registry logins on a server.
type ServerRegistryService struct {
	client *api.Client
}

// NewServerRegistryService creates a new server registry service.
func NewServerRegistryService(client *api.Client) *ServerRegistryService {
	return &ServerRegistryService{client: client}
}

var registrySchemePattern = regexp.MustCompile(`^[a-z][a-z0-9+.-]*://`)

// NormalizeRegistry turns user input such as "https://GHCR.io/" into the registry
// host the API expects ("ghcr.io"). It mirrors the server-side normalization so the
// value can be used as a single URL path segment.
func NormalizeRegistry(registry string) string {
	host := strings.ToLower(strings.TrimSpace(registry))
	host = registrySchemePattern.ReplaceAllString(host, "")
	host, _, _ = strings.Cut(host, "/")
	return host
}

func serverRegistriesPath(serverUUID string) string {
	return "servers/" + url.PathEscape(serverUUID) + "/registries"
}

func serverRegistryPath(serverUUID, registry string) string {
	return serverRegistriesPath(serverUUID) + "/" + url.PathEscape(NormalizeRegistry(registry))
}

// List returns the registries the server is logged in to and the registries its resources use.
func (s *ServerRegistryService) List(ctx context.Context, serverUUID string) (*models.ServerRegistriesResponse, error) {
	var out models.ServerRegistriesResponse
	if err := s.client.Get(ctx, serverRegistriesPath(serverUUID), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Login runs docker login on the server. The password is sent only in the JSON body.
func (s *ServerRegistryService) Login(ctx context.Context, serverUUID string, req models.ServerRegistryLoginRequest) (*models.Response, error) {
	req.Registry = NormalizeRegistry(req.Registry)
	var out models.Response
	if err := s.client.Post(ctx, serverRegistriesPath(serverUUID), req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Check verifies that the saved login of the server for a registry still works.
func (s *ServerRegistryService) Check(ctx context.Context, serverUUID, registry string) (*models.Response, error) {
	var out models.Response
	if err := s.client.Post(ctx, serverRegistryPath(serverUUID, registry)+"/check", map[string]any{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Logout runs docker logout on the server for a registry.
func (s *ServerRegistryService) Logout(ctx context.Context, serverUUID, registry string) error {
	return s.client.Delete(ctx, serverRegistryPath(serverUUID, registry))
}

// ServerRegistryStatus returns the human readable login status of a registry.
func ServerRegistryStatus(registry models.ServerRegistry) string {
	switch {
	case !registry.LoggedIn:
		return models.ServerRegistryStatusNotLoggedIn
	case registry.Source != nil && *registry.Source == models.ServerRegistrySourceCredHelpers:
		return models.ServerRegistryStatusCredentialHelper
	default:
		return models.ServerRegistryStatusLoggedIn
	}
}

// ServerRegistryUsageSummary summarizes the resources using a registry, for example
// "3 applications, 1 database". Types are listed in order of first appearance.
func ServerRegistryUsageSummary(users []models.ServerRegistryUser) string {
	counts := map[string]int{}
	var order []string
	for _, user := range users {
		kind := strings.ToLower(strings.TrimSpace(user.Type))
		if kind == "" {
			kind = "resource"
		}
		if counts[kind] == 0 {
			order = append(order, kind)
		}
		counts[kind]++
	}

	parts := make([]string, 0, len(order))
	for _, kind := range order {
		label := kind
		if counts[kind] != 1 {
			label += "s"
		}
		parts = append(parts, fmt.Sprintf("%d %s", counts[kind], label))
	}
	return strings.Join(parts, ", ")
}

// ServerRegistryRows converts registries into table rows. Usernames are replaced
// with mask unless showSensitive is set; empty usernames stay empty.
func ServerRegistryRows(registries []models.ServerRegistry, showSensitive bool, mask string) []models.ServerRegistryRow {
	rows := make([]models.ServerRegistryRow, 0, len(registries))
	for _, registry := range registries {
		username := ""
		if registry.Username != nil && *registry.Username != "" {
			username = *registry.Username
			if !showSensitive {
				username = mask
			}
		}
		rows = append(rows, models.ServerRegistryRow{
			Registry: registry.Registry,
			Status:   ServerRegistryStatus(registry),
			Username: username,
			UsedBy:   ServerRegistryUsageSummary(registry.UsedBy),
		})
	}
	return rows
}

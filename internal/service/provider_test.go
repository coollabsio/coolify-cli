package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coollabsio/coolify-cli/internal/api"
	"github.com/coollabsio/coolify-cli/internal/models"
)

func TestProviderServices_OptionsUseTokenQueryAndCreationUsesPost(t *testing.T) {
	type observed struct {
		method string
		uri    string
		body   map[string]any
	}
	// 6 Hetzner GETs + 1 POST + 4 DO GETs + 1 POST + 4 Vultr GETs + 1 POST = 17
	requests := make(chan observed, 32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		item := observed{method: r.Method, uri: r.URL.RequestURI()}
		if r.Method == http.MethodPost {
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&item.body))
		}
		requests <- item
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[]`))
		} else {
			_, _ = w.Write([]byte(`{"uuid":"server-1","ip":"192.0.2.1"}`))
		}
	}))
	defer server.Close()
	ctx := context.Background()
	// Disable retries so a single bad response cannot flood the request channel.
	client := api.NewClient(server.URL, "token", api.WithRetries(0))

	hetzner := NewHetznerService(client)
	_, _ = hetzner.Locations(ctx, "token/id")
	_, _ = hetzner.ServerTypes(ctx, "token/id")
	_, _ = hetzner.Images(ctx, "token/id")
	_, _ = hetzner.SSHKeys(ctx, "token/id")
	_, _ = hetzner.Firewalls(ctx, "token/id")
	_, _ = hetzner.Networks(ctx, "token/id")
	_, _ = hetzner.Create(ctx, models.HetznerServerCreateRequest{CloudProviderTokenUUID: "token-1", Location: "nbg1", ServerType: "cx22", Image: 1, PrivateKeyUUID: "key-1", EnableIPv4: true, EnableIPv6: true, EnableBackups: true, HetznerFirewallIDs: []int{2}, HetznerNetworkIDs: []int{3}})

	digitalOcean := NewDigitalOceanService(client)
	_, _ = digitalOcean.Regions(ctx, "token/id")
	_, _ = digitalOcean.Sizes(ctx, "token/id")
	_, _ = digitalOcean.Images(ctx, "token/id")
	_, _ = digitalOcean.SSHKeys(ctx, "token/id")
	_, _ = digitalOcean.Create(ctx, models.DigitalOceanServerCreateRequest{CloudProviderTokenUUID: "token-1", Region: "nyc3", Size: "s-1vcpu", Image: "ubuntu-24-04-x64", PrivateKeyUUID: "key-1"})

	vultr := NewVultrService(client)
	_, _ = vultr.Regions(ctx, "token/id")
	_, _ = vultr.Plans(ctx, "token/id")
	_, _ = vultr.OperatingSystems(ctx, "token/id")
	_, _ = vultr.SSHKeys(ctx, "token/id")
	_, _ = vultr.Create(ctx, models.VultrServerCreateRequest{CloudProviderTokenUUID: "token-1", Region: "ewr", Plan: "vc2", OSID: 2284, PrivateKeyUUID: "key-1"})

	for i := 0; i < 6; i++ {
		item := <-requests
		assert.Equal(t, http.MethodGet, item.method)
		assert.Contains(t, item.uri, "cloud_provider_token_uuid=token%2Fid")
	}
	hetznerCreate := <-requests
	assert.Equal(t, http.MethodPost, hetznerCreate.method)
	assert.Equal(t, "/api/v1/servers/hetzner", hetznerCreate.uri)
	assert.Equal(t, true, hetznerCreate.body["enable_backups"])
	assert.Equal(t, []any{float64(2)}, hetznerCreate.body["hetzner_firewall_ids"])
	assert.Equal(t, []any{float64(3)}, hetznerCreate.body["hetzner_network_ids"])

	for i := 0; i < 4; i++ {
		item := <-requests
		assert.Equal(t, http.MethodGet, item.method)
		assert.Contains(t, item.uri, "cloud_provider_token_uuid=token%2Fid")
	}
	assert.Equal(t, "/api/v1/servers/digitalocean", (<-requests).uri)
	for i := 0; i < 4; i++ {
		item := <-requests
		assert.Equal(t, http.MethodGet, item.method)
		assert.Contains(t, item.uri, "cloud_provider_token_uuid=token%2Fid")
	}
	assert.Equal(t, "/api/v1/servers/vultr", (<-requests).uri)
}

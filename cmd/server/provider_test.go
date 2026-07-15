package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerCommand_RegistersProviderTreesWithoutReplacingAdd(t *testing.T) {
	cmd := NewServerCommand()
	for _, name := range []string{"add", "hetzner", "digitalocean", "vultr", "destinations"} {
		child, _, err := cmd.Find([]string{name})
		require.NoError(t, err)
		assert.Equal(t, name, child.Name())
	}

	providerCases := map[string][]string{
		"hetzner":      {"locations", "server-types", "images", "ssh-keys", "firewalls", "networks", "create"},
		"digitalocean": {"regions", "sizes", "images", "ssh-keys", "create"},
		"vultr":        {"regions", "plans", "os", "ssh-keys", "create"},
	}
	for provider, subcommands := range providerCases {
		for _, subcommand := range subcommands {
			child, _, err := cmd.Find([]string{provider, subcommand})
			require.NoError(t, err)
			assert.Equal(t, subcommand, child.Name())
		}
	}
}

func TestProviderCreateCommands_RegisterSchemaFlags(t *testing.T) {
	hetzner := NewHetznerCommand()
	create, _, err := hetzner.Find([]string{"create"})
	require.NoError(t, err)
	for _, flag := range []string{"cloud-token", "location", "server-type", "image", "private-key", "enable-ipv4", "enable-ipv6", "enable-backups", "ssh-key-ids", "firewall-ids", "network-ids", "cloud-init", "validate"} {
		assert.NotNil(t, create.Flags().Lookup(flag), flag)
	}

	digitalOcean := NewDigitalOceanCommand()
	create, _, err = digitalOcean.Find([]string{"create"})
	require.NoError(t, err)
	for _, flag := range []string{"cloud-token", "region", "size", "image", "private-key", "enable-ipv6", "monitoring", "ssh-key-ids", "cloud-init", "validate"} {
		assert.NotNil(t, create.Flags().Lookup(flag), flag)
	}

	vultr := NewVultrCommand()
	create, _, err = vultr.Find([]string{"create"})
	require.NoError(t, err)
	for _, flag := range []string{"cloud-token", "region", "plan", "os-id", "private-key", "enable-ipv6", "disable-public-ipv4", "ssh-key-ids", "cloud-init", "validate"} {
		assert.NotNil(t, create.Flags().Lookup(flag), flag)
	}
}

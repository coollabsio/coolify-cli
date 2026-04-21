package firewall

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewFirewallCommand_Subcommands(t *testing.T) {
	cmd := NewFirewallCommand()
	assert.Equal(t, "firewall", cmd.Use)
	subs := map[string]*cobra.Command{}
	for _, s := range cmd.Commands() {
		subs[s.Use] = s
	}
	assert.Contains(t, subs, "containers")
	assert.Contains(t, subs, "list")
	assert.Contains(t, subs, "allow")
	assert.Contains(t, subs, "revoke")
}

func TestNewFirewallCommand_PersistentFlags(t *testing.T) {
	cmd := NewFirewallCommand()
	pf := cmd.PersistentFlags()
	for _, name := range []string{"servers", "ssh-key", "ssh-user", "ssh-port",
		"concurrency", "ssh-timeout", "namespace", "all-namespaces",
		"coold-token", "coold-port", "wg-interface"} {
		assert.NotNil(t, pf.Lookup(name), "missing --%s", name)
	}
	// Replaced by --namespace; must be gone.
	assert.Nil(t, pf.Lookup("podman-network"))
}

func TestAllowCommand_LocalFlags(t *testing.T) {
	cmd := NewFirewallCommand()
	var allow *cobra.Command
	for _, s := range cmd.Commands() {
		if s.Use == "allow" {
			allow = s
			break
		}
	}
	if allow == nil {
		t.Fatal("allow subcommand not found")
	}
	for _, name := range []string{"from", "to", "port", "proto", "bidirectional"} {
		assert.NotNil(t, allow.Flags().Lookup(name), "missing --%s on allow", name)
	}
}

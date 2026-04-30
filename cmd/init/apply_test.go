package initcmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewInitCommand verifies the command tree structure.
func TestNewInitCommand(t *testing.T) {
	cmd := NewInitCommand()

	assert.Equal(t, "init", cmd.Use)
	assert.NotEmpty(t, cmd.Short)

	subCmds := map[string]*cobra.Command{}
	for _, sub := range cmd.Commands() {
		subCmds[sub.Use] = sub
	}
	assert.Contains(t, subCmds, "plan")
	assert.Contains(t, subCmds, "bootstrap")
	assert.Contains(t, subCmds, "extend")
	assert.Contains(t, subCmds, "upgrade")
	assert.NotContains(t, subCmds, "apply", "apply removed in favor of bootstrap/extend/upgrade")
}

// TestNewInitCommand_PersistentFlags verifies shared flags are registered.
func TestNewInitCommand_PersistentFlags(t *testing.T) {
	cmd := NewInitCommand()
	pf := cmd.PersistentFlags()

	assert.NotNil(t, pf.Lookup("servers"))
	assert.NotNil(t, pf.Lookup("ssh-key"))
	assert.NotNil(t, pf.Lookup("ssh-user"))
	assert.NotNil(t, pf.Lookup("ssh-port"))
	assert.NotNil(t, pf.Lookup("wg-mgmt-pool"))
	assert.NotNil(t, pf.Lookup("container-pool"))
	assert.NotNil(t, pf.Lookup("container-prefix"))
	assert.NotNil(t, pf.Lookup("wg-interface"))
	assert.NotNil(t, pf.Lookup("wg-listen-port"))
	assert.NotNil(t, pf.Lookup("namespaces"))
	assert.NotNil(t, pf.Lookup("skip-default-deny"))
	assert.NotNil(t, pf.Lookup("concurrency"))
	assert.NotNil(t, pf.Lookup("ssh-timeout"))
	assert.NotNil(t, pf.Lookup("yes"))
	// Old flags removed.
	assert.Nil(t, pf.Lookup("wg-pool"))
	assert.Nil(t, pf.Lookup("wg-host-prefix"))
	assert.Nil(t, pf.Lookup("wg-subnet"))
	assert.Nil(t, pf.Lookup("podman"))
	assert.Nil(t, pf.Lookup("default-deny"))
	assert.Nil(t, pf.Lookup("install-coold"))
	// Replaced by --namespaces.
	assert.Nil(t, pf.Lookup("podman-network"))
}

// TestNewInitCommand_FlagDefaults verifies default values.
func TestNewInitCommand_FlagDefaults(t *testing.T) {
	cmd := NewInitCommand()
	pf := cmd.PersistentFlags()

	user, err := pf.GetString("ssh-user")
	require.NoError(t, err)
	assert.Equal(t, "root", user)

	port, err := pf.GetInt("ssh-port")
	require.NoError(t, err)
	assert.Equal(t, 22, port)

	mgmtPool, err := pf.GetString("wg-mgmt-pool")
	require.NoError(t, err)
	assert.Equal(t, "100.64.0.0/16", mgmtPool)

	contPool, err := pf.GetString("container-pool")
	require.NoError(t, err)
	assert.Equal(t, "10.210.0.0/16", contPool)

	contPrefix, err := pf.GetInt("container-prefix")
	require.NoError(t, err)
	assert.Equal(t, 24, contPrefix)

	iface, err := pf.GetString("wg-interface")
	require.NoError(t, err)
	assert.Equal(t, "wg0", iface)

	listenPort, err := pf.GetInt("wg-listen-port")
	require.NoError(t, err)
	assert.Equal(t, 51820, listenPort)

	namespaces, err := pf.GetStringSlice("namespaces")
	require.NoError(t, err)
	assert.Equal(t, []string{"default"}, namespaces)

	skipDefaultDeny, err := pf.GetBool("skip-default-deny")
	require.NoError(t, err)
	assert.False(t, skipDefaultDeny)

	concurrency, err := pf.GetInt("concurrency")
	require.NoError(t, err)
	assert.Equal(t, 10, concurrency)

	timeout, err := pf.GetString("ssh-timeout")
	require.NoError(t, err)
	assert.Equal(t, "30s", timeout)
}

// TestPlanCommand_FlagsInherited verifies that plan inherits parent persistent flags.
func TestPlanCommand_FlagsInherited(t *testing.T) {
	init := NewInitCommand()
	_ = init.ParseFlags([]string{})

	var planCmd *cobra.Command
	for _, sub := range init.Commands() {
		if sub.Use == "plan" {
			planCmd = sub
			break
		}
	}
	require.NotNil(t, planCmd)

	f := planCmd.InheritedFlags().Lookup("servers")
	assert.NotNil(t, f, "plan should inherit --servers from parent")
}

// Package firewall implements the `coolify firewall` command tree: a test
// harness for the v5 cross-host allow-rule flow that will eventually be
// owned by the coold agent. See CONTROL_PLANE.md §3.
package firewall

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/common"
)

// FirewallFlags is the shared flag set for every `coolify firewall`
// subcommand: SSH plumbing (via embed) + podman network name.
type FirewallFlags struct {
	common.SSHMeshFlags
	PodmanNetworkName string
}

// bindFirewallFlags registers the persistent flags on the parent command.
func bindFirewallFlags(cmd *cobra.Command, f *FirewallFlags) {
	common.BindSSHMeshFlags(cmd, &f.SSHMeshFlags)
	cmd.PersistentFlags().StringVar(&f.PodmanNetworkName, "podman-network",
		"coolify-mesh",
		"Podman bridge network name (must match --podman-network used at init)")
}

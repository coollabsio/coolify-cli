// Package initcmd implements the `coolify init` alpha WireGuard mesh
// bootstrap command tree (Coolify v5).
package initcmd

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/common"
)

// InitFlags holds all flags shared between `plan` and `apply`.
type InitFlags struct {
	common.SSHMeshFlags
	common.MeshNetFlags

	WGMgmtPool          string
	WGInterface         string
	WGListenPort        int
	SkipDefaultDeny     bool
	CooldVersion        string
	CorrosionVersion    string
	CorrosionGossipPort int
	CorrosionAPIPort    int
	Yes                 bool
}

// bindInitFlags registers all shared flags as PersistentFlags on cmd.
func bindInitFlags(cmd *cobra.Command, f *InitFlags) {
	common.BindSSHMeshFlags(cmd, &f.SSHMeshFlags)
	common.BindMeshNetMultiFlags(cmd, &f.MeshNetFlags)

	pf := cmd.PersistentFlags()

	pf.StringVar(&f.WGMgmtPool, "wg-mgmt-pool", "100.64.0.0/16",
		"WireGuard management address pool — each host gets a /32 from here, assigned to wg0")
	pf.StringVar(&f.WGInterface, "wg-interface", "wg0",
		"WireGuard interface name on the remote hosts")
	pf.IntVar(&f.WGListenPort, "wg-listen-port", 51820,
		"WireGuard UDP listen port")
	pf.BoolVar(&f.SkipDefaultDeny, "skip-default-deny", false,
		"Skip installing the default-deny firewall scaffold. By default, both cross-host and intra-host (same bridge) container traffic is blocked; coold manages the allow list at runtime")
	pf.StringVar(&f.CooldVersion, "coold-version", "nightly",
		`Release tag to download for coold (e.g. "nightly", "v1.2.3"). nightly always re-installs on every apply.`)
	pf.StringVar(&f.CorrosionVersion, "corrosion-version", "nightly",
		`Release tag to download for corrosion (e.g. "nightly", "v1.2.3"). nightly always re-installs on every apply.`)
	pf.IntVar(&f.CorrosionGossipPort, "corrosion-gossip-port", 8787,
		"Corrosion SWIM gossip port (bound to the wg0 mgmt IP)")
	pf.IntVar(&f.CorrosionAPIPort, "corrosion-api-port", 8080,
		"Corrosion HTTP API port (bound to 127.0.0.1)")
	pf.BoolVarP(&f.Yes, "yes", "y", false,
		"Skip the interactive alpha confirmation prompt")
}

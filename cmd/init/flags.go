// Package initcmd implements the `coolify init` alpha WireGuard mesh
// bootstrap command tree (Coolify v5).
package initcmd

import (
	"os"

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
	CooldBinaryPath     string
	CorrosionBinaryPath string
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
		"Skip installing the default-deny iptables scaffold (COOLIFY-INTRA / COOLIFY-ALLOW). By default, cross-host container traffic is blocked except where coold installs allow rules. Intra-host (same bridge) traffic is NOT enforced — defer to per-app podman networks")
	pf.StringVar(&f.CooldBinaryPath, "coold-binary",
		os.ExpandEnv("$HOME/devel/coold/target/release/coold"),
		"Local path to the coold Linux/arm64 binary")
	pf.StringVar(&f.CorrosionBinaryPath, "corrosion-binary",
		os.ExpandEnv("$HOME/devel/corrosion/target/release/corrosion"),
		"Local path to the corrosion Linux/arm64 binary")
	pf.IntVar(&f.CorrosionGossipPort, "corrosion-gossip-port", 8787,
		"Corrosion SWIM gossip port (bound to the wg0 mgmt IP)")
	pf.IntVar(&f.CorrosionAPIPort, "corrosion-api-port", 8080,
		"Corrosion HTTP API port (bound to 127.0.0.1)")
	pf.BoolVarP(&f.Yes, "yes", "y", false,
		"Skip the interactive alpha confirmation prompt")
}

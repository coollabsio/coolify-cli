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

	WGMgmtPool            string
	ContainerPool         string
	ContainerPrefix       int
	WGInterface           string
	WGListenPort          int
	InstallPodman         bool
	PodmanNetworkName     string
	DefaultDenyContainers bool
	InstallCoold          bool
	CooldBinaryPath       string
	CorrosionBinaryPath   string
	CorrosionGossipPort   int
	CorrosionAPIPort      int
	Yes                   bool
}

// bindInitFlags registers all shared flags as PersistentFlags on cmd.
func bindInitFlags(cmd *cobra.Command, f *InitFlags) {
	common.BindSSHMeshFlags(cmd, &f.SSHMeshFlags)

	pf := cmd.PersistentFlags()

	pf.StringVar(&f.WGMgmtPool, "wg-mgmt-pool", "100.64.0.0/16",
		"WireGuard management address pool — each host gets a /32 from here, assigned to wg0")
	pf.StringVar(&f.ContainerPool, "container-pool", "10.210.0.0/16",
		"Container address pool — each host gets a /<container-prefix> from here, owned by the Podman bridge")
	pf.IntVar(&f.ContainerPrefix, "container-prefix", 24,
		"Prefix length of the per-host container subnet (e.g. 24 → /24, 254 usable container IPs per host)")
	pf.StringVar(&f.WGInterface, "wg-interface", "wg0",
		"WireGuard interface name on the remote hosts")
	pf.IntVar(&f.WGListenPort, "wg-listen-port", 51820,
		"WireGuard UDP listen port")
	pf.BoolVar(&f.InstallPodman, "podman", false,
		"Install Podman, enable its socket, create a per-host bridge network, install firewall rules, and enable IP forwarding")
	pf.StringVar(&f.PodmanNetworkName, "podman-network", "coolify-mesh",
		"Name of the Podman bridge network created on each host (requires --podman)")
	pf.BoolVar(&f.DefaultDenyContainers, "default-deny", false,
		"With --podman: install default-deny iptables rules for CROSS-HOST container traffic (between hosts via wg0). Intra-host (same bridge) traffic is NOT enforced — defer to per-app podman networks. The v5 control plane manages allows in the COOLIFY-ALLOW chain on the host that owns each destination IP")
	pf.BoolVar(&f.InstallCoold, "install-coold", false,
		"Install the Coolify v5 control-plane agents (corrosion + coold). Requires --podman. Uploads the binaries from --corrosion-binary / --coold-binary")
	pf.StringVar(&f.CooldBinaryPath, "coold-binary",
		os.ExpandEnv("$HOME/devel/coold/target/release/coold"),
		"Local path to the coold Linux/arm64 binary (used with --install-coold)")
	pf.StringVar(&f.CorrosionBinaryPath, "corrosion-binary",
		os.ExpandEnv("$HOME/devel/corrosion/target/release/corrosion"),
		"Local path to the corrosion Linux/arm64 binary (used with --install-coold)")
	pf.IntVar(&f.CorrosionGossipPort, "corrosion-gossip-port", 8787,
		"Corrosion SWIM gossip port (bound to the wg0 mgmt IP)")
	pf.IntVar(&f.CorrosionAPIPort, "corrosion-api-port", 8080,
		"Corrosion HTTP API port (bound to 127.0.0.1)")
	pf.BoolVarP(&f.Yes, "yes", "y", false,
		"Skip the interactive alpha confirmation prompt")
}

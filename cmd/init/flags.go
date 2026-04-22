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

	// CentralHost is the SSH address of the central VM (from --central flag).
	// When non-empty, phases 4+5 install Redis + broker on that host and push
	// per-host JWTs to all other hosts. Default empty = no broker setup.
	CentralHost   string
	BrokerVersion string

	// EnableBuilder is a cluster-wide shorthand: when true (and BuilderHosts
	// is empty), every host in Servers is enrolled as builder-capable. When
	// BuilderHosts is non-empty, EnableBuilder is ignored and only the
	// listed subset gets the capability.
	EnableBuilder bool

	// BuilderHosts is an explicit list of SSH addresses (subset of Servers)
	// to enroll with the builder capability. Empty = fall back to
	// EnableBuilder semantics. Mutually exclusive in practice with
	// EnableBuilder=false (leaves builder fully disabled).
	BuilderHosts    []string
	BuilderCapacity int
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
	pf.StringVar(&f.CentralHost, "central", "",
		`SSH address of the central VM that will run broker + Redis (and later Laravel).
Must be one of the --servers entries. When set, phases 4+5 install Redis + broker on that host
and push a per-host JWT to every other server. Leave empty to skip broker setup.`)
	pf.StringVar(&f.BrokerVersion, "broker-version", "nightly",
		`Release tag to download for broker (e.g. "nightly", "v1.2.3").`)
	pf.BoolVar(&f.EnableBuilder, "enable-builder", true,
		`Cluster-wide shorthand: enable the builder capability on every host
(requires --central). Ignored when --builder-hosts is set.`)
	pf.StringSliceVar(&f.BuilderHosts, "builder-hosts", nil,
		`Explicit subset of --servers to enroll with the builder capability.
Takes precedence over --enable-builder. Empty (default) means fall back to
--enable-builder for the whole cluster.`)
	pf.IntVar(&f.BuilderCapacity, "builder-capacity", 2,
		"Concurrent builds accepted per host (COOLD_BUILDER_CAPACITY).")
}

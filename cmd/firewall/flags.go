// Package firewall implements the `coolify firewall` command tree. It is a
// thin SSH-bounced client for the coold agent's REST API: `allow` / `revoke`
// / `list` POST/DELETE/GET against coold on the destination host, while
// `containers` stays SSH+podman because coold has no container surface.
// See CONTROL_PLANE.md §3.
package firewall

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/cmd/common"
	ifw "github.com/coollabsio/coolify-cli/internal/firewall"
)

// FirewallFlags is the shared flag set for every `coolify firewall`
// subcommand: SSH plumbing (via embed) + podman network name + coold REST
// endpoint/token.
type FirewallFlags struct {
	common.SSHMeshFlags
	PodmanNetworkName string

	// CooldToken is an optional bearer-token override for coold's REST API.
	// When unset (and COOLIFY_COOLD_TOKEN env is unset), the CLI SSHes into
	// each host and reads /etc/coolify/api-token instead — tokens are
	// generated per-host at install time and are not centrally shared.
	CooldToken string
	// CooldPort is the TCP port coold listens on (bound to the WG mgmt IP).
	// Must match COOLD_API_BIND emitted by internal/services/coold.go.
	CooldPort int
	// WGInterface is the WireGuard interface name used to discover coold's
	// bind IP on each host. Must match --wg-interface used at `coolify init`.
	WGInterface string
}

// bindFirewallFlags registers the persistent flags on the parent command.
func bindFirewallFlags(cmd *cobra.Command, f *FirewallFlags) {
	common.BindSSHMeshFlags(cmd, &f.SSHMeshFlags)
	pf := cmd.PersistentFlags()
	pf.StringVar(&f.PodmanNetworkName, "podman-network",
		"coolify-mesh",
		"Podman bridge network name (must match --podman-network used at init)")
	pf.StringVar(&f.CooldToken, "coold-token", "",
		"Bearer token override for coold REST API (also reads COOLIFY_COOLD_TOKEN env). "+
			"When unset, CLI reads /etc/coolify/api-token over SSH per host.")
	pf.IntVar(&f.CooldPort, "coold-port", 8443,
		"TCP port coold's REST API listens on (bound to the WG mgmt IP)")
	pf.StringVar(&f.WGInterface, "wg-interface", ifw.DefaultWGInterface,
		"WireGuard interface name on remote hosts (must match --wg-interface at init)")
}

// ResolveCooldToken returns the bearer-token override supplied via flag or
// env, or "" when neither is set. Callers treat an empty string as "no
// override — SSH-fetch the per-host token instead".
func (f *FirewallFlags) ResolveCooldToken() (string, error) {
	if f.CooldToken != "" {
		return f.CooldToken, nil
	}
	if env := os.Getenv("COOLIFY_COOLD_TOKEN"); env != "" {
		return env, nil
	}
	return "", nil
}

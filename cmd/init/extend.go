package initcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/wireguard"
)

// NewExtendCommand creates the `coolify init extend` subcommand. It adds the
// hosts listed in --new-hosts to an existing mesh: new hosts get the full
// first-time install; existing hosts get only peer-refresh actions (WG
// AllowedIPs update, corrosion config refresh, firewall unit reinstall if
// namespace list changed). Destructive actions on existing hosts are blocked
// unless --allow-replace is set.
func NewExtendCommand(flags *InitFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extend",
		Short: "Add new hosts to an existing mesh (existing hosts stay untouched)",
		Long: `Extend an existing mesh with brand-new hosts. --new-hosts lists the
subset of --servers that is brand-new; those hosts receive the full
first-time install (install WG, generate keys, install podman, install
coold/corrosion, create bridges, etc.).

Existing hosts in --servers are re-probed and get only the peer-refresh
actions required to route traffic to the new peer: WG config rewrite,
corrosion peer list refresh, firewall unit reinstall when the namespace
list changed. Agent binaries are not re-downloaded on existing hosts —
use ` + "`coolify init upgrade`" + ` for that.

--allow-replace unlocks destructive-replace actions (e.g. recreating a
drifted podman bridge) on existing hosts. Handle with care: containers
on a recreated bridge are disconnected.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(flags.NewHosts) == 0 {
				return fmt.Errorf("--new-hosts is required: list the subset of --servers that is brand-new")
			}
			servers := make(map[string]struct{}, len(flags.Servers))
			for _, s := range flags.Servers {
				servers[s] = struct{}{}
			}
			for _, nh := range flags.NewHosts {
				if _, ok := servers[nh]; !ok {
					return fmt.Errorf("--new-hosts: %q is not in --servers", nh)
				}
			}

			flags.Intent = string(wireguard.IntentExtend)

			header := fmt.Sprintf("Extending mesh with %d new host(s): %v", len(flags.NewHosts), flags.NewHosts)
			return runApply(cmd.Context(), cmd, flags, applyOptions{
				SkipAlphaGate: true,
				Header:        header,
			})
		},
	}

	cmd.Flags().StringSliceVar(&flags.NewHosts, "new-hosts", nil,
		"Comma-separated subset of --servers that is brand-new this run (required). Only these hosts receive the full first-time install; all other hosts get peer-refresh only.")
	cmd.Flags().BoolVar(&flags.AllowReplace, "allow-replace", false,
		"Unlock destructive-replace actions on existing hosts (e.g. recreating a drifted podman bridge). Off by default — drifted existing hosts are surfaced as skipped actions instead.")

	return cmd
}

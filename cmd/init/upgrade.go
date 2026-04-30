package initcmd

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/wireguard"
)

// NewUpgradeCommand creates the `coolify init upgrade` subcommand: bumps
// coold/corrosion/scheduler/builder binaries across every host. Does not touch
// WG config, podman networks, firewall rules, or the corrosion schema. Rejects
// "nightly" version tags unless --allow-nightly is set.
func NewUpgradeCommand(flags *InitFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Bump agent binary versions (coold / corrosion / scheduler / builder) on every host",
		Long: `Upgrade the agent binaries managed by coolify init across every host in
--servers. Only binary-fetch actions and their follow-up service restarts
run; WG config, podman networks, firewall rules, and the corrosion schema
are left untouched.

Pin each binary with --coold-version / --corrosion-version /
--scheduler-version. "nightly" is rejected by default because it forces a
re-install on every run; pass --allow-nightly to override.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			flags.Intent = string(wireguard.IntentUpgrade)
			return runApply(cmd.Context(), cmd, flags, applyOptions{
				SkipAlphaGate: true,
				Header:        "Upgrading agent binaries...",
			})
		},
	}

	cmd.Flags().BoolVar(&flags.AllowNightly, "allow-nightly", false,
		"Permit --coold-version/--corrosion-version/--scheduler-version=nightly. Off by default because nightly re-installs on every run instead of only when the pinned version changes.")

	return cmd
}

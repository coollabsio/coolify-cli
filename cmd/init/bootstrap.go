package initcmd

import (
	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/wireguard"
)

// NewBootstrapCommand creates the `coolify init bootstrap` subcommand — the
// first-time mesh install. Runs every applicable action on every host and
// keeps the interactive alpha gate (unless --yes / non-TTY / env override).
func NewBootstrapCommand(flags *InitFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "First-time mesh install (all actions allowed)",
		Long: `Bootstrap a fresh WireGuard + Podman + coold mesh across every host in
--servers. Idempotent: re-running with no changes produces an empty plan.

Use this for the initial install. For adding hosts later, see
` + "`coolify init extend`" + `; for bumping agent versions, see
` + "`coolify init upgrade`" + `.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			flags.Intent = string(wireguard.IntentBootstrap)
			return runApply(cmd.Context(), cmd, flags, applyOptions{
				Header: "Bootstrapping mesh...",
			})
		},
	}
}

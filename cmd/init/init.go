package initcmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const alphaBanner = `
[ALPHA] coolify init targets Coolify v5 and is experimental.
[ALPHA] WireGuard mesh bootstrap requires root/sudo and modifies network configuration.
[ALPHA] Test in non-production environments first. Stability is not guaranteed.
`

// NewInitCommand creates the parent `coolify init` command.
// The command tree is kept for tests and future v5 work but is not registered
// on the public root CLI.
// On bare invocation (no subcommand) it prints the alpha banner and help.
func NewInitCommand() *cobra.Command {
	flags := &InitFlags{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "[ALPHA] Initialize WireGuard mesh for Coolify v5",
		Long: `[ALPHA] Bootstrap a WireGuard full-mesh overlay between servers and
provision each host with the Coolify v5 runtime stack: Podman + bridge
network, default-deny iptables scaffold, and the coold/corrosion
control-plane agents.

Subcommands:
  plan       Show what would change without touching anything (--intent
             selects the filter: bootstrap / extend / upgrade).
  bootstrap  First-time install (all actions allowed).
  extend     Add new hosts to an existing mesh; existing hosts get only
             peer-refresh actions.
  upgrade    Bump agent versions (coold / corrosion / scheduler / builder);
             WG / podman / firewall untouched.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprint(os.Stderr, alphaBanner)
			return cmd.Help()
		},
	}

	bindInitFlags(cmd, flags)

	cmd.AddCommand(NewPlanCommand(flags))
	cmd.AddCommand(NewBootstrapCommand(flags))
	cmd.AddCommand(NewExtendCommand(flags))
	cmd.AddCommand(NewUpgradeCommand(flags))

	return cmd
}

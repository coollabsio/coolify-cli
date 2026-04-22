package firewall

import (
	"github.com/spf13/cobra"
)

// NewFirewallCommand creates the parent `coolify firewall` command.
// On bare invocation (no subcommand) it prints help.
func NewFirewallCommand() *cobra.Command {
	flags := &Flags{}

	cmd := &cobra.Command{
		Use:   "firewall",
		Short: "[ALPHA] Manage cross-host container allow rules (Coolify v5)",
		Long: `[ALPHA] Manage the COOLIFY-ALLOW iptables chain installed by
"coolify init --podman --default-deny". This is a test harness for the v5
control-plane firewall flow: it SSHes into every server, discovers running
containers on the Coolify mesh bridge (override with --podman-network), and
lets you add/remove cross-host allow rules.

Subcommands:
  containers  List discovered containers across the mesh.
  list        Show installed allow rules.
  allow       Add an allow rule (src container → dst container:port).
  revoke      Remove an allow rule.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	bindFlags(cmd, flags)

	cmd.AddCommand(newContainersCommand(flags))
	cmd.AddCommand(newListCommand(flags))
	cmd.AddCommand(newAllowCommand(flags))
	cmd.AddCommand(newRevokeCommand(flags))

	return cmd
}

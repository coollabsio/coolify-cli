package firewall

import (
	"context"

	ifw "github.com/coollabsio/coolify-cli/internal/firewall"
	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// discoverAllViaPkg is a thin wrapper around ifw.DiscoverAll that threads
// the FirewallFlags in. Exists so command files stay short.
func discoverAllViaPkg(
	ctx context.Context,
	runner ssh.Runner,
	flags *FirewallFlags,
) ([]ifw.Container, []ssh.ServerResult[[]ifw.Container]) {
	return ifw.DiscoverAll(ctx, runner, flags.Servers, flags.SSHUser, flags.SSHPort,
		flags.PodmanNetworkName, flags.Concurrency)
}

// listAllViaPkg wraps ifw.ListAll for the same reason.
func listAllViaPkg(
	ctx context.Context,
	runner ssh.Runner,
	flags *FirewallFlags,
) ([]ifw.AllowRule, []ssh.ServerResult[[]ifw.AllowRule]) {
	return ifw.ListAll(ctx, runner, flags.Servers, flags.SSHUser, flags.SSHPort,
		flags.Concurrency)
}

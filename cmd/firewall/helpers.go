package firewall

import (
	"context"
	"sync"

	ifw "github.com/coollabsio/coolify-cli/internal/firewall"
	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// discoverAllViaPkg is a thin wrapper around ifw.DiscoverAll that threads
// the FirewallFlags in. Used by `containers` (SSH+podman) and by `allow` /
// `revoke` for endpoint resolution; `list` goes straight to coold REST.
func discoverAllViaPkg(
	ctx context.Context,
	runner ssh.Runner,
	flags *FirewallFlags,
) ([]ifw.Container, []ssh.ServerResult[[]ifw.Container]) {
	return ifw.DiscoverAll(ctx, runner, flags.Servers, flags.SSHUser, flags.SSHPort,
		flags.PodmanNetworkName, flags.Concurrency)
}

// tokenResolver returns a closure that hands out coold bearer tokens
// per-host. Precedence: explicit --coold-token (or COOLIFY_COOLD_TOKEN env)
// wins for every host; otherwise SSH into the host once and cache the
// contents of /etc/coolify/api-token. The cache is goroutine-safe so the
// closure can be passed straight into CooldListAll's fanout.
func tokenResolver(
	ctx context.Context,
	runner ssh.Runner,
	flags *FirewallFlags,
) func(host string) (string, error) {
	if override, _ := flags.ResolveCooldToken(); override != "" {
		return func(string) (string, error) { return override, nil }
	}
	var (
		mu    sync.Mutex
		cache = map[string]string{}
	)
	return func(host string) (string, error) {
		mu.Lock()
		if tok, ok := cache[host]; ok {
			mu.Unlock()
			return tok, nil
		}
		mu.Unlock()
		tok, err := ifw.FetchCooldToken(ctx, runner, host, flags.SSHUser, flags.SSHPort)
		if err != nil {
			return "", err
		}
		mu.Lock()
		cache[host] = tok
		mu.Unlock()
		return tok, nil
	}
}

package firewall

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/coollabsio/coolify-cli/cmd/common"
	ifw "github.com/coollabsio/coolify-cli/internal/firewall"
	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// discoverAllViaPkg is a thin wrapper around ifw.DiscoverAll /
// ifw.DiscoverAllNamespaces that threads the FirewallFlags in. Used by
// `containers` (SSH+podman) and by `allow` / `revoke` for endpoint
// resolution; `list` goes straight to coold REST.
//
// When AllNamespaces is set, the fanout walks every supplied namespace; the
// caller (containers subcommand) is responsible for enumerating which
// namespaces exist on the hosts — absent that, falls back to the selected
// single namespace.
func discoverAllViaPkg(
	ctx context.Context,
	runner ssh.Runner,
	flags *FirewallFlags,
) ([]ifw.Container, []ssh.ServerResult[[]ifw.Container]) {
	return ifw.DiscoverAll(ctx, runner, flags.Servers, flags.SSHUser,
		flags.SSHPort, flags.Namespace, flags.PodmanNetworkName(),
		flags.Concurrency)
}

// discoverAcrossNamespaces runs DiscoverAllNamespaces for every supplied
// namespace. Network name is derived from common.PodmanNetworkFor so the
// caller only has to supply the namespace list.
func discoverAcrossNamespaces(
	ctx context.Context,
	runner ssh.Runner,
	flags *FirewallFlags,
	namespaces []string,
) ([]ifw.Container, []ssh.ServerResult[[]ifw.Container]) {
	return ifw.DiscoverAllNamespaces(ctx, runner, flags.Servers,
		flags.SSHUser, flags.SSHPort, namespaces,
		common.PodmanNetworkFor, flags.Concurrency)
}

// discoverNamespacesOnHosts SSHes into every host and lists every podman
// network carrying the io.coolify.managed=true label, collecting the unique
// io.coolify.namespace label values. Used by `containers --all-namespaces`.
// Returns the per-host results so host-level failures surface as warnings
// instead of aborting the fanout.
func discoverNamespacesOnHosts(
	ctx context.Context,
	runner ssh.Runner,
	flags *FirewallFlags,
) ([]string, []ssh.ServerResult[[]string], error) {
	// `podman network ls`'s `{{.Labels}}` renders as a comma-separated `k=v`
	// string (not a map, unlike `podman network inspect`), so `index` can't be
	// used — pull `io.coolify.namespace=<val>` out with sed instead.
	script := `podman network ls --filter label=io.coolify.managed=true ` +
		`--format '{{.Labels}}' 2>/dev/null | ` +
		`sed -n 's/.*io\.coolify\.namespace=\([^,]*\).*/\1/p' || true`
	results := ssh.ForEachServer(ctx, flags.Servers, flags.Concurrency,
		func(ctx context.Context, host string) ([]string, error) {
			stdout, _, err := runner.Run(ctx, host, flags.SSHUser,
				flags.SSHPort, script)
			if err != nil {
				return nil, err
			}
			var nss []string
			for _, line := range strings.Split(stdout, "\n") {
				ns := strings.TrimSpace(line)
				if ns != "" {
					nss = append(nss, ns)
				}
			}
			return nss, nil
		})
	seen := map[string]struct{}{}
	for _, r := range results {
		for _, ns := range r.Result {
			seen[ns] = struct{}{}
		}
	}
	// Always probe the selected namespace too — caller may have just created
	// it and we haven't seen it on any host yet.
	seen[flags.Namespace] = struct{}{}
	all := make([]string, 0, len(seen))
	for ns := range seen {
		all = append(all, ns)
	}
	sort.Strings(all)
	return all, results, nil
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

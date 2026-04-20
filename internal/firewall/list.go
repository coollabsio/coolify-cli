package firewall

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// ListAllow returns every rule currently installed in COOLIFY-ALLOW on host.
// Missing chain (e.g. default-deny not installed yet) yields an empty slice.
func ListAllow(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	port int,
) ([]AllowRule, error) {
	cmd := "iptables -S " + ChainName + " 2>/dev/null || true"
	stdout, _, err := runner.Run(ctx, host, user, port, cmd)
	if err != nil {
		return nil, fmt.Errorf("list %s on %s: %w", ChainName, host, err)
	}
	var out []AllowRule
	for _, line := range strings.Split(stdout, "\n") {
		r, ok := ParseChainLine(line)
		if !ok {
			continue
		}
		r.Host = host
		out = append(out, r)
	}
	return out, nil
}

// ListAll runs ListAllow on every host in parallel. Returns the flattened
// slice sorted by (host, src, dst, port) plus per-host results so callers
// can surface partial failures.
func ListAll(
	ctx context.Context,
	runner ssh.Runner,
	hosts []string,
	user string,
	port int,
	concurrency int,
) ([]AllowRule, []ssh.ServerResult[[]AllowRule]) {
	results := ssh.ForEachServer(ctx, hosts, concurrency,
		func(ctx context.Context, host string) ([]AllowRule, error) {
			return ListAllow(ctx, runner, host, user, port)
		})
	var all []AllowRule
	for _, r := range results {
		all = append(all, r.Result...)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Host != all[j].Host {
			return all[i].Host < all[j].Host
		}
		si, sj := all[i].Src.String(), all[j].Src.String()
		if si != sj {
			return si < sj
		}
		di, dj := all[i].Dst.String(), all[j].Dst.String()
		if di != dj {
			return di < dj
		}
		return all[i].Port < all[j].Port
	})
	return all, results
}

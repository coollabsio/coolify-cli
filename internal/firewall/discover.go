package firewall

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// Container is a single running podman container on one mesh host.
type Container struct {
	Host string // SSH host the container runs on
	ID   string // short (12-char) podman ID
	Name string // podman container name
	IP   net.IP // IP on the coolify-mesh bridge network
}

// discoverScript prints one `id|name|ip` line per running container on the
// target network. Piped through `podman inspect` to resolve the per-network
// IP because `podman ps` doesn't surface that directly. `|| true` keeps the
// script from erroring when podman is absent or the network has no members.
func discoverScript(networkName string) string {
	return fmt.Sprintf(
		`podman ps --filter network=%[1]s --format '{{.ID}}|{{.Names}}' 2>/dev/null | `+
			`while IFS='|' read id name; do `+
			`  [ -z "$id" ] && continue; `+
			`  ip=$(podman inspect --format '{{(index .NetworkSettings.Networks %[2]q).IPAddress}}' "$id" 2>/dev/null); `+
			`  printf '%%s|%%s|%%s\n' "$id" "$name" "$ip"; `+
			`done || true`,
		networkName, networkName)
}

// ParseDiscoverLine parses one `id|name|ip` line from discoverScript.
// Returns (_, false) when the line is blank or malformed.
func ParseDiscoverLine(line string) (id, name string, ip net.IP, ok bool) {
	parts := strings.SplitN(strings.TrimSpace(line), "|", 3)
	if len(parts) != 3 {
		return "", "", nil, false
	}
	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", nil, false
	}
	ip = net.ParseIP(parts[2])
	if ip == nil {
		return "", "", nil, false
	}
	id = parts[0]
	if len(id) > 12 {
		id = id[:12]
	}
	return id, parts[1], ip, true
}

// DiscoverContainers SSHes into host and returns every container on
// networkName with its bridge IP.
func DiscoverContainers(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	port int,
	networkName string,
) ([]Container, error) {
	stdout, _, err := runner.Run(ctx, host, user, port, discoverScript(networkName))
	if err != nil {
		return nil, fmt.Errorf("discover containers on %s: %w", host, err)
	}
	var out []Container
	for _, line := range strings.Split(stdout, "\n") {
		id, name, ip, ok := ParseDiscoverLine(line)
		if !ok {
			continue
		}
		out = append(out, Container{Host: host, ID: id, Name: name, IP: ip})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Host != out[j].Host {
			return out[i].Host < out[j].Host
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// DiscoverAll runs DiscoverContainers across every host in parallel.
// Returns a flattened, sort-stable slice plus the per-host results so
// callers can surface partial failures.
func DiscoverAll(
	ctx context.Context,
	runner ssh.Runner,
	hosts []string,
	user string,
	port int,
	networkName string,
	concurrency int,
) ([]Container, []ssh.ServerResult[[]Container]) {
	results := ssh.ForEachServer(ctx, hosts, concurrency,
		func(ctx context.Context, host string) ([]Container, error) {
			return DiscoverContainers(ctx, runner, host, user, port, networkName)
		})
	var all []Container
	for _, r := range results {
		all = append(all, r.Result...)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Host != all[j].Host {
			return all[i].Host < all[j].Host
		}
		return all[i].Name < all[j].Name
	})
	return all, results
}

package wireguard

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// Probe SSHes into host and reads its current WireGuard + Podman state.
// All commands use `|| true` so a missing package or interface never
// causes a non-zero exit that would abort the probe.
func Probe(ctx context.Context, runner ssh.Runner, host, user string, port int, iface, podmanNetName string) (*ServerState, error) {
	state := &ServerState{
		Host:      host,
		Interface: iface,
	}

	// 1. Check if WireGuard is installed.
	stdout, _, _ := runner.Run(ctx, host, user, port,
		`dpkg-query -W -f='${Status}' wireguard 2>/dev/null | grep -c 'install ok installed' || echo 0`)
	if strings.TrimSpace(stdout) == "1" {
		state.Installed = true
	}

	// 2. Read public key.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`cat /etc/wireguard/publickey 2>/dev/null || true`)
	if pk := strings.TrimSpace(stdout); pk != "" {
		state.PublicKey = pk
		state.KeysExist = true
	}

	// 3. Parse the config file for management IP and peer list.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		fmt.Sprintf(`cat /etc/wireguard/%s.conf 2>/dev/null || true`, iface))
	if strings.TrimSpace(stdout) != "" {
		parseConfigFile(state, stdout)
	}

	// 4. Check if WG interface is currently up.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		fmt.Sprintf(`wg show %s dump 2>/dev/null || true`, iface))
	if strings.TrimSpace(stdout) != "" {
		state.Active = true
	}

	// 5. Podman package installed.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`dpkg-query -W -f='${Status}' podman 2>/dev/null | grep -c 'install ok installed' || echo 0`)
	if strings.TrimSpace(stdout) == "1" {
		state.PodmanInstalled = true
	}

	// 6. podman.socket active.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`systemctl is-active podman.socket 2>/dev/null || true`)
	if strings.TrimSpace(stdout) == "active" {
		state.PodmanSocketActive = true
	}

	// 7. Podman network exists + read its subnet (only if Podman installed).
	if state.PodmanInstalled && podmanNetName != "" {
		stdout, _, _ = runner.Run(ctx, host, user, port,
			fmt.Sprintf(`podman network exists %s 2>/dev/null && echo yes || echo no`, podmanNetName))
		if strings.TrimSpace(stdout) == "yes" {
			state.PodmanNetExists = true
			// Read the subnet so the allocator can preserve stable assignments.
			stdout, _, _ = runner.Run(ctx, host, user, port,
				fmt.Sprintf(`podman network inspect %s -f '{{(index .Subnets 0).Subnet}}' 2>/dev/null || true`, podmanNetName))
			if s := strings.TrimSpace(stdout); s != "" {
				if _, n, err := net.ParseCIDR(s); err == nil {
					state.ContainerSubnet = n
				}
			}
		}
	}

	// 8. IP forwarding enabled.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`sysctl -n net.ipv4.ip_forward 2>/dev/null || echo 0`)
	if strings.TrimSpace(stdout) == "1" {
		state.IPForwardEnabled = true
	}

	// 9. coolify-mesh-fw.service active.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`systemctl is-active coolify-mesh-fw.service 2>/dev/null || true`)
	if strings.TrimSpace(stdout) == "active" {
		state.FirewallActive = true
	}

	// 10. Default-deny scaffold present (COOLIFY-INTRA chain ends in DROP).
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`iptables -nL COOLIFY-INTRA 2>/dev/null | grep -q DROP && echo yes || echo no`)
	if strings.TrimSpace(stdout) == "yes" {
		state.DefaultDenyActive = true
	}

	// 11. Corrosion binary installed.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`test -x /usr/local/bin/corrosion && echo yes || echo no`)
	if strings.TrimSpace(stdout) == "yes" {
		state.CorrosionInstalled = true
	}

	// 12. Corrosion systemd service active.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`systemctl is-active corrosion 2>/dev/null || true`)
	if strings.TrimSpace(stdout) == "active" {
		state.CorrosionActive = true
	}

	// 13. Corrosion config hash (empty when missing).
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`sha256sum /etc/corrosion/config.toml 2>/dev/null | awk '{print $1}' || true`)
	if h := strings.TrimSpace(stdout); h != "" {
		state.CorrosionConfigHash = h
	}

	// 14. Corrosion schema file present.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`test -f /etc/corrosion/schemas/coolify.sql && echo yes || echo no`)
	if strings.TrimSpace(stdout) == "yes" {
		state.CorrosionSchemaExists = true
	}

	// 14a. sha256 of remote schema file (empty when absent). Used to detect
	// schema revisions so a new schema triggers re-write + DB reset.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`sha256sum /etc/corrosion/schemas/coolify.sql 2>/dev/null | awk '{print $1}' || true`)
	if h := strings.TrimSpace(stdout); h != "" {
		state.CorrosionSchemaSha256 = h
	}

	// 15. Coold binary installed.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`test -x /usr/local/bin/coold && echo yes || echo no`)
	if strings.TrimSpace(stdout) == "yes" {
		state.CooldInstalled = true
	}

	// 15a. sha256 of remote corrosion binary (empty when absent).
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`sha256sum /usr/local/bin/corrosion 2>/dev/null | awk '{print $1}' || true`)
	if h := strings.TrimSpace(stdout); h != "" {
		state.CorrosionBinarySha256 = h
	}

	// 15b. sha256 of remote coold binary (empty when absent).
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`sha256sum /usr/local/bin/coold 2>/dev/null | awk '{print $1}' || true`)
	if h := strings.TrimSpace(stdout); h != "" {
		state.CooldBinarySha256 = h
	}

	// 15c. sha256 of remote coold.service unit (empty when absent).
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`sha256sum /etc/systemd/system/coold.service 2>/dev/null | awk '{print $1}' || true`)
	if h := strings.TrimSpace(stdout); h != "" {
		state.CooldUnitSha256 = h
	}

	// 16. Coold systemd service active.
	stdout, _, _ = runner.Run(ctx, host, user, port,
		`systemctl is-active coold 2>/dev/null || true`)
	if strings.TrimSpace(stdout) == "active" {
		state.CooldActive = true
	}

	return state, nil
}

// Reconstruct runs Probe on every host in parallel and assembles a MeshState.
func Reconstruct(
	ctx context.Context,
	runner ssh.Runner,
	hosts []string,
	user string,
	port int,
	iface string,
	podmanNetName string,
	concurrency int,
) (MeshState, error) {
	results := ssh.ForEachServer(ctx, hosts, concurrency, func(ctx context.Context, host string) (*ServerState, error) {
		return Probe(ctx, runner, host, user, port, iface, podmanNetName)
	})

	mesh := MeshState{Servers: make(map[string]*ServerState, len(hosts))}
	var errs []string
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", r.Host, r.Err))
			mesh.Servers[r.Host] = &ServerState{Host: r.Host, Interface: iface}
			continue
		}
		mesh.Servers[r.Host] = r.Result
	}

	if len(errs) > 0 {
		return mesh, fmt.Errorf("probe errors:\n  %s", strings.Join(errs, "\n  "))
	}
	return mesh, nil
}

// parseConfigFile extracts WireGuard management IP, listen port, and peer list
// from the text content of /etc/wireguard/<iface>.conf.
func parseConfigFile(state *ServerState, content string) {
	var (
		inInterface bool
		inPeer      bool
		currentPeer Peer
	)

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		switch strings.ToLower(line) {
		case "[interface]":
			inInterface = true
			inPeer = false
			continue
		case "[peer]":
			if inPeer {
				state.Peers = append(state.Peers, currentPeer)
				currentPeer = Peer{}
			}
			inInterface = false
			inPeer = true
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if inInterface {
			switch strings.ToLower(key) {
			case "address":
				// Parse the host portion of "<ip>/<prefix>"; this is the
				// actual management IP, not the network address.
				ip, _, err := net.ParseCIDR(value)
				if err == nil {
					state.WireGuardMgmtIP = ip.To4()
				}
			case "listenport":
				if p, err := strconv.Atoi(value); err == nil {
					state.ListenPort = p
				}
			}
		}

		if inPeer {
			switch strings.ToLower(key) {
			case "publickey":
				currentPeer.PublicKey = value
			case "endpoint":
				currentPeer.Endpoint = value
			case "allowedips":
				for _, a := range strings.Split(value, ",") {
					currentPeer.AllowedIPs = append(currentPeer.AllowedIPs, strings.TrimSpace(a))
				}
			case "presharedkey":
				currentPeer.PresharedKey = value
			case "persistentkeepalive":
				if n, err := strconv.Atoi(value); err == nil {
					currentPeer.PersistentKeepalive = n
				}
			}
		}
	}

	if inPeer && currentPeer.PublicKey != "" {
		state.Peers = append(state.Peers, currentPeer)
	}
}

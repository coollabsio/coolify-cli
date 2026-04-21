package wireguard

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/services"
	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// ActionResult pairs a PlannedAction with its execution outcome.
type ActionResult struct {
	Action PlannedAction
	Err    error
}

// VerifyResult holds the post-apply verification for one server.
type VerifyResult struct {
	Host        string
	WireGuardIP net.IP
	PeerCount   int
	Active      bool
	Err         error
}

const aptInstallCmd = `DEBIAN_FRONTEND=noninteractive apt-get update -qq 2>/dev/null && ` +
	`DEBIAN_FRONTEND=noninteractive apt-get install -y ` +
	`-o Dpkg::Options::="--force-confold" ` +
	`wireguard wireguard-tools 2>&1`

const podmanInstallCmd = `DEBIAN_FRONTEND=noninteractive apt-get update -qq 2>/dev/null && ` +
	`DEBIAN_FRONTEND=noninteractive apt-get install -y ` +
	`-o Dpkg::Options::="--force-confold" ` +
	`podman 2>&1`

// enablePodmanSocketCmd ensures /run/podman/podman.sock exists via systemd
// socket activation. The socket is NEVER exposed on TCP — it stays a Unix
// socket on the host so the per-host coold agent can bind-mount it and
// proxy a curated REST API over wg0. See CONTROL_PLANE.md §2 + §12.
const enablePodmanSocketCmd = `systemctl enable --now podman.socket 2>&1`

const enableIPForwardCmd = `sysctl -w net.ipv4.ip_forward=1 && ` +
	`mkdir -p /etc/sysctl.d && ` +
	`echo 'net.ipv4.ip_forward=1' > /etc/sysctl.d/99-coolify-mesh.conf`

// podmanNetCreateCmd creates a per-namespace Podman bridge network. Idempotent:
// skips if the network already exists. The bridge gateway is MachineIP(subnet)
// (the .1 of the subnet).
//
// --disable-dns prevents netavark from starting aardvark-dns on the bridge
// gateway IP:53 — coold owns that socket for cluster-wide service discovery
// (see CONTROL_PLANE.md §5). Labels mark the network as ours + carry its
// namespace so `podman network inspect` drift checks can assert it.
func podmanNetCreateCmd(name, namespace string, subnet *net.IPNet, gateway net.IP) string {
	return fmt.Sprintf(
		`podman network exists %s 2>/dev/null && echo "network exists, skipping" || `+
			`podman network create --driver bridge --disable-dns `+
			`--label io.coolify.managed=true --label io.coolify.namespace=%s `+
			`--subnet=%s --gateway=%s %s`,
		name, namespace, subnet, gateway, name)
}

// podmanNetRecreateCmd drops and recreates a per-namespace Podman bridge
// network to clear drift (dns_enabled=true, subnet mismatch, missing label).
// Uses `rm -f` to detach any attached containers first.
func podmanNetRecreateCmd(name, namespace string, subnet *net.IPNet, gateway net.IP) string {
	return fmt.Sprintf(
		`podman network rm -f %s 2>&1 && `+
			`podman network create --driver bridge --disable-dns `+
			`--label io.coolify.managed=true --label io.coolify.namespace=%s `+
			`--subnet=%s --gateway=%s %s`,
		name, namespace, subnet, gateway, name)
}

// runStep executes a single shell command on a remote host, appends an
// ActionResult to out, and returns an error if the command failed.
func runStep(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	port int,
	out *[]ActionResult,
	atype ActionType,
	namespace, cmd, errFmt string,
) error {
	stdout, stderr, err := runner.Run(ctx, host, user, port, cmd)
	detail := ""
	if err != nil {
		detail = firstLine(stderr)
		if detail == "" {
			detail = firstLine(stdout)
		}
		if detail == "" {
			detail = err.Error()
		}
	}
	*out = append(*out, ActionResult{
		Action: PlannedAction{Host: host, Namespace: namespace, Type: atype, Detail: detail},
		Err:    err,
	})
	if err != nil {
		return fmt.Errorf(errFmt+": %w", err)
	}
	return nil
}

// ApplyMesh executes the mesh convergence in two phases:
//
//   - Phase 1 (per-server, parallel): install WG + Podman, generate keypair,
//     enable podman socket + IP forwarding.
//   - Re-probe to collect fresh public keys.
//   - Phase 2 (per-server, parallel): write WG config, enable/reload service,
//     create per-namespace Podman networks, install firewall service.
//   - Phase 3 (per-server, parallel, optional): download + enable corrosion/coold.
func ApplyMesh(
	ctx context.Context,
	runner ssh.Runner,
	user string,
	port int,
	desired *DesiredMesh,
	current MeshState,
	concurrency int,
) ([]ActionResult, error) {
	var results []ActionResult

	p1 := ssh.ForEachServer(ctx, desired.Hosts, concurrency,
		func(ctx context.Context, host string) ([]ActionResult, error) {
			return phase1Server(ctx, runner, host, user, port, desired, current)
		})

	phase1Failed := false
	for _, r := range p1 {
		results = append(results, r.Result...)
		if r.Err != nil {
			phase1Failed = true
		}
	}

	if phase1Failed {
		return results, fmt.Errorf("phase 1 (install/keygen) failed on one or more servers; aborting")
	}

	fresh, err := Reconstruct(ctx, runner, desired.Hosts, user, port,
		desired.Interface, desired.Namespaces, concurrency)
	if err != nil {
		return results, fmt.Errorf("re-probe after phase 1: %w", err)
	}

	mgmtAssignments, _, err := AllocateMgmtIPs(desired.MgmtPool, fresh.AssignedMgmtIPs(), desired.Hosts)
	if err != nil {
		return results, fmt.Errorf("mgmt IP allocation: %w", err)
	}
	containerAssignments, _, err := AllocateNamespaced(desired.ContainerPool, desired.ContainerPrefix,
		fresh.AssignedContainerSubnets(), desired.Namespaces, desired.Hosts)
	if err != nil {
		return results, fmt.Errorf("container subnet allocation: %w", err)
	}

	p2 := ssh.ForEachServer(ctx, desired.Hosts, concurrency,
		func(ctx context.Context, host string) ([]ActionResult, error) {
			return phase2Server(ctx, runner, host, user, port, desired, fresh, mgmtAssignments, containerAssignments)
		})

	for _, r := range p2 {
		results = append(results, r.Result...)
		if r.Err != nil {
			err = fmt.Errorf("phase 2 failed on one or more servers")
		}
	}

	if desired.InstallCoold && err == nil {
		p3 := ssh.ForEachServer(ctx, desired.Hosts, concurrency,
			func(ctx context.Context, host string) ([]ActionResult, error) {
				return phase3Server(ctx, runner, host, user, port,
					desired, fresh, mgmtAssignments, containerAssignments)
			})

		for _, r := range p3 {
			results = append(results, r.Result...)
			if r.Err != nil {
				err = fmt.Errorf("phase 3 failed on one or more servers")
			}
		}
	}

	return results, err
}

// phase1Server installs WireGuard, generates a keypair, and (if requested)
// installs Podman, enables its socket, and enables IP forwarding.
func phase1Server(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	port int,
	desired *DesiredMesh,
	current MeshState,
) ([]ActionResult, error) {
	state, ok := current.Servers[host]
	if !ok {
		state = &ServerState{Host: host}
	}

	var out []ActionResult

	if !state.Installed {
		if err := runStep(ctx, runner, host, user, port, &out,
			ActionInstallWG, "", aptInstallCmd,
			fmt.Sprintf("install WireGuard on %s", host)); err != nil {
			return out, err
		}
	}

	if !state.KeysExist {
		genCmd := `mkdir -p /etc/wireguard && ` +
			`wg genkey | tee /etc/wireguard/privatekey | wg pubkey | tee /etc/wireguard/publickey && ` +
			`chmod 600 /etc/wireguard/privatekey`
		if err := runStep(ctx, runner, host, user, port, &out,
			ActionGenKeyPair, "", genCmd,
			fmt.Sprintf("generate keypair on %s", host)); err != nil {
			return out, err
		}
	}

	if desired.InstallPodman {
		if !state.PodmanInstalled {
			if err := runStep(ctx, runner, host, user, port, &out,
				ActionInstallPodman, "", podmanInstallCmd,
				fmt.Sprintf("install Podman on %s", host)); err != nil {
				return out, err
			}
		}
		if !state.PodmanSocketActive {
			if err := runStep(ctx, runner, host, user, port, &out,
				ActionEnablePodmanSocket, "", enablePodmanSocketCmd,
				fmt.Sprintf("enable podman.socket on %s", host)); err != nil {
				return out, err
			}
		}
		if !state.IPForwardEnabled {
			if err := runStep(ctx, runner, host, user, port, &out,
				ActionEnableIPForward, "", enableIPForwardCmd,
				fmt.Sprintf("enable IP forwarding on %s", host)); err != nil {
				return out, err
			}
		}
	}

	return out, nil
}

// phase2Server writes the WireGuard config, enables/reloads the service,
// creates per-namespace Podman bridges, and installs the firewall service.
func phase2Server(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	port int,
	desired *DesiredMesh,
	fresh MeshState,
	mgmtAssignments map[string]net.IP,
	containerAssignments map[string]map[string]*net.IPNet,
) ([]ActionResult, error) {
	var out []ActionResult

	mgmtIP := mgmtAssignments[host]
	nsSorted := desired.SortedNamespaces()

	// Build peer list (everyone except self, skip hosts with no pubkey).
	// Each peer's AllowedIPs covers every namespace subnet that peer owns.
	var peers []PeerConfig
	for _, peer := range desired.Hosts {
		if peer == host {
			continue
		}
		ps, ok := fresh.Servers[peer]
		if !ok || ps.PublicKey == "" {
			continue
		}
		var subnets []*net.IPNet
		for _, ns := range nsSorted {
			if sn := containerAssignments[ns][peer]; sn != nil {
				subnets = append(subnets, sn)
			}
		}
		peers = append(peers, PeerConfig{
			Endpoint:         peer,
			PublicKey:        ps.PublicKey,
			MgmtIP:           mgmtAssignments[peer],
			ContainerSubnets: subnets,
		})
	}

	// Write WG config.
	configCmd := WriteConfigCommand(desired.Interface, mgmtIP, desired.ListenPort, peers)
	if err := runStep(ctx, runner, host, user, port, &out,
		ActionWriteConfig, "", configCmd,
		fmt.Sprintf("write config on %s", host)); err != nil {
		return out, err
	}

	// Enable or reload wg-quick.
	state := fresh.Servers[host]
	var serviceCmd string
	actionType := ActionEnableService
	if state != nil && state.Active {
		serviceCmd = fmt.Sprintf(`systemctl restart wg-quick@%s 2>&1 || wg syncconf %s <(wg-quick strip %s) 2>&1`,
			desired.Interface, desired.Interface, desired.Interface)
		actionType = ActionReloadService
	} else {
		serviceCmd = fmt.Sprintf(`systemctl enable --now wg-quick@%s 2>&1`, desired.Interface)
	}
	if err := runStep(ctx, runner, host, user, port, &out,
		actionType, "", serviceCmd,
		fmt.Sprintf("enable/reload service on %s", host)); err != nil {
		return out, err
	}

	if desired.InstallPodman {
		freshState := fresh.Servers[host]

		// Per-namespace podman network reconcile.
		for _, ns := range nsSorted {
			contSubnet := containerAssignments[ns][host]
			if contSubnet == nil {
				continue
			}
			netName := PodmanNetworkFor(ns)
			gw := MachineIP(contSubnet)

			var nss *NamespaceServerState
			if freshState != nil {
				nss = freshState.Namespaces[ns]
			}

			if nss == nil || !nss.NetworkExists {
				netCmd := podmanNetCreateCmd(netName, ns, contSubnet, gw)
				if err := runStep(ctx, runner, host, user, port, &out,
					ActionCreatePodmanNet, ns, netCmd,
					fmt.Sprintf("create Podman network %s on %s", netName, host)); err != nil {
					return out, err
				}
				continue
			}

			subnetDrift := nss.ContainerSubnet != nil && nss.ContainerSubnet.String() != contSubnet.String()
			if nss.DNSEnabled || subnetDrift || nss.Label != ns {
				recreateCmd := podmanNetRecreateCmd(netName, ns, contSubnet, gw)
				if err := runStep(ctx, runner, host, user, port, &out,
					ActionRecreatePodmanNet, ns, recreateCmd,
					fmt.Sprintf("recreate Podman network %s on %s", netName, host)); err != nil {
					return out, err
				}
			}
		}

		// Firewall service: union of namespace subnets; reinstall when missing,
		// default-deny flipped, or unit text drifted (e.g. namespace added).
		var subnets []*net.IPNet
		for _, ns := range nsSorted {
			if sn := containerAssignments[ns][host]; sn != nil {
				subnets = append(subnets, sn)
			}
		}
		expectedUnit := FirewallServiceUnit(desired.Interface, subnets, desired.DefaultDenyContainers)
		expectedUnitHash := sha256Hex([]byte(expectedUnit))
		unitDrift := freshState != nil && freshState.FirewallUnitSha256 != expectedUnitHash

		if freshState == nil || !freshState.FirewallActive ||
			freshState.DefaultDenyActive != desired.DefaultDenyContainers ||
			unitDrift {
			fwCmd := InstallFirewallCommand(desired.Interface, subnets, desired.DefaultDenyContainers)
			if err := runStep(ctx, runner, host, user, port, &out,
				ActionInstallFirewall, "", fwCmd,
				fmt.Sprintf("install firewall service on %s", host)); err != nil {
				return out, err
			}
		}
	}

	return out, nil
}

// Verify SSHes into each host and checks that WireGuard is active with the
// expected number of peers.
func Verify(
	ctx context.Context,
	runner ssh.Runner,
	hosts []string,
	user string,
	port int,
	iface string,
	concurrency int,
) []VerifyResult {
	results := ssh.ForEachServer(ctx, hosts, concurrency,
		func(ctx context.Context, host string) (VerifyResult, error) {
			return verifyHost(ctx, runner, host, user, port, iface, len(hosts)-1)
		})

	out := make([]VerifyResult, len(results))
	for i, r := range results {
		if r.Err != nil {
			out[i] = VerifyResult{Host: r.Host, Err: r.Err}
		} else {
			out[i] = r.Result
		}
	}
	return out
}

func verifyHost(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	port int,
	iface string,
	expectedPeers int,
) (VerifyResult, error) {
	result := VerifyResult{Host: host}

	stdout, _, err := runner.Run(ctx, host, user, port,
		fmt.Sprintf(`wg show %s dump 2>/dev/null || true`, iface))
	if err != nil {
		return result, fmt.Errorf("wg show on %s: %w", host, err)
	}

	lines := nonEmptyLines(stdout)
	if len(lines) == 0 {
		result.Err = fmt.Errorf("interface %s not active", iface)
		return result, nil
	}

	result.Active = true
	result.PeerCount = len(lines) - 1

	stdout2, _, _ := runner.Run(ctx, host, user, port,
		fmt.Sprintf(`grep '^Address' /etc/wireguard/%s.conf 2>/dev/null || true`, iface))
	if addr := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(stdout2), "Address =")); addr != "" {
		ip, _, _ := net.ParseCIDR(strings.TrimSpace(addr))
		result.WireGuardIP = ip
	}

	if result.PeerCount < expectedPeers {
		result.Err = fmt.Errorf("expected %d peer(s), got %d", expectedPeers, result.PeerCount)
	}

	return result, nil
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

// heredocWrite emits a shell command that atomically writes body to remotePath
// via a single-quoted heredoc.  Body is trusted (generated by us).
// chmod runs before mv so the final rename is atomic with the intended mode.
func heredocWrite(remotePath, body, tag string, mode os.FileMode) string {
	return fmt.Sprintf(`cat > %[1]s.tmp <<'%[3]s'
%[2]s%[3]s
chmod %[4]o %[1]s.tmp
mv %[1]s.tmp %[1]s`, remotePath, body, tag, mode)
}

// phase3Server downloads corrosion + coold from GitHub releases, writes their
// configs/unit files, and enables both services.
// Guarded by desired.InstallCoold at the caller.
func phase3Server(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	port int,
	desired *DesiredMesh,
	fresh MeshState,
	mgmtAssignments map[string]net.IP,
	containerAssignments map[string]map[string]*net.IPNet,
) ([]ActionResult, error) {
	var out []ActionResult

	mgmtIP := mgmtAssignments[host]
	if mgmtIP == nil {
		return out, fmt.Errorf("no mgmt IP allocated for %s", host)
	}
	nsSorted := desired.SortedNamespaces()
	nsConfigs := buildNamespaceConfigs(host, nsSorted, containerAssignments)
	if len(nsConfigs) == 0 {
		return out, fmt.Errorf("no namespace subnets allocated for %s", host)
	}

	freshState := fresh.Servers[host]

	// 1. Download + install corrosion if version drifted.
	if binaryVersionDrift(desired.CorrosionVersion,
		freshState != nil && freshState.CorrosionInstalled,
		func() string {
			if freshState != nil {
				return freshState.CorrosionVersion
			}
			return ""
		}()) {
		if err := runStep(ctx, runner, host, user, port, &out,
			ActionInstallCorrosion, "",
			services.CorrosionInstallCommand(desired.CorrosionVersion),
			fmt.Sprintf("install corrosion on %s", host)); err != nil {
			return out, err
		}
	}

	// 2. Download + install coold if version drifted.
	if binaryVersionDrift(desired.CooldVersion,
		freshState != nil && freshState.CooldInstalled,
		func() string {
			if freshState != nil {
				return freshState.CooldVersion
			}
			return ""
		}()) {
		if err := runStep(ctx, runner, host, user, port, &out,
			ActionInstallCoold, "",
			services.CooldInstallCommand(desired.CooldVersion),
			fmt.Sprintf("install coold on %s", host)); err != nil {
			return out, err
		}
	}

	// 3. Create dirs for corrosion state/config/admin socket.
	if err := runStep(ctx, runner, host, user, port, &out,
		ActionWriteCorrosionConfig, "",
		`mkdir -p /etc/corrosion/schemas /var/lib/corrosion /var/run/corrosion`,
		fmt.Sprintf("mkdir corrosion dirs on %s", host)); err != nil {
		return out, err
	}

	// 4. Write config.toml.
	peers := peerMgmtIPs(host, desired.Hosts, mgmtAssignments)
	configBody := string(services.CorrosionConfigBytes(mgmtIP,
		desired.CorrosionGossipPort, desired.CorrosionAPIPort, peers))
	configCmd := heredocWrite("/etc/corrosion/config.toml", configBody, "COOLIFY_CORROSION_EOF", 0o600)
	if err := runStep(ctx, runner, host, user, port, &out,
		ActionWriteCorrosionConfig, "", configCmd,
		fmt.Sprintf("write corrosion config on %s", host)); err != nil {
		return out, err
	}

	// 5. Write schema. When schema content drifts (not first install) the
	// CR-SQLite on-disk DB is incompatible — stop corrosion and wipe the DB
	// so it re-bootstraps from the new schema. Coold repopulates within ~2s.
	expectedSchemaSha := sha256Hex([]byte(services.CoolifySchemaSQL))
	schemaDrift := freshState != nil &&
		freshState.CorrosionSchemaSha256 != "" &&
		freshState.CorrosionSchemaSha256 != expectedSchemaSha
	schemaCmd := heredocWrite("/etc/corrosion/schemas/coolify.sql",
		services.CoolifySchemaSQL, "COOLIFY_SCHEMA_EOF", 0o600)
	if schemaDrift {
		schemaCmd = `systemctl stop corrosion 2>/dev/null || true; ` +
			`rm -f /var/lib/corrosion/corrosion.db ` +
			`/var/lib/corrosion/corrosion.db-shm ` +
			`/var/lib/corrosion/corrosion.db-wal && ` +
			schemaCmd
	}
	if err := runStep(ctx, runner, host, user, port, &out,
		ActionWriteCorrosionSchema, "", schemaCmd,
		fmt.Sprintf("write corrosion schema on %s", host)); err != nil {
		return out, err
	}

	// 6. Write corrosion unit + 7. Write coold unit + 8. daemon-reload + enable.
	// Use enable + restart (not enable --now) so an already-active service still
	// picks up new unit/config/schema without a separate reload step.
	//
	// Also ensure the coold API bearer token exists before the unit starts.
	// The command is idempotent — reruns keep the existing token so clients
	// don't get invalidated on every `apply`.
	corrosionUnit := services.CorrosionServiceUnit(desired.Interface)
	cooldUnit := services.CooldServiceUnit(mgmtIP, nsConfigs)

	serviceCmd := services.EnsureCooldAPITokenCommand() +
		" && " +
		heredocWrite("/etc/systemd/system/corrosion.service",
			corrosionUnit, "COOLIFY_CORROSION_UNIT_EOF", 0o644) +
		" && " +
		heredocWrite("/etc/systemd/system/coold.service",
			cooldUnit, "COOLIFY_COOLD_UNIT_EOF", 0o644) +
		` && systemctl daemon-reload` +
		` && systemctl enable corrosion coold` +
		` && systemctl restart corrosion` +
		` && sleep 1` +
		` && systemctl restart coold`

	if err := runStep(ctx, runner, host, user, port, &out,
		ActionInstallCorrosionService, "", serviceCmd,
		fmt.Sprintf("install corrosion+coold services on %s", host)); err != nil {
		return out, err
	}

	// Append a trailing coold install result so the rendered table matches
	// the planned action list (install-coold-service).
	out = append(out, ActionResult{
		Action: PlannedAction{
			Host:   host,
			Type:   ActionInstallCooldService,
			Detail: fmt.Sprintf("coold.service (mgmt=%s, namespaces=%d)", mgmtIP, len(nsConfigs)),
		},
	})

	return out, nil
}

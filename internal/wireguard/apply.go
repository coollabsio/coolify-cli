package wireguard

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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

// podmanNetCreateCmd creates a Podman bridge network bound to the host's
// container subnet. Idempotent: skips if the network already exists.
// The bridge gateway is MachineIP(subnet) (the .1 of the /24).
func podmanNetCreateCmd(name string, subnet *net.IPNet, gateway net.IP) string {
	return fmt.Sprintf(
		`podman network exists %s 2>/dev/null && echo "network exists, skipping" || `+
			`podman network create --driver bridge --subnet=%s --gateway=%s %s`,
		name, subnet, gateway, name)
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
	cmd string,
	errFmt string,
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
		Action: PlannedAction{Host: host, Type: atype, Detail: detail},
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
//     create Podman network, install firewall service.
func ApplyMesh(
	ctx context.Context,
	runner ssh.Runner,
	uploader ssh.FileUploader,
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
		desired.Interface, desired.PodmanNetworkName, concurrency)
	if err != nil {
		return results, fmt.Errorf("re-probe after phase 1: %w", err)
	}

	mgmtAssignments, _, err := AllocateMgmtIPs(desired.MgmtPool, fresh.AssignedMgmtIPs(), desired.Hosts)
	if err != nil {
		return results, fmt.Errorf("mgmt IP allocation: %w", err)
	}
	containerAssignments, _, err := Allocate(desired.ContainerPool, desired.ContainerPrefix,
		fresh.AssignedContainerSubnets(), desired.Hosts)
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
		// Compute binary sha256 once; reused across all hosts.
		corrosionSha, shaErr := fileSha256(desired.CorrosionBinaryPath)
		if shaErr != nil {
			return results, fmt.Errorf("hash corrosion binary: %w", shaErr)
		}
		cooldSha, shaErr := fileSha256(desired.CooldBinaryPath)
		if shaErr != nil {
			return results, fmt.Errorf("hash coold binary: %w", shaErr)
		}

		p3 := ssh.ForEachServer(ctx, desired.Hosts, concurrency,
			func(ctx context.Context, host string) ([]ActionResult, error) {
				return phase3Server(ctx, runner, uploader, host, user, port,
					desired, fresh, mgmtAssignments, corrosionSha, cooldSha)
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
			ActionInstallWG, aptInstallCmd,
			fmt.Sprintf("install WireGuard on %s", host)); err != nil {
			return out, err
		}
	}

	if !state.KeysExist {
		genCmd := `mkdir -p /etc/wireguard && ` +
			`wg genkey | tee /etc/wireguard/privatekey | wg pubkey | tee /etc/wireguard/publickey && ` +
			`chmod 600 /etc/wireguard/privatekey`
		if err := runStep(ctx, runner, host, user, port, &out,
			ActionGenKeyPair, genCmd,
			fmt.Sprintf("generate keypair on %s", host)); err != nil {
			return out, err
		}
	}

	if desired.InstallPodman {
		if !state.PodmanInstalled {
			if err := runStep(ctx, runner, host, user, port, &out,
				ActionInstallPodman, podmanInstallCmd,
				fmt.Sprintf("install Podman on %s", host)); err != nil {
				return out, err
			}
		}
		if !state.PodmanSocketActive {
			if err := runStep(ctx, runner, host, user, port, &out,
				ActionEnablePodmanSocket, enablePodmanSocketCmd,
				fmt.Sprintf("enable podman.socket on %s", host)); err != nil {
				return out, err
			}
		}
		if !state.IPForwardEnabled {
			if err := runStep(ctx, runner, host, user, port, &out,
				ActionEnableIPForward, enableIPForwardCmd,
				fmt.Sprintf("enable IP forwarding on %s", host)); err != nil {
				return out, err
			}
		}
	}

	return out, nil
}

// phase2Server writes the WireGuard config, enables/reloads the service,
// creates the per-host Podman bridge, and installs the firewall service.
func phase2Server(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	port int,
	desired *DesiredMesh,
	fresh MeshState,
	mgmtAssignments map[string]net.IP,
	containerAssignments map[string]*net.IPNet,
) ([]ActionResult, error) {
	var out []ActionResult

	mgmtIP := mgmtAssignments[host]
	contSubnet := containerAssignments[host]

	// Build peer list (everyone except self, skip hosts with no pubkey).
	var peers []PeerConfig
	for _, peer := range desired.Hosts {
		if peer == host {
			continue
		}
		ps, ok := fresh.Servers[peer]
		if !ok || ps.PublicKey == "" {
			continue
		}
		peers = append(peers, PeerConfig{
			Endpoint:        peer,
			PublicKey:       ps.PublicKey,
			MgmtIP:          mgmtAssignments[peer],
			ContainerSubnet: containerAssignments[peer],
		})
	}

	// Write WG config.
	configCmd := WriteConfigCommand(desired.Interface, mgmtIP, desired.ListenPort, peers)
	if err := runStep(ctx, runner, host, user, port, &out,
		ActionWriteConfig, configCmd,
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
		actionType, serviceCmd,
		fmt.Sprintf("enable/reload service on %s", host)); err != nil {
		return out, err
	}

	if desired.InstallPodman {
		freshState := fresh.Servers[host]

		// Create Podman bridge network if not already present.
		if freshState == nil || !freshState.PodmanNetExists {
			netCmd := podmanNetCreateCmd(desired.PodmanNetworkName, contSubnet, MachineIP(contSubnet))
			if err := runStep(ctx, runner, host, user, port, &out,
				ActionCreatePodmanNet, netCmd,
				fmt.Sprintf("create Podman network on %s", host)); err != nil {
				return out, err
			}
		}

		// Install firewall service if not active or default-deny mode drifted.
		if freshState == nil || !freshState.FirewallActive ||
			freshState.DefaultDenyActive != desired.DefaultDenyContainers {
			fwCmd := InstallFirewallCommand(desired.Interface, contSubnet, desired.DefaultDenyContainers)
			if err := runStep(ctx, runner, host, user, port, &out,
				ActionInstallFirewall, fwCmd,
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

// FileSha256 hashes the file at path (hex). Exported so cmd/init can prefill
// DesiredMesh binary hashes, enabling plan-layer drift detection.
func FileSha256(path string) (string, error) { return fileSha256(path) }

// fileSha256 hashes the file at path.  Used to skip unnecessary binary uploads.
func fileSha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// remoteSha256 reads sha256sum for a remote path; returns empty string on miss.
func remoteSha256(ctx context.Context, runner ssh.Runner, host, user string, port int, path string) string {
	stdout, _, _ := runner.Run(ctx, host, user, port,
		fmt.Sprintf(`sha256sum %s 2>/dev/null | awk '{print $1}' || true`, path))
	return strings.TrimSpace(stdout)
}

// uploadStep uploads localPath to remotePath and appends an ActionResult.
// It is the FileUploader analogue of runStep.
func uploadStep(
	ctx context.Context,
	uploader ssh.FileUploader,
	host, user string,
	port int,
	out *[]ActionResult,
	atype ActionType,
	localPath, remotePath string,
	mode os.FileMode,
	errFmt string,
) error {
	err := uploader.UploadFile(ctx, host, user, port, localPath, remotePath, mode)
	detail := remotePath
	if err != nil {
		detail = err.Error()
	}
	*out = append(*out, ActionResult{
		Action: PlannedAction{Host: host, Type: atype, Detail: detail},
		Err:    err,
	})
	if err != nil {
		return fmt.Errorf(errFmt+": %w", err)
	}
	return nil
}

// heredocWrite emits a shell command that atomically writes body to remotePath
// via a single-quoted heredoc.  Body is trusted (generated by us).
func heredocWrite(remotePath, body, tag string) string {
	return fmt.Sprintf(`cat > %[1]s.tmp <<'%[3]s'
%[2]s%[3]s
mv %[1]s.tmp %[1]s`, remotePath, body, tag)
}

// phase3Server uploads corrosion + coold, writes their configs/unit files,
// and enables both services.  Guarded by desired.InstallCoold at the caller.
func phase3Server(
	ctx context.Context,
	runner ssh.Runner,
	uploader ssh.FileUploader,
	host, user string,
	port int,
	desired *DesiredMesh,
	fresh MeshState,
	mgmtAssignments map[string]net.IP,
	corrosionSha, cooldSha string,
) ([]ActionResult, error) {
	var out []ActionResult

	mgmtIP := mgmtAssignments[host]
	if mgmtIP == nil {
		return out, fmt.Errorf("no mgmt IP allocated for %s", host)
	}

	// 1. Upload corrosion binary (skip if sha matches).
	if remoteSha256(ctx, runner, host, user, port, "/usr/local/bin/corrosion") != corrosionSha {
		if err := uploadStep(ctx, uploader, host, user, port, &out,
			ActionUploadCorrosion,
			desired.CorrosionBinaryPath, "/usr/local/bin/corrosion",
			0o755,
			fmt.Sprintf("upload corrosion to %s", host)); err != nil {
			return out, err
		}
	}

	// 2. Upload coold binary.
	if remoteSha256(ctx, runner, host, user, port, "/usr/local/bin/coold") != cooldSha {
		if err := uploadStep(ctx, uploader, host, user, port, &out,
			ActionUploadCoold,
			desired.CooldBinaryPath, "/usr/local/bin/coold",
			0o755,
			fmt.Sprintf("upload coold to %s", host)); err != nil {
			return out, err
		}
	}

	// 3. Create dirs for corrosion state/config/admin socket.
	if err := runStep(ctx, runner, host, user, port, &out,
		ActionWriteCorrosionConfig,
		`mkdir -p /etc/corrosion/schemas /var/lib/corrosion /var/run/corrosion`,
		fmt.Sprintf("mkdir corrosion dirs on %s", host)); err != nil {
		return out, err
	}

	// 4. Write config.toml.
	peers := peerMgmtIPs(host, desired.Hosts, mgmtAssignments)
	configBody := string(services.CorrosionConfigBytes(mgmtIP,
		desired.CorrosionGossipPort, desired.CorrosionAPIPort, peers))
	configCmd := heredocWrite("/etc/corrosion/config.toml", configBody, "COOLIFY_CORROSION_EOF")
	if err := runStep(ctx, runner, host, user, port, &out,
		ActionWriteCorrosionConfig, configCmd,
		fmt.Sprintf("write corrosion config on %s", host)); err != nil {
		return out, err
	}

	// 5. Write schema.
	schemaCmd := heredocWrite("/etc/corrosion/schemas/coolify.sql",
		services.CoolifySchemaSQL, "COOLIFY_SCHEMA_EOF")
	if err := runStep(ctx, runner, host, user, port, &out,
		ActionWriteCorrosionSchema, schemaCmd,
		fmt.Sprintf("write corrosion schema on %s", host)); err != nil {
		return out, err
	}

	// 6. Write corrosion unit + 7. Write coold unit + 8. daemon-reload + enable.
	corrosionUnit := services.CorrosionServiceUnit(desired.Interface)
	cooldUnit := services.CooldServiceUnit(mgmtIP)

	serviceCmd := heredocWrite("/etc/systemd/system/corrosion.service",
		corrosionUnit, "COOLIFY_CORROSION_UNIT_EOF") +
		" && " +
		heredocWrite("/etc/systemd/system/coold.service",
			cooldUnit, "COOLIFY_COOLD_UNIT_EOF") +
		` && systemctl daemon-reload` +
		` && systemctl enable --now corrosion` +
		` && sleep 1` +
		` && systemctl enable --now coold`

	if err := runStep(ctx, runner, host, user, port, &out,
		ActionInstallCorrosionService, serviceCmd,
		fmt.Sprintf("install corrosion+coold services on %s", host)); err != nil {
		return out, err
	}

	// Append a trailing coold install result so the rendered table matches
	// the planned action list (install-coold-service).
	out = append(out, ActionResult{
		Action: PlannedAction{
			Host:   host,
			Type:   ActionInstallCooldService,
			Detail: fmt.Sprintf("coold.service (mgmt=%s)", mgmtIP),
		},
	})

	return out, nil
}

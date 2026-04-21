package wireguard

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/services"
)

// ActionType identifies the kind of change required.
type ActionType string

const (
	ActionInstallWG               ActionType = "install-wg"
	ActionGenKeyPair              ActionType = "gen-keypair"
	ActionAllocateMgmtIP          ActionType = "allocate-mgmt-ip"
	ActionAllocateContainerSubnet ActionType = "allocate-container-subnet"
	ActionWriteConfig             ActionType = "write-config"
	ActionEnableService           ActionType = "enable-service"
	ActionReloadService           ActionType = "reload-service"
	ActionAddPeer                 ActionType = "add-peer"
	ActionRemovePeer              ActionType = "remove-peer"
	ActionInstallPodman           ActionType = "install-podman"
	ActionEnablePodmanSocket      ActionType = "enable-podman-socket"
	ActionEnableIPForward         ActionType = "enable-ip-forward"
	ActionCreatePodmanNet         ActionType = "create-podman-network"
	ActionRecreatePodmanNet       ActionType = "recreate-podman-network"
	ActionInstallFirewall         ActionType = "install-firewall"
	ActionInstallCorrosion        ActionType = "install-corrosion"
	ActionInstallCoold            ActionType = "install-coold"
	ActionWriteCorrosionConfig    ActionType = "write-corrosion-config"
	ActionWriteCorrosionSchema    ActionType = "write-corrosion-schema"
	ActionInstallCorrosionService ActionType = "install-corrosion-service"
	ActionInstallCooldService     ActionType = "install-coold-service"
	ActionInstallRedis            ActionType = "install-redis"
	ActionInstallBroker           ActionType = "install-broker"
	ActionGenerateJWTKeypair      ActionType = "generate-jwt-keypair"
	ActionInstallBrokerService    ActionType = "install-broker-service"
	ActionWriteHostJWT            ActionType = "write-host-jwt"
	ActionUpdateCooldBrokerEnv    ActionType = "update-coold-broker-env"
)

// PlannedAction is one step that apply must execute on a host.
type PlannedAction struct {
	Host      string
	Namespace string // empty for host-global actions
	Type      ActionType
	Detail    string
}

// Plan is the list of actions needed to converge the mesh to the desired state.
type Plan struct {
	Actions []PlannedAction
	// MgmtAssignments maps host → planned WG management /32 IP.
	MgmtAssignments map[string]net.IP
	// SubnetAssignments maps namespace → host → planned container subnet.
	SubnetAssignments map[string]map[string]*net.IPNet
	// Warnings contains non-fatal conflict messages from the IP allocator.
	Warnings []Warning
}

// IsEmpty returns true when the mesh is already converged (no changes needed).
func (p *Plan) IsEmpty() bool { return len(p.Actions) == 0 }

// BuildPlan computes the actions required to bring current into alignment
// with desired.  It is a pure function: no SSH, no I/O.
func BuildPlan(desired *DesiredMesh, current MeshState) (*Plan, error) {
	if desired.DefaultDenyContainers && !desired.InstallPodman {
		return nil, fmt.Errorf("--default-deny requires --podman")
	}
	if desired.InstallCoold && !desired.InstallPodman {
		return nil, fmt.Errorf("--install-coold requires --podman")
	}
	if desired.InstallPodman && len(desired.Namespaces) == 0 {
		return nil, fmt.Errorf("at least one namespace is required")
	}

	// Validate per-host preconditions before computing actions.
	for _, host := range desired.Hosts {
		if state, ok := current.Servers[host]; ok && desired.DefaultDenyContainers {
			if !state.NftAvailable {
				return nil, fmt.Errorf(
					"host %s: nft binary not available; install nftables or pass --skip-default-deny",
					host,
				)
			}
		}
	}

	mgmtAssignments, mgmtWarns, err := AllocateMgmtIPs(desired.MgmtPool, current.AssignedMgmtIPs(), desired.Hosts)
	if err != nil {
		return nil, fmt.Errorf("mgmt IP allocation: %w", err)
	}

	containerAssignments, contWarns, err := AllocateNamespaced(
		desired.ContainerPool, desired.ContainerPrefix,
		current.AssignedContainerSubnets(), desired.Namespaces, desired.Hosts)
	if err != nil {
		return nil, fmt.Errorf("container subnet allocation: %w", err)
	}

	plan := &Plan{
		MgmtAssignments:   mgmtAssignments,
		SubnetAssignments: containerAssignments,
		Warnings:          append(mgmtWarns, contWarns...),
	}

	nsSorted := desired.SortedNamespaces()

	for _, host := range desired.Hosts {
		state, ok := current.Servers[host]
		if !ok {
			state = &ServerState{Host: host, Namespaces: map[string]*NamespaceServerState{}}
		}
		if state.Namespaces == nil {
			state.Namespaces = map[string]*NamespaceServerState{}
		}

		// --- WireGuard installation ---
		if !state.Installed {
			plan.Actions = append(plan.Actions, PlannedAction{
				Host:   host,
				Type:   ActionInstallWG,
				Detail: "wireguard not installed",
			})
		}

		// --- Key generation ---
		if !state.KeysExist {
			plan.Actions = append(plan.Actions, PlannedAction{
				Host:   host,
				Type:   ActionGenKeyPair,
				Detail: "no keys at /etc/wireguard/privatekey",
			})
		}

		// --- Mgmt IP allocation ---
		mgmtIP := mgmtAssignments[host]
		if state.WireGuardMgmtIP == nil ||
			!state.WireGuardMgmtIP.Equal(mgmtIP) {
			plan.Actions = append(plan.Actions, PlannedAction{
				Host:   host,
				Type:   ActionAllocateMgmtIP,
				Detail: fmt.Sprintf("%s/32", mgmtIP),
			})
		}

		// --- Container subnet allocation (one per namespace) ---
		if desired.InstallPodman {
			for _, ns := range nsSorted {
				contSubnet := containerAssignments[ns][host]
				current := state.Namespaces[ns]
				if current == nil || current.ContainerSubnet == nil ||
					current.ContainerSubnet.String() != contSubnet.String() {
					plan.Actions = append(plan.Actions, PlannedAction{
						Host:      host,
						Namespace: ns,
						Type:      ActionAllocateContainerSubnet,
						Detail:    contSubnet.String(),
					})
				}
			}
		}

		// --- Peer diff ---
		desiredPeerKeys := make(map[string]bool)
		for _, peer := range desired.Hosts {
			if peer == host {
				continue
			}
			if ps, ok2 := current.Servers[peer]; ok2 && ps.PublicKey != "" {
				desiredPeerKeys[ps.PublicKey] = true
			}
		}

		currentPeerKeys := make(map[string]bool)
		for _, p := range state.Peers {
			currentPeerKeys[p.PublicKey] = true
		}

		for key := range desiredPeerKeys {
			if !currentPeerKeys[key] {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionAddPeer,
					Detail: truncateKey(key),
				})
			}
		}
		for key := range currentPeerKeys {
			if !desiredPeerKeys[key] {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionRemovePeer,
					Detail: truncateKey(key),
				})
			}
		}

		// --- Config write ---
		mgmtMismatch := state.WireGuardMgmtIP == nil || !state.WireGuardMgmtIP.Equal(mgmtIP)
		allowedIPsDrift := allowedIPsNeedsRewrite(host, desired, current, containerAssignments, mgmtAssignments, state)
		needsConfig := mgmtMismatch ||
			allowedIPsDrift ||
			len(plan.actionsForHost(host, ActionAddPeer)) > 0 ||
			len(plan.actionsForHost(host, ActionRemovePeer)) > 0 ||
			!state.KeysExist ||
			!state.Installed ||
			len(desired.Hosts) > 1 && state.ListenPort != desired.ListenPort

		if needsConfig {
			plan.Actions = append(plan.Actions, PlannedAction{
				Host:   host,
				Type:   ActionWriteConfig,
				Detail: fmt.Sprintf("%s.conf (%d peer(s))", desired.Interface, len(desired.Hosts)-1),
			})
		}

		// --- WG service ---
		if !state.Active {
			plan.Actions = append(plan.Actions, PlannedAction{
				Host:   host,
				Type:   ActionEnableService,
				Detail: fmt.Sprintf("systemctl enable --now wg-quick@%s", desired.Interface),
			})
		} else if needsConfig {
			plan.Actions = append(plan.Actions, PlannedAction{
				Host:   host,
				Type:   ActionReloadService,
				Detail: fmt.Sprintf("systemctl reload wg-quick@%s (config changed)", desired.Interface),
			})
		}

		// --- Podman stack ---
		if desired.InstallPodman {
			if !state.PodmanInstalled {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionInstallPodman,
					Detail: "podman not installed",
				})
			}
			if !state.PodmanSocketActive {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionEnablePodmanSocket,
					Detail: "systemctl enable --now podman.socket",
				})
			}
			if !state.IPForwardEnabled {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionEnableIPForward,
					Detail: "net.ipv4.ip_forward=1",
				})
			}

			for _, ns := range nsSorted {
				contSubnet := containerAssignments[ns][host]
				netName := PodmanNetworkFor(ns)
				nss := state.Namespaces[ns]
				gw := MachineIP(contSubnet)

				if nss == nil || !nss.NetworkExists {
					plan.Actions = append(plan.Actions, PlannedAction{
						Host:      host,
						Namespace: ns,
						Type:      ActionCreatePodmanNet,
						Detail:    fmt.Sprintf("%s subnet=%s gateway=%s", netName, contSubnet, gw),
					})
					continue
				}
				if nss.DNSEnabled ||
					(nss.ContainerSubnet != nil && nss.ContainerSubnet.String() != contSubnet.String()) ||
					nss.Label != ns {
					reasons := []string{}
					if nss.DNSEnabled {
						reasons = append(reasons, "dns_enabled=true")
					}
					if nss.ContainerSubnet != nil && nss.ContainerSubnet.String() != contSubnet.String() {
						reasons = append(reasons, fmt.Sprintf("subnet drift (have %s, want %s)", nss.ContainerSubnet, contSubnet))
					}
					if nss.Label != ns {
						reasons = append(reasons, fmt.Sprintf("label=%q mismatch", nss.Label))
					}
					plan.Actions = append(plan.Actions, PlannedAction{
						Host:      host,
						Namespace: ns,
						Type:      ActionRecreatePodmanNet,
						Detail:    fmt.Sprintf("%s — %s", netName, strings.Join(reasons, "; ")),
					})
				}
			}

			// Expected firewall unit text — hash it and compare against the
			// remote unit so adding/removing a namespace reinstalls the unit.
			var subnets []*net.IPNet
			for _, ns := range nsSorted {
				subnets = append(subnets, containerAssignments[ns][host])
			}
			expectedUnit := FirewallServiceUnit(desired.Interface, desired.SortedNamespaces(), subnets, desired.DefaultDenyContainers)
			expectedUnitHash := sha256Hex([]byte(expectedUnit))
			unitDrift := state.FirewallUnitSha256 != expectedUnitHash

			if !state.FirewallActive ||
				state.DefaultDenyActive != desired.DefaultDenyContainers ||
				unitDrift {
				detail := fmt.Sprintf("coolify-mesh-fw.service (%s, %d namespace(s), default-deny=%v)",
					desired.Interface, len(subnets), desired.DefaultDenyContainers)
				if unitDrift && state.FirewallUnitSha256 != "" {
					detail += " [unit drift]"
				}
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionInstallFirewall,
					Detail: detail,
				})
			}
		}

		// --- Corrosion + coold stack (v5 control plane) ---
		if desired.InstallCoold {
			corrosionDrift := binaryVersionDrift(desired.CorrosionVersion, state.CorrosionInstalled, state.CorrosionVersion)
			cooldDrift := binaryVersionDrift(desired.CooldVersion, state.CooldInstalled, state.CooldVersion)

			if corrosionDrift {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionInstallCorrosion,
					Detail: fmt.Sprintf("corrosion %s → /usr/local/bin/corrosion", desired.CorrosionVersion),
				})
			}
			if cooldDrift {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionInstallCoold,
					Detail: fmt.Sprintf("coold %s → /usr/local/bin/coold", desired.CooldVersion),
				})
			}

			peers := peerMgmtIPs(host, desired.Hosts, mgmtAssignments)
			expectedConfig := services.CorrosionConfigBytes(mgmtIP,
				desired.CorrosionGossipPort, desired.CorrosionAPIPort, peers)
			expectedHash := sha256Hex(expectedConfig)
			configDrift := state.CorrosionConfigHash != expectedHash

			if configDrift {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionWriteCorrosionConfig,
					Detail: fmt.Sprintf("/etc/corrosion/config.toml (peers=%d)", len(peers)),
				})
			}
			expectedSchemaSha := sha256Hex([]byte(services.CoolifySchemaSQL))
			schemaDrift := state.CorrosionSchemaSha256 != expectedSchemaSha
			if !state.CorrosionSchemaExists || schemaDrift {
				detail := "/etc/corrosion/schemas/coolify.sql"
				if schemaDrift && state.CorrosionSchemaSha256 != "" {
					detail += " [schema drift — DB will be reset]"
				}
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionWriteCorrosionSchema,
					Detail: detail,
				})
			}

			nsConfigs := buildNamespaceConfigs(host, nsSorted, containerAssignments)
			expectedCooldUnit := services.CooldServiceUnit(mgmtIP, nsConfigs)
			cooldUnitDrift := state.CooldUnitSha256 != sha256Hex([]byte(expectedCooldUnit))

			if !state.CorrosionActive || configDrift || corrosionDrift || schemaDrift {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionInstallCorrosionService,
					Detail: "systemctl enable --now corrosion",
				})
			}
			if !state.CooldActive || configDrift || cooldDrift || cooldUnitDrift {
				detail := fmt.Sprintf("systemctl enable --now coold (mgmt=%s, namespaces=%d)", mgmtIP, len(nsConfigs))
				if cooldUnitDrift && state.CooldUnitSha256 != "" {
					detail += " [unit drift]"
				}
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionInstallCooldService,
					Detail: detail,
				})
			}
		}
	}

	return plan, nil
}

// buildNamespaceConfigs builds the per-namespace CooldNamespace slice for this
// host, in namespace name order. Gateway IP for each namespace is the .1 of
// that namespace's per-host container subnet.
func buildNamespaceConfigs(host string, nsSorted []string, assignments map[string]map[string]*net.IPNet) []services.CooldNamespace {
	out := make([]services.CooldNamespace, 0, len(nsSorted))
	for _, ns := range nsSorted {
		subnet := assignments[ns][host]
		if subnet == nil {
			continue
		}
		out = append(out, services.CooldNamespace{
			Name:       ns,
			Network:    PodmanNetworkFor(ns),
			BridgeGateway: MachineIP(subnet),
		})
	}
	return out
}

// binaryVersionDrift returns true when a binary needs (re-)installation.
// Rules:
//   - not installed → always drift
//   - marker absent (empty haveVersion) → treat as drift (first-migration case)
//   - "nightly" tag → always re-install (moving target)
//   - pinned tag → drift only when marker differs from desired
func binaryVersionDrift(desiredVersion string, installed bool, haveVersion string) bool {
	if !installed || haveVersion == "" {
		return true
	}
	if desiredVersion == "nightly" {
		return true
	}
	return haveVersion != desiredVersion
}

// allowedIPsNeedsRewrite returns true when any [Peer] block on host does not
// have the expected AllowedIPs (peer mgmt /32 + every namespace subnet).
func allowedIPsNeedsRewrite(
	host string,
	desired *DesiredMesh,
	current MeshState,
	containerAssignments map[string]map[string]*net.IPNet,
	mgmtAssignments map[string]net.IP,
	state *ServerState,
) bool {
	if state == nil {
		return false
	}
	nsSorted := desired.SortedNamespaces()

	// Build pub-key → expected AllowedIPs set for every peer we should have.
	want := map[string]map[string]struct{}{}
	for _, peer := range desired.Hosts {
		if peer == host {
			continue
		}
		ps, ok := current.Servers[peer]
		if !ok || ps.PublicKey == "" {
			continue
		}
		mgmtIP := mgmtAssignments[peer]
		if mgmtIP == nil {
			continue
		}
		entries := map[string]struct{}{fmt.Sprintf("%s/32", mgmtIP): {}}
		for _, ns := range nsSorted {
			if sn := containerAssignments[ns][peer]; sn != nil {
				entries[sn.String()] = struct{}{}
			}
		}
		want[ps.PublicKey] = entries
	}

	// Compare against parsed peers in the current config. If any desired peer
	// has different AllowedIPs (missing or extra), we need to rewrite.
	have := map[string]map[string]struct{}{}
	for _, p := range state.Peers {
		s := map[string]struct{}{}
		for _, a := range p.AllowedIPs {
			s[strings.TrimSpace(a)] = struct{}{}
		}
		have[p.PublicKey] = s
	}
	for pk, wantSet := range want {
		haveSet, ok := have[pk]
		if !ok {
			return true
		}
		if !sameStringSet(wantSet, haveSet) {
			return true
		}
	}
	return false
}

func sameStringSet(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

// peerMgmtIPs returns the mgmt IPs of all hosts except self, drawn from the
// planned assignments so the result is stable even before any host has been
// probed.
func peerMgmtIPs(self string, hosts []string, assignments map[string]net.IP) []net.IP {
	out := make([]net.IP, 0, len(hosts)-1)
	for _, h := range hosts {
		if h == self {
			continue
		}
		if ip, ok := assignments[h]; ok && ip != nil {
			out = append(out, ip)
		}
	}
	return out
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// actionsForHost returns the subset of plan.Actions matching host and atype.
func (p *Plan) actionsForHost(host string, atype ActionType) []PlannedAction {
	var out []PlannedAction
	for _, a := range p.Actions {
		if a.Host == host && a.Type == atype {
			out = append(out, a)
		}
	}
	return out
}

// truncateKey shortens a base64 key to the first 8 chars + "…" for display.
func truncateKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:8] + "..."
}

// nsSortedCopy returns a fresh sorted copy (package-private helper exposed
// for tests and for callers that can't reach into DesiredMesh directly).
func nsSortedCopy(ns []string) []string {
	out := append([]string(nil), ns...)
	sort.Strings(out)
	return out
}

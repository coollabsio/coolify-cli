package wireguard

import (
	"fmt"
	"net"
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
	ActionInstallFirewall         ActionType = "install-firewall"
)

// PlannedAction is one step that apply must execute on a host.
type PlannedAction struct {
	Host   string
	Type   ActionType
	Detail string
}

// Plan is the list of actions needed to converge the mesh to the desired state.
type Plan struct {
	Actions []PlannedAction
	// MgmtAssignments maps host → planned WG management /32 IP.
	MgmtAssignments map[string]net.IP
	// SubnetAssignments maps host → planned container /24 subnet.
	SubnetAssignments map[string]*net.IPNet
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

	mgmtAssignments, mgmtWarns, err := AllocateMgmtIPs(desired.MgmtPool, current.AssignedMgmtIPs(), desired.Hosts)
	if err != nil {
		return nil, fmt.Errorf("mgmt IP allocation: %w", err)
	}

	containerAssignments, contWarns, err := Allocate(desired.ContainerPool, desired.ContainerPrefix,
		current.AssignedContainerSubnets(), desired.Hosts)
	if err != nil {
		return nil, fmt.Errorf("container subnet allocation: %w", err)
	}

	plan := &Plan{
		MgmtAssignments:   mgmtAssignments,
		SubnetAssignments: containerAssignments,
		Warnings:          append(mgmtWarns, contWarns...),
	}

	for _, host := range desired.Hosts {
		state, ok := current.Servers[host]
		if !ok {
			state = &ServerState{Host: host}
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

		// --- Container subnet allocation (only when --podman is set) ---
		contSubnet := containerAssignments[host]
		if desired.InstallPodman {
			if state.ContainerSubnet == nil ||
				state.ContainerSubnet.String() != contSubnet.String() {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionAllocateContainerSubnet,
					Detail: contSubnet.String(),
				})
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
		needsConfig := mgmtMismatch ||
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
			if !state.PodmanNetExists {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionCreatePodmanNet,
					Detail: fmt.Sprintf("%s subnet=%s gateway=%s", desired.PodmanNetworkName, contSubnet, MachineIP(contSubnet)),
				})
			}
			if !state.FirewallActive || state.DefaultDenyActive != desired.DefaultDenyContainers {
				plan.Actions = append(plan.Actions, PlannedAction{
					Host:   host,
					Type:   ActionInstallFirewall,
					Detail: fmt.Sprintf("coolify-mesh-fw.service (%s ↔ %s, default-deny=%v)", desired.Interface, contSubnet, desired.DefaultDenyContainers),
				})
			}
		}
	}

	return plan, nil
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

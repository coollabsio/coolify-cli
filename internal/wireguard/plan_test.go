package wireguard

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaultMgmtPool      = mustParseCIDR("100.64.0.0/16")
	defaultContainerPool = mustParseCIDR("10.210.0.0/16")
)

func desiredTwoHosts() *DesiredMesh {
	return &DesiredMesh{
		Hosts:           []string{"1.1.1.1", "2.2.2.2"},
		Interface:       "wg0",
		MgmtPool:        defaultMgmtPool,
		ContainerPool:   defaultContainerPool,
		ContainerPrefix: 24,
		ListenPort:      51820,
	}
}

func desiredWithPodman() *DesiredMesh {
	d := desiredTwoHosts()
	d.InstallPodman = true
	d.Namespaces = []string{DefaultNamespace}
	return d
}

// convergedServer returns a ServerState fully reconciled for the single
// `default` namespace with the supplied subnet.
func convergedServer(host, pubkey, peerKey, mgmtIP, contSubnet string) *ServerState {
	sn := mustParseCIDR(contSubnet)
	firewallHash := sha256Hex([]byte(FirewallServiceUnit("wg0", []string{"default"}, []*net.IPNet{sn}, false)))
	return &ServerState{
		Host:               host,
		Installed:          true,
		KeysExist:          true,
		PublicKey:          pubkey,
		WireGuardMgmtIP:    net.ParseIP(mgmtIP).To4(),
		ListenPort:         51820,
		Active:             true,
		Peers: []Peer{{
			PublicKey:  peerKey,
			AllowedIPs: []string{peerMgmtForPub(peerKey), peerSubnetForPub(peerKey)},
		}},
		PodmanInstalled:    true,
		PodmanSocketActive: true,
		IPForwardEnabled:   true,
		FirewallActive:     true,
		FirewallUnitSha256: firewallHash,
		Namespaces: map[string]*NamespaceServerState{
			DefaultNamespace: {
				Namespace:       DefaultNamespace,
				NetworkExists:   true,
				ContainerSubnet: sn,
				DNSEnabled:      false,
				Label:           DefaultNamespace,
			},
		},
	}
}

// peerMgmtForPub / peerSubnetForPub map the well-known test public keys to
// the mgmt /32 and /24 each peer is expected to own in the two-host fixture.
func peerMgmtForPub(pub string) string {
	switch pub {
	case "AAAAAAAA=":
		return "100.64.0.1/32"
	case "BBBBBBBB=":
		return "100.64.0.2/32"
	}
	return ""
}
func peerSubnetForPub(pub string) string {
	switch pub {
	case "AAAAAAAA=":
		return "10.210.0.0/24"
	case "BBBBBBBB=":
		return "10.210.1.0/24"
	}
	return ""
}

func TestBuildPlan_AlreadyConverged_NoPodman(t *testing.T) {
	desired := desiredTwoHosts()
	current := MeshState{
		Servers: map[string]*ServerState{
			"1.1.1.1": {
				Host:            "1.1.1.1",
				Installed:       true,
				KeysExist:       true,
				PublicKey:       "AAAAAAAA=",
				WireGuardMgmtIP: net.ParseIP("100.64.0.1").To4(),
				ListenPort:      51820,
				Active:          true,
				Peers:           []Peer{{PublicKey: "BBBBBBBB=", AllowedIPs: []string{"100.64.0.2/32"}}},
			},
			"2.2.2.2": {
				Host:            "2.2.2.2",
				Installed:       true,
				KeysExist:       true,
				PublicKey:       "BBBBBBBB=",
				WireGuardMgmtIP: net.ParseIP("100.64.0.2").To4(),
				ListenPort:      51820,
				Active:          true,
				Peers:           []Peer{{PublicKey: "AAAAAAAA=", AllowedIPs: []string{"100.64.0.1/32"}}},
			},
		},
	}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)
	assert.True(t, plan.IsEmpty(), "expected empty plan, got: %+v", plan.Actions)
}

func TestBuildPlan_FreshBootstrap(t *testing.T) {
	desired := desiredTwoHosts()
	current := MeshState{Servers: map[string]*ServerState{}}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)

	assert.False(t, plan.IsEmpty())

	actionTypes := func(host string) []ActionType {
		var out []ActionType
		for _, a := range plan.Actions {
			if a.Host == host {
				out = append(out, a.Type)
			}
		}
		return out
	}

	for _, host := range []string{"1.1.1.1", "2.2.2.2"} {
		types := actionTypes(host)
		assert.Contains(t, types, ActionInstallWG, host)
		assert.Contains(t, types, ActionGenKeyPair, host)
		assert.Contains(t, types, ActionAllocateMgmtIP, host)
		assert.Contains(t, types, ActionWriteConfig, host)
		assert.Contains(t, types, ActionEnableService, host)
	}
}

func TestBuildPlan_MgmtIPMismatchTriggersRewrite(t *testing.T) {
	desired := desiredTwoHosts()
	current := MeshState{
		Servers: map[string]*ServerState{
			"1.1.1.1": {
				Host:            "1.1.1.1",
				Installed:       true,
				KeysExist:       true,
				PublicKey:       "AAAAAAAA=",
				WireGuardMgmtIP: net.ParseIP("10.210.0.1").To4(), // outside 100.64/16
				ListenPort:      51820,
				Active:          true,
				Peers:           []Peer{{PublicKey: "BBBBBBBB="}},
			},
			"2.2.2.2": {
				Host:            "2.2.2.2",
				Installed:       true,
				KeysExist:       true,
				PublicKey:       "BBBBBBBB=",
				WireGuardMgmtIP: net.ParseIP("100.64.0.2").To4(),
				ListenPort:      51820,
				Active:          true,
				Peers:           []Peer{{PublicKey: "AAAAAAAA="}},
			},
		},
	}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)
	assert.NotEmpty(t, plan.Warnings)

	var aTypes []ActionType
	for _, a := range plan.Actions {
		if a.Host == "1.1.1.1" {
			aTypes = append(aTypes, a.Type)
		}
	}
	assert.Contains(t, aTypes, ActionAllocateMgmtIP)
	assert.Contains(t, aTypes, ActionWriteConfig)
}

func TestBuildPlan_AddPeer(t *testing.T) {
	desired := desiredTwoHosts()
	current := MeshState{
		Servers: map[string]*ServerState{
			"1.1.1.1": {
				Host:            "1.1.1.1",
				Installed:       true,
				KeysExist:       true,
				PublicKey:       "AAAAAAAA=",
				WireGuardMgmtIP: net.ParseIP("100.64.0.1").To4(),
				ListenPort:      51820,
				Active:          true,
				Peers:           []Peer{},
			},
			"2.2.2.2": {
				Host:            "2.2.2.2",
				Installed:       true,
				KeysExist:       true,
				PublicKey:       "BBBBBBBB=",
				WireGuardMgmtIP: net.ParseIP("100.64.0.2").To4(),
				ListenPort:      51820,
				Active:          true,
				Peers:           []Peer{{PublicKey: "AAAAAAAA="}},
			},
		},
	}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)

	var types []ActionType
	for _, a := range plan.Actions {
		if a.Host == "1.1.1.1" {
			types = append(types, a.Type)
		}
	}
	assert.Contains(t, types, ActionAddPeer)
	assert.Contains(t, types, ActionWriteConfig)
	assert.Contains(t, types, ActionReloadService)
}

func TestBuildPlan_RemovePeer(t *testing.T) {
	desired := &DesiredMesh{
		Hosts:           []string{"1.1.1.1"},
		Interface:       "wg0",
		MgmtPool:        defaultMgmtPool,
		ContainerPool:   defaultContainerPool,
		ContainerPrefix: 24,
		ListenPort:      51820,
	}
	current := MeshState{
		Servers: map[string]*ServerState{
			"1.1.1.1": {
				Host:            "1.1.1.1",
				Installed:       true,
				KeysExist:       true,
				PublicKey:       "AAAAAAAA=",
				WireGuardMgmtIP: net.ParseIP("100.64.0.1").To4(),
				ListenPort:      51820,
				Active:          true,
				Peers:           []Peer{{PublicKey: "STALEKEY="}},
			},
		},
	}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)

	var types []ActionType
	for _, a := range plan.Actions {
		if a.Host == "1.1.1.1" {
			types = append(types, a.Type)
		}
	}
	assert.Contains(t, types, ActionRemovePeer)
	assert.Contains(t, types, ActionWriteConfig)
}

func TestBuildPlan_StableMgmtAndContainerAssignments(t *testing.T) {
	desired := desiredWithPodman()
	current := MeshState{
		Servers: map[string]*ServerState{
			"1.1.1.1": {
				Host:            "1.1.1.1",
				WireGuardMgmtIP: net.ParseIP("100.64.0.7").To4(),
				Namespaces: map[string]*NamespaceServerState{
					DefaultNamespace: {
						Namespace:       DefaultNamespace,
						NetworkExists:   true,
						ContainerSubnet: mustParseCIDR("10.210.5.0/24"),
						Label:           DefaultNamespace,
					},
				},
			},
			"2.2.2.2": {
				Host:            "2.2.2.2",
				WireGuardMgmtIP: nil,
			},
		},
	}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)

	assert.Equal(t, "100.64.0.7", plan.MgmtAssignments["1.1.1.1"].String())
	assert.Equal(t, "10.210.5.0/24", plan.SubnetAssignments[DefaultNamespace]["1.1.1.1"].String())
}

func TestBuildPlan_PodmanFullStack(t *testing.T) {
	desired := desiredWithPodman()
	current := MeshState{Servers: map[string]*ServerState{}}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)

	collect := func(host string) []ActionType {
		var out []ActionType
		for _, a := range plan.Actions {
			if a.Host == host {
				out = append(out, a.Type)
			}
		}
		return out
	}

	for _, h := range []string{"1.1.1.1", "2.2.2.2"} {
		types := collect(h)
		assert.Contains(t, types, ActionInstallPodman, h)
		assert.Contains(t, types, ActionEnablePodmanSocket, h)
		assert.Contains(t, types, ActionEnableIPForward, h)
		assert.Contains(t, types, ActionCreatePodmanNet, h)
		assert.Contains(t, types, ActionInstallFirewall, h)
		assert.Contains(t, types, ActionAllocateContainerSubnet, h)
	}
}

func TestBuildPlan_PodmanIdempotent(t *testing.T) {
	desired := desiredWithPodman()
	current := MeshState{
		Servers: map[string]*ServerState{
			"1.1.1.1": convergedServer("1.1.1.1", "AAAAAAAA=", "BBBBBBBB=", "100.64.0.1", "10.210.0.0/24"),
			"2.2.2.2": convergedServer("2.2.2.2", "BBBBBBBB=", "AAAAAAAA=", "100.64.0.2", "10.210.1.0/24"),
		},
	}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)
	assert.True(t, plan.IsEmpty(), "expected empty plan, got: %+v", plan.Actions)
}

func TestBuildPlan_PodmanNotRequested(t *testing.T) {
	desired := desiredTwoHosts() // InstallPodman == false
	current := MeshState{
		Servers: map[string]*ServerState{
			"1.1.1.1": {
				Host:            "1.1.1.1",
				Installed:       true,
				KeysExist:       true,
				PublicKey:       "AAAAAAAA=",
				WireGuardMgmtIP: net.ParseIP("100.64.0.1").To4(),
				ListenPort:      51820,
				Active:          true,
				Peers:           []Peer{{PublicKey: "BBBBBBBB="}},
			},
			"2.2.2.2": {
				Host:            "2.2.2.2",
				Installed:       true,
				KeysExist:       true,
				PublicKey:       "BBBBBBBB=",
				WireGuardMgmtIP: net.ParseIP("100.64.0.2").To4(),
				ListenPort:      51820,
				Active:          true,
				Peers:           []Peer{{PublicKey: "AAAAAAAA="}},
			},
		},
	}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)
	for _, a := range plan.Actions {
		assert.NotEqual(t, ActionInstallPodman, a.Type)
		assert.NotEqual(t, ActionEnablePodmanSocket, a.Type)
		assert.NotEqual(t, ActionEnableIPForward, a.Type)
		assert.NotEqual(t, ActionCreatePodmanNet, a.Type)
		assert.NotEqual(t, ActionInstallFirewall, a.Type)
		assert.NotEqual(t, ActionAllocateContainerSubnet, a.Type)
	}
}

func TestBuildPlan_PodmanDNSEnabledTriggersRecreate(t *testing.T) {
	desired := desiredWithPodman()
	srvA := convergedServer("1.1.1.1", "AAAAAAAA=", "BBBBBBBB=", "100.64.0.1", "10.210.0.0/24")
	srvA.Namespaces[DefaultNamespace].DNSEnabled = true // drift: aardvark-dns would squat :53
	srvB := convergedServer("2.2.2.2", "BBBBBBBB=", "AAAAAAAA=", "100.64.0.2", "10.210.1.0/24")
	current := MeshState{Servers: map[string]*ServerState{"1.1.1.1": srvA, "2.2.2.2": srvB}}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)

	var aTypes, bTypes []ActionType
	for _, a := range plan.Actions {
		if a.Host == "1.1.1.1" {
			aTypes = append(aTypes, a.Type)
		}
		if a.Host == "2.2.2.2" {
			bTypes = append(bTypes, a.Type)
		}
	}

	assert.Contains(t, aTypes, ActionRecreatePodmanNet, "host A must recreate (dns_enabled=true)")
	assert.NotContains(t, aTypes, ActionCreatePodmanNet, "host A already exists — only recreate")
	assert.NotContains(t, bTypes, ActionRecreatePodmanNet, "host B fine, no recreate")
	assert.NotContains(t, bTypes, ActionCreatePodmanNet, "host B fine, no create")
}

func TestBuildPlan_FirewallMissing(t *testing.T) {
	desired := desiredWithPodman()
	srvA := convergedServer("1.1.1.1", "AAAAAAAA=", "BBBBBBBB=", "100.64.0.1", "10.210.0.0/24")
	srvA.FirewallActive = false
	current := MeshState{
		Servers: map[string]*ServerState{
			"1.1.1.1": srvA,
			"2.2.2.2": convergedServer("2.2.2.2", "BBBBBBBB=", "AAAAAAAA=", "100.64.0.2", "10.210.1.0/24"),
		},
	}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)

	var aTypes []ActionType
	for _, a := range plan.Actions {
		if a.Host == "1.1.1.1" {
			aTypes = append(aTypes, a.Type)
		}
	}
	assert.Equal(t, []ActionType{ActionInstallFirewall}, aTypes)
}

func TestBuildPlan_NftUnavailable_ReturnsError(t *testing.T) {
	desired := desiredWithPodman()
	desired.DefaultDenyContainers = true

	current := MeshState{
		Servers: map[string]*ServerState{
			"1.1.1.1": {
				Host:         "1.1.1.1",
				NftAvailable: false,
			},
			"2.2.2.2": {
				Host:         "2.2.2.2",
				NftAvailable: false,
			},
		},
	}

	_, err := BuildPlan(desired, current)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nft binary not available")
}

func TestBuildPlan_DefaultDenyRequiresPodman(t *testing.T) {
	desired := desiredTwoHosts()
	desired.DefaultDenyContainers = true // InstallPodman left false

	_, err := BuildPlan(desired, MeshState{Servers: map[string]*ServerState{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--default-deny requires --podman")
}

func TestBuildPlan_DefaultDenyDriftReinstalls(t *testing.T) {
	desired := desiredWithPodman()
	desired.DefaultDenyContainers = true

	// Both hosts converged in mode A (default-deny OFF) — must reinstall to flip on.
	srvA := convergedServer("1.1.1.1", "AAAAAAAA=", "BBBBBBBB=", "100.64.0.1", "10.210.0.0/24")
	srvA.DefaultDenyActive = false
	srvA.NftAvailable = true
	srvB := convergedServer("2.2.2.2", "BBBBBBBB=", "AAAAAAAA=", "100.64.0.2", "10.210.1.0/24")
	srvB.DefaultDenyActive = false
	srvB.NftAvailable = true
	current := MeshState{Servers: map[string]*ServerState{"1.1.1.1": srvA, "2.2.2.2": srvB}}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)

	for _, h := range []string{"1.1.1.1", "2.2.2.2"} {
		var found bool
		for _, a := range plan.Actions {
			if a.Host == h && a.Type == ActionInstallFirewall {
				found = true
				break
			}
		}
		assert.True(t, found, "expected ActionInstallFirewall for %s", h)
	}
}

func TestBuildPlan_DefaultDenyConverged(t *testing.T) {
	desired := desiredWithPodman()
	desired.DefaultDenyContainers = true

	srvA := convergedServer("1.1.1.1", "AAAAAAAA=", "BBBBBBBB=", "100.64.0.1", "10.210.0.0/24")
	srvA.DefaultDenyActive = true
	srvA.NftAvailable = true
	srvA.FirewallUnitSha256 = sha256Hex([]byte(FirewallServiceUnit("wg0",
		[]string{"default"}, []*net.IPNet{mustParseCIDR("10.210.0.0/24")}, true)))
	srvB := convergedServer("2.2.2.2", "BBBBBBBB=", "AAAAAAAA=", "100.64.0.2", "10.210.1.0/24")
	srvB.DefaultDenyActive = true
	srvB.NftAvailable = true
	srvB.FirewallUnitSha256 = sha256Hex([]byte(FirewallServiceUnit("wg0",
		[]string{"default"}, []*net.IPNet{mustParseCIDR("10.210.1.0/24")}, true)))
	current := MeshState{Servers: map[string]*ServerState{"1.1.1.1": srvA, "2.2.2.2": srvB}}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)
	assert.True(t, plan.IsEmpty(), "expected empty plan, got: %+v", plan.Actions)
}

func TestBuildPlan_SurfacesWarnings(t *testing.T) {
	desired := desiredTwoHosts()
	current := MeshState{
		Servers: map[string]*ServerState{
			"1.1.1.1": {Host: "1.1.1.1", WireGuardMgmtIP: net.ParseIP("100.64.0.5").To4()},
			"2.2.2.2": {Host: "2.2.2.2", WireGuardMgmtIP: net.ParseIP("100.64.0.5").To4()},
		},
	}

	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)
	assert.NotEmpty(t, plan.Warnings, "expected warning for duplicate mgmt IP")
}

func TestBuildPlan_MultiNamespacePlansPerNamespace(t *testing.T) {
	desired := desiredWithPodman()
	desired.Namespaces = []string{DefaultNamespace, "alpha"}

	current := MeshState{Servers: map[string]*ServerState{}}
	plan, err := BuildPlan(desired, current)
	require.NoError(t, err)

	// Two hosts × two namespaces = four create-podman-net actions.
	var creates []PlannedAction
	for _, a := range plan.Actions {
		if a.Type == ActionCreatePodmanNet {
			creates = append(creates, a)
		}
	}
	assert.Len(t, creates, 4)

	namespaces := map[string]bool{}
	for _, a := range creates {
		namespaces[a.Namespace] = true
	}
	assert.True(t, namespaces[DefaultNamespace])
	assert.True(t, namespaces["alpha"])

	// SubnetAssignments is namespace → host → subnet.
	assert.NotNil(t, plan.SubnetAssignments[DefaultNamespace])
	assert.NotNil(t, plan.SubnetAssignments["alpha"])
	assert.NotEqual(t, plan.SubnetAssignments[DefaultNamespace]["1.1.1.1"].String(),
		plan.SubnetAssignments["alpha"]["1.1.1.1"].String(),
		"namespaces must carve disjoint subnets")
}

func TestBuildPlan_PodmanRequiresNamespace(t *testing.T) {
	desired := desiredTwoHosts()
	desired.InstallPodman = true
	// no namespaces set

	_, err := BuildPlan(desired, MeshState{Servers: map[string]*ServerState{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "namespace")
}

func TestBinaryVersionDrift(t *testing.T) {
	tests := []struct {
		name            string
		desiredVersion  string
		installed       bool
		haveVersion     string
		wantDrift       bool
	}{
		{"not installed", "nightly", false, "", true},
		{"installed no marker", "nightly", true, "", true},
		{"nightly always drifts", "nightly", true, "nightly", true},
		{"pinned matches", "v1.2.3", true, "v1.2.3", false},
		{"pinned mismatch", "v1.2.4", true, "v1.2.3", true},
		{"pinned no marker", "v1.2.3", true, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := binaryVersionDrift(tt.desiredVersion, tt.installed, tt.haveVersion)
			assert.Equal(t, tt.wantDrift, got)
		})
	}
}

func TestBuildPlan_CooldVersionDrift(t *testing.T) {
	desired := desiredWithPodman()
	desired.InstallCoold = true
	desired.CooldVersion = "v1.2.3"
	desired.CorrosionVersion = "v1.2.3"
	desired.CorrosionGossipPort = 8787
	desired.CorrosionAPIPort = 8080

	host := "1.1.1.1"
	sn := mustParseCIDR("10.210.0.0/24")
	fwHash := sha256Hex([]byte(FirewallServiceUnit("wg0", []string{"default"}, []*net.IPNet{sn}, false)))
	state := &ServerState{
		Host: host, Installed: true, KeysExist: true, Active: true,
		PodmanInstalled: true, PodmanSocketActive: true, IPForwardEnabled: true,
		FirewallActive: true, DefaultDenyActive: false, FirewallUnitSha256: fwHash,
		CorrosionInstalled: true, CooldInstalled: true,
		CorrosionVersion: "v1.2.3", CooldVersion: "v1.2.2", // coold is stale
		CorrosionActive: true, CooldActive: true,
		Namespaces: map[string]*NamespaceServerState{
			DefaultNamespace: {Namespace: DefaultNamespace, NetworkExists: true, ContainerSubnet: sn, Label: DefaultNamespace},
		},
	}

	plan, err := BuildPlan(desired, MeshState{Servers: map[string]*ServerState{host: state}})
	require.NoError(t, err)

	types := make(map[ActionType]bool)
	for _, a := range plan.Actions {
		if a.Host == host {
			types[a.Type] = true
		}
	}
	assert.True(t, types[ActionInstallCoold], "stale coold version should trigger install-coold")
	assert.False(t, types[ActionInstallCorrosion], "matching corrosion version should not trigger install")
}

func TestBuildPlan_CooldNightlyAlwaysDrifts(t *testing.T) {
	desired := desiredWithPodman()
	desired.InstallCoold = true
	desired.CooldVersion = "nightly"
	desired.CorrosionVersion = "nightly"
	desired.CorrosionGossipPort = 8787
	desired.CorrosionAPIPort = 8080

	host := "1.1.1.1"
	sn := mustParseCIDR("10.210.0.0/24")
	fwHash := sha256Hex([]byte(FirewallServiceUnit("wg0", []string{"default"}, []*net.IPNet{sn}, false)))
	state := &ServerState{
		Host: host, Installed: true, KeysExist: true, Active: true,
		PodmanInstalled: true, PodmanSocketActive: true, IPForwardEnabled: true,
		FirewallActive: true, DefaultDenyActive: false, FirewallUnitSha256: fwHash,
		CorrosionInstalled: true, CooldInstalled: true,
		CorrosionVersion: "nightly", CooldVersion: "nightly",
		CorrosionActive: true, CooldActive: true,
		Namespaces: map[string]*NamespaceServerState{
			DefaultNamespace: {Namespace: DefaultNamespace, NetworkExists: true, ContainerSubnet: sn, Label: DefaultNamespace},
		},
	}

	plan, err := BuildPlan(desired, MeshState{Servers: map[string]*ServerState{host: state}})
	require.NoError(t, err)

	types := make(map[ActionType]bool)
	for _, a := range plan.Actions {
		if a.Host == host {
			types[a.Type] = true
		}
	}
	assert.True(t, types[ActionInstallCoold], "nightly tag always triggers install-coold")
	assert.True(t, types[ActionInstallCorrosion], "nightly tag always triggers install-corrosion")
}

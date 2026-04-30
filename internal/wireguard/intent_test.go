package wireguard

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateIntent_Bootstrap(t *testing.T) {
	d := &DesiredMesh{Intent: IntentBootstrap}
	require.NoError(t, ValidateIntent(d))
}

func TestValidateIntent_ExtendRequiresNewHosts(t *testing.T) {
	d := &DesiredMesh{
		Intent: IntentExtend,
		Hosts:  []string{"A", "B"},
	}
	err := ValidateIntent(d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NewHosts")
}

func TestValidateIntent_ExtendNewHostMustBeInServers(t *testing.T) {
	d := &DesiredMesh{
		Intent:   IntentExtend,
		Hosts:    []string{"A", "B"},
		NewHosts: []string{"C"},
	}
	err := ValidateIntent(d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"C"`)
	assert.Contains(t, err.Error(), "--servers")
}

func TestValidateIntent_ExtendHappy(t *testing.T) {
	d := &DesiredMesh{
		Intent:   IntentExtend,
		Hosts:    []string{"A", "B", "C"},
		NewHosts: []string{"C"},
	}
	require.NoError(t, ValidateIntent(d))
}

func TestValidateIntent_UpgradeRejectsNightlyByDefault(t *testing.T) {
	for _, tc := range []struct {
		name string
		d    DesiredMesh
	}{
		{"coold", DesiredMesh{Intent: IntentUpgrade, CooldVersion: "nightly", CorrosionVersion: "v1", SchedulerVersion: "v1"}},
		{"corrosion", DesiredMesh{Intent: IntentUpgrade, CooldVersion: "v1", CorrosionVersion: "nightly", SchedulerVersion: "v1"}},
		{"scheduler", DesiredMesh{Intent: IntentUpgrade, CooldVersion: "v1", CorrosionVersion: "v1", SchedulerVersion: "nightly"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateIntent(&tc.d)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "nightly")
		})
	}
}

func TestValidateIntent_UpgradeAllowsNightlyWhenOpted(t *testing.T) {
	d := &DesiredMesh{
		Intent:           IntentUpgrade,
		CooldVersion:     "nightly",
		CorrosionVersion: "nightly",
		SchedulerVersion:    "nightly",
		AllowNightly:     true,
	}
	require.NoError(t, ValidateIntent(d))
}

func TestValidateIntent_UpgradeAllowsPinned(t *testing.T) {
	d := &DesiredMesh{
		Intent:           IntentUpgrade,
		CooldVersion:     "v1.2.3",
		CorrosionVersion: "v0.9.0",
		SchedulerVersion:    "v0.3.0",
	}
	require.NoError(t, ValidateIntent(d))
}

func TestValidateIntent_UnknownIntent(t *testing.T) {
	d := &DesiredMesh{Intent: Intent("bogus")}
	err := ValidateIntent(d)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus")
}

func TestCategorize(t *testing.T) {
	cases := []struct {
		t    ActionType
		want actionCategory
	}{
		{ActionInstallWG, catSafeAlways},
		{ActionGenKeyPair, catSafeAlways},
		{ActionAllocateMgmtIP, catSafeAlways},
		{ActionAllocateContainerSubnet, catSafeAlways},
		{ActionEnableService, catSafeAlways},
		{ActionInstallPodman, catSafeAlways},
		{ActionEnablePodmanSocket, catSafeAlways},
		{ActionEnableIPForward, catSafeAlways},
		{ActionCreatePodmanNet, catSafeAlways},
		{ActionGenerateJWTKeypair, catSafeAlways},
		{ActionAddPeer, catSafeAlways},
		{ActionRemovePeer, catSafeAlways},

		{ActionWriteConfig, catPeerRefresh},
		{ActionReloadService, catPeerRefresh},
		{ActionInstallFirewall, catPeerRefresh},
		{ActionWriteCorrosionConfig, catPeerRefresh},
		{ActionInstallCorrosionService, catPeerRefresh},
		{ActionInstallCooldService, catPeerRefresh},
		{ActionInstallSchedulerService, catPeerRefresh},
		{ActionWriteHostJWT, catPeerRefresh},
		{ActionUpdateCooldSchedulerEnv, catPeerRefresh},

		{ActionRecreatePodmanNet, catDestructiveReplace},

		{ActionInstallCorrosion, catVersionBump},
		{ActionInstallCoold, catVersionBump},
		{ActionInstallScheduler, catVersionBump},
		{ActionInstallBuilder, catVersionBump},
	}
	for _, tc := range cases {
		t.Run(string(tc.t), func(t *testing.T) {
			assert.Equal(t, tc.want, categorize(PlannedAction{Type: tc.t}))
		})
	}
}

func TestCategorize_SchemaWipeVsFirstWrite(t *testing.T) {
	firstWrite := PlannedAction{
		Type:   ActionWriteCorrosionSchema,
		Detail: "/etc/corrosion/schemas/coolify.sql",
	}
	wipe := PlannedAction{
		Type:   ActionWriteCorrosionSchema,
		Detail: "/etc/corrosion/schemas/coolify.sql [schema drift — DB will be reset]",
	}
	assert.Equal(t, catCorrosionSchemaFirstWrite, categorize(firstWrite))
	assert.Equal(t, catWipeDB, categorize(wipe))
}

func TestFilterByIntent_BootstrapNoop(t *testing.T) {
	plan := &Plan{Actions: []PlannedAction{
		{Host: "A", Type: ActionInstallCoold},
		{Host: "B", Type: ActionRecreatePodmanNet},
		{Host: "B", Type: ActionWriteCorrosionSchema, Detail: "DB will be reset"},
	}}
	filterByIntent(plan, &DesiredMesh{Intent: IntentBootstrap})
	assert.Len(t, plan.Actions, 3)
	assert.Empty(t, plan.Skipped)
}

func TestFilterByIntent_ExtendNewHostRunsEverything(t *testing.T) {
	plan := &Plan{Actions: []PlannedAction{
		{Host: "A-new", Type: ActionInstallCoold},
		{Host: "A-new", Type: ActionInstallCorrosion},
		{Host: "A-new", Type: ActionCreatePodmanNet},
		{Host: "A-new", Type: ActionWriteCorrosionSchema, Detail: "first write"},
	}}
	filterByIntent(plan, &DesiredMesh{
		Intent:   IntentExtend,
		NewHosts: []string{"A-new"},
	})
	assert.Len(t, plan.Actions, 4)
	assert.Empty(t, plan.Skipped)
}

func TestFilterByIntent_ExtendExistingHostPeerRefreshOnly(t *testing.T) {
	plan := &Plan{Actions: []PlannedAction{
		{Host: "A-old", Type: ActionWriteConfig},
		{Host: "A-old", Type: ActionReloadService},
		{Host: "A-old", Type: ActionWriteCorrosionConfig},
		{Host: "A-old", Type: ActionInstallFirewall},
		{Host: "A-old", Type: ActionInstallCoold},    // version bump: skipped
		{Host: "A-old", Type: ActionInstallBuilder},  // version bump: skipped
		{Host: "A-new", Type: ActionInstallCoold},    // new host: kept
	}}
	filterByIntent(plan, &DesiredMesh{
		Intent:   IntentExtend,
		NewHosts: []string{"A-new"},
	})

	kept := map[ActionType]bool{}
	for _, a := range plan.Actions {
		kept[a.Type] = true
	}
	assert.True(t, kept[ActionWriteConfig])
	assert.True(t, kept[ActionReloadService])
	assert.True(t, kept[ActionWriteCorrosionConfig])
	assert.True(t, kept[ActionInstallFirewall])

	skippedTypes := map[ActionType]int{}
	for _, s := range plan.Skipped {
		skippedTypes[s.Action.Type]++
	}
	// InstallCoold/InstallBuilder appear once for the existing host and kept once for the new host.
	assert.Equal(t, 1, skippedTypes[ActionInstallCoold])
	assert.Equal(t, 1, skippedTypes[ActionInstallBuilder])

	// Exactly one InstallCoold survived — the one targeting the new host.
	var survivors []string
	for _, a := range plan.Actions {
		if a.Type == ActionInstallCoold {
			survivors = append(survivors, a.Host)
		}
	}
	assert.Equal(t, []string{"A-new"}, survivors)
}

func TestFilterByIntent_ExtendBlocksDestructiveOnExistingWithoutAllowReplace(t *testing.T) {
	plan := &Plan{Actions: []PlannedAction{
		{Host: "A-old", Type: ActionRecreatePodmanNet, Detail: "coolify-default-mesh — dns_enabled=true"},
	}}
	filterByIntent(plan, &DesiredMesh{
		Intent:   IntentExtend,
		NewHosts: []string{"A-new"},
	})
	assert.Empty(t, plan.Actions)
	require.Len(t, plan.Skipped, 1)
	assert.Contains(t, plan.Skipped[0].Reason, "--allow-replace")
}

func TestFilterByIntent_ExtendAllowReplaceUnlocksDestructive(t *testing.T) {
	plan := &Plan{Actions: []PlannedAction{
		{Host: "A-old", Type: ActionRecreatePodmanNet},
	}}
	filterByIntent(plan, &DesiredMesh{
		Intent:       IntentExtend,
		NewHosts:     []string{"A-new"},
		AllowReplace: true,
	})
	assert.Len(t, plan.Actions, 1)
	assert.Empty(t, plan.Skipped)
}

func TestFilterByIntent_ExtendAllowReplaceDoesNotUnlockWipeDB(t *testing.T) {
	plan := &Plan{Actions: []PlannedAction{
		{Host: "A-old", Type: ActionWriteCorrosionSchema, Detail: "schema drift — DB will be reset"},
	}}
	filterByIntent(plan, &DesiredMesh{
		Intent:       IntentExtend,
		NewHosts:     []string{"A-new"},
		AllowReplace: true,
	})
	assert.Empty(t, plan.Actions)
	require.Len(t, plan.Skipped, 1)
	assert.Contains(t, plan.Skipped[0].Reason, "never allowed")
}

func TestFilterByIntent_UpgradeOnlyKeepsVersionBumpsAndServiceRestarts(t *testing.T) {
	plan := &Plan{Actions: []PlannedAction{
		{Host: "A", Type: ActionInstallCoold},
		{Host: "A", Type: ActionInstallCorrosion},
		{Host: "A", Type: ActionInstallScheduler},
		{Host: "A", Type: ActionInstallBuilder},
		{Host: "A", Type: ActionInstallCooldService},
		{Host: "A", Type: ActionInstallCorrosionService},
		{Host: "A", Type: ActionInstallSchedulerService},
		{Host: "A", Type: ActionWriteConfig},        // skipped
		{Host: "A", Type: ActionReloadService},      // skipped
		{Host: "A", Type: ActionCreatePodmanNet},    // skipped
		{Host: "A", Type: ActionRecreatePodmanNet},  // skipped
		{Host: "A", Type: ActionInstallFirewall},    // skipped (non-restart peer-refresh)
	}}
	filterByIntent(plan, &DesiredMesh{Intent: IntentUpgrade})

	kept := map[ActionType]bool{}
	for _, a := range plan.Actions {
		kept[a.Type] = true
	}
	for _, want := range []ActionType{
		ActionInstallCoold, ActionInstallCorrosion, ActionInstallScheduler, ActionInstallBuilder,
		ActionInstallCooldService, ActionInstallCorrosionService, ActionInstallSchedulerService,
	} {
		assert.True(t, kept[want], "expected %s kept in upgrade", want)
	}

	skippedTypes := map[ActionType]bool{}
	for _, s := range plan.Skipped {
		skippedTypes[s.Action.Type] = true
	}
	for _, want := range []ActionType{
		ActionWriteConfig, ActionReloadService, ActionCreatePodmanNet, ActionRecreatePodmanNet, ActionInstallFirewall,
	} {
		assert.True(t, skippedTypes[want], "expected %s skipped in upgrade", want)
	}
}

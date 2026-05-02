package wireguard

import (
	"fmt"
	"strings"
)

// actionCategory classifies every ActionType so the intent filter can decide
// whether to allow, skip, or block it per host.
type actionCategory int

const (
	// catSafeAlways: pure add/first-time install. Idempotent, no runtime
	// disruption on re-run. Included in every intent.
	catSafeAlways actionCategory = iota

	// catPeerRefresh: rewrites a config or restarts a service as part of
	// keeping peer/namespace state in sync. Idempotent, short service blip
	// at worst. Allowed in every intent (extend needs it on existing hosts
	// to pick up the new peer's AllowedIPs; upgrade needs it for the
	// post-install service restart).
	catPeerRefresh

	// catDestructiveReplace: recreates a resource that may currently be in
	// use (running containers on a podman bridge). Blocked on existing
	// hosts in extend mode unless --allow-replace is set. Always blocked
	// in upgrade mode.
	catDestructiveReplace

	// catVersionBump: re-downloads an agent binary (coold, corrosion,
	// scheduler, builder). Runs on new hosts in extend mode (first install)
	// but not on existing hosts. Always allowed in upgrade mode.
	catVersionBump

	// catWipeDB: special-case for ActionWriteCorrosionSchema when the
	// schema drift branch fires (pre-existing sqlite DB gets deleted).
	// Only allowed in bootstrap mode and on brand-new hosts in extend
	// mode. Never allowed on existing hosts, even with --allow-replace.
	catWipeDB

	// catCorrosionSchemaFirstWrite: ActionWriteCorrosionSchema when no
	// prior schema is present (CorrosionSchemaSha256 is empty). Safe
	// everywhere because nothing gets wiped.
	catCorrosionSchemaFirstWrite
)

// categorize returns the category for a planned action. The schema action is
// looked up contextually (plan fills Detail with "[schema drift — DB will be
// reset]" when the wipe branch applies).
func categorize(a PlannedAction) actionCategory {
	switch a.Type {
	case ActionInstallWG,
		ActionGenKeyPair,
		ActionAllocateMgmtIP,
		ActionAllocateContainerSubnet,
		ActionEnableService,
		ActionInstallPodman,
		ActionEnablePodmanSocket,
		ActionEnableIPForward,
		ActionCreatePodmanNet,
		ActionGenerateJWTKeypair,
		ActionAddPeer,
		ActionRemovePeer:
		return catSafeAlways
	case ActionWriteConfig,
		ActionReloadService,
		ActionInstallFirewall,
		ActionWriteCorrosionConfig,
		ActionInstallCorrosionService,
		ActionInstallCooldService,
		ActionInstallSchedulerService,
		ActionWriteHostJWT,
		ActionUpdateCooldSchedulerEnv:
		return catPeerRefresh
	case ActionRecreatePodmanNet:
		return catDestructiveReplace
	case ActionInstallCorrosion,
		ActionInstallCoold,
		ActionInstallScheduler,
		ActionInstallBuilder:
		return catVersionBump
	case ActionWriteCorrosionSchema:
		if strings.Contains(a.Detail, "DB will be reset") {
			return catWipeDB
		}
		return catCorrosionSchemaFirstWrite
	}
	return catSafeAlways
}

// ValidateIntent enforces pre-plan invariants the filter itself can't express.
func ValidateIntent(d *DesiredMesh) error {
	switch d.Intent {
	case IntentBootstrap:
		return nil
	case IntentExtend:
		if len(d.NewHosts) == 0 {
			return fmt.Errorf("extend mode requires at least one host in NewHosts")
		}
		hostSet := make(map[string]struct{}, len(d.Hosts))
		for _, h := range d.Hosts {
			hostSet[h] = struct{}{}
		}
		for _, nh := range d.NewHosts {
			if _, ok := hostSet[nh]; !ok {
				return fmt.Errorf("extend mode: new host %q not in --servers list", nh)
			}
		}
		return nil
	case IntentUpgrade:
		if !d.AllowNightly {
			for _, pair := range [][2]string{
				{"--coold-version", d.CooldVersion},
				{"--corrosion-version", d.CorrosionVersion},
				{"--scheduler-version", d.SchedulerVersion},
			} {
				if pair[1] == "nightly" {
					return fmt.Errorf(
						"upgrade mode rejects %s=nightly (moving target forces re-install every run); pin a version or pass --allow-nightly",
						pair[0],
					)
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown intent %q", d.Intent)
	}
}

// filterByIntent mutates plan.Actions in place, moving blocked/skipped actions
// into plan.Skipped with a reason. For IntentBootstrap (default) it is a no-op.
func filterByIntent(plan *Plan, d *DesiredMesh) {
	if d.Intent == IntentBootstrap {
		return
	}

	newHostSet := make(map[string]struct{}, len(d.NewHosts))
	for _, h := range d.NewHosts {
		newHostSet[h] = struct{}{}
	}

	kept := plan.Actions[:0]
	for _, a := range plan.Actions {
		reason := decide(a, d, newHostSet)
		if reason == "" {
			kept = append(kept, a)
			continue
		}
		plan.Skipped = append(plan.Skipped, SkippedAction{Action: a, Reason: reason})
	}
	plan.Actions = kept
}

// decide returns an empty string when the action should run, or a short
// human-readable reason when it should be skipped.
func decide(a PlannedAction, d *DesiredMesh, newHostSet map[string]struct{}) string {
	cat := categorize(a)
	_, isNewHost := newHostSet[a.Host]

	switch d.Intent {
	case IntentExtend:
		if isNewHost {
			// Everything runs on a brand-new host — it needs the full install.
			return ""
		}
		// Existing host in extend mode: only peer-refresh and safe-always
		// (whose guards prevent re-runs on converged hosts) actions run.
		switch cat {
		case catSafeAlways, catPeerRefresh:
			return ""
		case catDestructiveReplace:
			if d.AllowReplace {
				return ""
			}
			return "extend: destructive-replace on existing host blocked; pass --allow-replace to override"
		case catVersionBump:
			return "extend: version-bump on existing host skipped; use `coolify init upgrade` to bump versions"
		case catWipeDB:
			return "extend: corrosion DB wipe on existing host is never allowed; resolve schema drift with `coolify init upgrade` on a fresh schema"
		case catCorrosionSchemaFirstWrite:
			return ""
		}
	case IntentUpgrade:
		switch cat {
		case catVersionBump:
			return ""
		case catPeerRefresh:
			if isUpgradeServiceRestart(a.Type) {
				return ""
			}
			return "upgrade: peer-refresh skipped; use `coolify init extend` for mesh topology changes"
		case catSafeAlways, catDestructiveReplace, catWipeDB, catCorrosionSchemaFirstWrite:
			return "upgrade: non-version-bump action skipped"
		}
	default:
		// IntentBootstrap (and unknown intents) keep every action.
	}
	return ""
}

// isUpgradeServiceRestart returns true when a peer-refresh action is the
// follow-up systemctl restart after a binary install and must run in upgrade
// mode to pick up the new binary.
func isUpgradeServiceRestart(t ActionType) bool {
	switch t {
	case ActionInstallCorrosionService,
		ActionInstallCooldService,
		ActionInstallSchedulerService:
		return true
	default:
		return false
	}
}

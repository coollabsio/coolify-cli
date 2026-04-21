// Package common holds flag structs and helpers shared between the
// `coolify init` and `coolify firewall` command trees. Kept intentionally
// small: only cross-command plumbing (SSH mesh flags, namespace validation)
// lives here.
//
//nolint:revive // "common" is the conventional sharing point for these cobra subtrees
package common

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

// DefaultNamespace is the namespace used when the user does not pass
// --namespaces. It is also always present (implicitly) so existing workflows
// and coold defaults keep working.
const DefaultNamespace = "default"

// PodmanNetworkFor returns the podman bridge network name that backs
// namespace ns on every host. Derived as `coolify-<ns>-mesh` so the
// namespace name is visible in `podman network ls`.
func PodmanNetworkFor(ns string) string {
	return "coolify-" + ns + "-mesh"
}

// MeshNetFlags holds the flag set shared between `coolify init` (which creates
// per-namespace podman networks on every host) and `coolify firewall` (which
// talks to coold about per-namespace rules).
//
// `init` binds it as a slice so a single command sets up the entire cluster;
// `firewall` binds it as a single value since each allow/revoke/list call
// operates on one namespace at a time.
type MeshNetFlags struct {
	// Namespaces enumerates every namespace the mesh should carry. At least
	// one entry is required; the first element is the implicit "default"
	// unless the user overrides it.
	Namespaces []string

	// ContainerPool is the shared address pool every namespace carves its
	// per-host /<ContainerPrefix> from. One pool covers all namespaces;
	// subnets never overlap.
	ContainerPool string

	// ContainerPrefix is the prefix length of each per-host, per-namespace
	// container subnet (default 24 → 254 container IPs per host per ns).
	ContainerPrefix int
}

// BindMeshNetMultiFlags registers --namespaces/--container-pool/--container-prefix
// on cmd (init-style: many namespaces per invocation).
func BindMeshNetMultiFlags(cmd *cobra.Command, f *MeshNetFlags) {
	pf := cmd.PersistentFlags()
	pf.StringSliceVar(&f.Namespaces, "namespaces", []string{DefaultNamespace},
		"Comma-separated list of namespaces to create on each host. Each "+
			"namespace is a separate Podman bridge network (coolify-<ns>-mesh) "+
			"with its own /<container-prefix> per host")
	pf.StringVar(&f.ContainerPool, "container-pool", "10.210.0.0/16",
		"Shared container address pool — each (namespace, host) pair gets a "+
			"/<container-prefix> from here, owned by that namespace's Podman bridge")
	pf.IntVar(&f.ContainerPrefix, "container-prefix", 24,
		"Prefix length of each per-host, per-namespace container subnet")
}

// BindMeshNetSingleFlags registers --namespace on cmd (firewall-style: one
// namespace per invocation).
func BindMeshNetSingleFlags(cmd *cobra.Command, ns *string) {
	pf := cmd.PersistentFlags()
	pf.StringVar(ns, "namespace", DefaultNamespace,
		"Namespace the command operates against (must match a namespace created by `coolify init`)")
}

// namespaceRegex matches a valid DNS label (namespace names appear in the
// podman network name, in iptables chain names, and — post-coold-changes —
// as DNS labels like web.<ns>.coolify.internal).
var namespaceRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// ValidateNamespaces checks that every namespace is a valid DNS label and
// that the list has no duplicates.
func (f *MeshNetFlags) ValidateNamespaces() error {
	if len(f.Namespaces) == 0 {
		return fmt.Errorf("--namespaces must list at least one namespace")
	}
	seen := make(map[string]struct{}, len(f.Namespaces))
	for _, ns := range f.Namespaces {
		if !namespaceRegex.MatchString(ns) {
			return fmt.Errorf("invalid namespace %q (must be a DNS label: lowercase alphanumerics + '-', 1-63 chars)", ns)
		}
		if _, dup := seen[ns]; dup {
			return fmt.Errorf("duplicate namespace %q in --namespaces", ns)
		}
		seen[ns] = struct{}{}
	}
	return nil
}

// ValidateNamespace validates a single namespace value (used by the firewall
// command's --namespace flag).
func ValidateNamespace(ns string) error {
	if !namespaceRegex.MatchString(ns) {
		return fmt.Errorf("invalid --namespace %q (must be a DNS label: lowercase alphanumerics + '-', 1-63 chars)", ns)
	}
	return nil
}

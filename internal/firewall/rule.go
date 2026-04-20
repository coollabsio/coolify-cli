// Package firewall implements the `coolify firewall` command logic: per-host
// container discovery (SSH+podman) and the SSH-bounced REST client that
// drives the coold agent's firewall surface on each mesh host.
//
// Rule-rendering and iptables IO live entirely in coold now (see the coold
// repo, `src/firewall/`). The CLI's job is to resolve endpoints, compute
// stable rule identities, and POST/DELETE/GET against coold over SSH. Rules
// go on the host that owns the destination IP, matching CONTROL_PLANE.md §3.
package firewall

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

// AllowRule is a single cross-host container allow entry.
//
// The rule lives on the host that owns Dst's container subnet (the default-
// deny jump fires on `-d <subnet> -j COOLIFY-INTRA`). Src may belong to any
// host in the mesh. Proto/Port are optional; zero values mean "any".
type AllowRule struct {
	Host    string // host that owns Dst's container subnet
	Src     net.IP
	Dst     net.IP
	Proto   string // "tcp" | "udp" | ""
	Port    int    // 0 = any
	Comment string // "cid:<12-hex>" stable identity for list/revoke
}

// ComputeID returns a 12-hex stable identity hash over (src, dst, proto, port).
// Used as the rule comment so `list` can display it and `revoke --from ...
// --to ... --port ...` finds the right rule without needing to parse.
//
// Byte-compatible with coold's ComputeID_ (src/firewall/rule.rs): proto
// lowercased, empty when unset; port rendered as 0 when unset. Mixed
// writers (CLI + coold) produce identical IDs for identical tuples.
func ComputeID(src, dst net.IP, proto string, port int) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%d", src.String(), dst.String(), strings.ToLower(proto), port)
	return hex.EncodeToString(h.Sum(nil))[:12]
}

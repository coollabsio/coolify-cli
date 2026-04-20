// Package firewall implements the `coolify firewall` command logic:
// per-host container discovery, COOLIFY-ALLOW rule rendering and parsing,
// iptables apply/revoke over SSH, and reboot-persistence scaffolding.
//
// The package is the companion to `internal/wireguard`: it owns the dynamic
// allow-rule layer on top of the default-deny scaffold installed by
// `coolify init --podman --default-deny`. Rules go on the host that owns the
// destination IP (COOLIFY-INTRA fires on `-d <container-subnet>`), matching
// the design in CONTROL_PLANE.md §3.
package firewall

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// ChainName is the iptables chain the v5 control plane owns. `coolify init
// --default-deny` creates it empty; this package fills it.
const ChainName = "COOLIFY-ALLOW"

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
func ComputeID(src, dst net.IP, proto string, port int) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%d", src.String(), dst.String(), strings.ToLower(proto), port)
	return hex.EncodeToString(h.Sum(nil))[:12]
}

// matchArgs renders the common iptables match portion for this rule.
// Used by RenderAppend / RenderDelete / RenderCheck so they stay in sync.
func (r AllowRule) matchArgs() string {
	var b strings.Builder
	fmt.Fprintf(&b, "-s %s -d %s", r.Src.String(), r.Dst.String())
	if r.Proto != "" {
		fmt.Fprintf(&b, " -p %s", r.Proto)
		if r.Port > 0 {
			fmt.Fprintf(&b, " --dport %d", r.Port)
		}
	}
	if r.Comment != "" {
		fmt.Fprintf(&b, " -m comment --comment %q", r.Comment)
	}
	b.WriteString(" -j ACCEPT")
	return b.String()
}

// RenderAppend produces the `iptables -A COOLIFY-ALLOW ...` command.
func (r AllowRule) RenderAppend() string {
	return fmt.Sprintf("iptables -A %s %s", ChainName, r.matchArgs())
}

// RenderDelete produces the `iptables -D COOLIFY-ALLOW ...` command.
func (r AllowRule) RenderDelete() string {
	return fmt.Sprintf("iptables -D %s %s", ChainName, r.matchArgs())
}

// RenderCheck produces the `iptables -C COOLIFY-ALLOW ...` command,
// suitable as an idempotency guard (`-C ... || -A ...`).
func (r AllowRule) RenderCheck() string {
	return fmt.Sprintf("iptables -C %s %s", ChainName, r.matchArgs())
}

// chainLineRegex parses one `-A COOLIFY-ALLOW ...` line from `iptables -S`.
// Captures in order: src, dst, optional proto, optional dport, optional comment.
var (
	reSrc     = regexp.MustCompile(`-s (\S+?)(?:/32)?(?:\s|$)`)
	reDst     = regexp.MustCompile(`-d (\S+?)(?:/32)?(?:\s|$)`)
	reProto   = regexp.MustCompile(`-p (\S+)`)
	reDport   = regexp.MustCompile(`--dport (\d+)`)
	reComment = regexp.MustCompile(`--comment "([^"]*)"|--comment (\S+)`)
)

// ParseChainLine parses a single `-A COOLIFY-ALLOW ...` output line from
// `iptables -S COOLIFY-ALLOW` into an AllowRule. Returns (_, false) when
// the line is not an append or cannot be parsed (missing src/dst).
func ParseChainLine(line string) (AllowRule, bool) {
	line = strings.TrimSpace(line)
	prefix := "-A " + ChainName + " "
	if !strings.HasPrefix(line, prefix) {
		return AllowRule{}, false
	}
	rest := line[len(prefix):]

	srcMatch := reSrc.FindStringSubmatch(rest)
	dstMatch := reDst.FindStringSubmatch(rest)
	if len(srcMatch) < 2 || len(dstMatch) < 2 {
		return AllowRule{}, false
	}
	src := net.ParseIP(srcMatch[1])
	dst := net.ParseIP(dstMatch[1])
	if src == nil || dst == nil {
		return AllowRule{}, false
	}

	r := AllowRule{Src: src, Dst: dst}

	if m := reProto.FindStringSubmatch(rest); len(m) >= 2 {
		r.Proto = m[1]
	}
	if m := reDport.FindStringSubmatch(rest); len(m) >= 2 {
		if n, err := strconv.Atoi(m[1]); err == nil {
			r.Port = n
		}
	}
	if m := reComment.FindStringSubmatch(rest); len(m) >= 2 {
		if m[1] != "" {
			r.Comment = m[1]
		} else if len(m) >= 3 {
			r.Comment = m[2]
		}
	}
	return r, true
}

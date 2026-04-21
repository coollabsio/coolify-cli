package firewall

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/coollabsio/coolify-cli/internal/ssh"
)

// CooldAPIBasePath is the path prefix the coold REST router serves under.
// Mirrors `src/firewall/api.rs` in the coold repo.
const CooldAPIBasePath = "/api/v1/firewall"

// CooldAPITokenPath is the remote file coold reads its bearer token from.
// Kept in sync with internal/services/coold.go — the CLI falls back to
// reading this file over SSH when the user hasn't supplied --coold-token.
const CooldAPITokenPath = "/etc/coolify/api-token"

// FetchCooldToken SSHes into host and reads the coold bearer token at
// CooldAPITokenPath. Each host generates its own random token at install
// time (see EnsureCooldAPITokenCommand), so per-host fetch is the default
// path when the user hasn't provided a global --coold-token override.
func FetchCooldToken(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	sshPort int,
) (string, error) {
	cmd := "cat " + CooldAPITokenPath
	stdout, stderr, err := runner.Run(ctx, host, user, sshPort, cmd)
	if err != nil {
		return "", fmt.Errorf("fetch coold token from %s: %w (stderr: %s)",
			host, err, strings.TrimSpace(stderr))
	}
	tok := strings.TrimSpace(stdout)
	if tok == "" {
		return "", fmt.Errorf("coold token on %s is empty — is coold installed? (expected at %s)",
			host, CooldAPITokenPath)
	}
	return tok, nil
}

// cooldRulePayload mirrors the JSON shape coold's REST API expects on POST
// and returns on GET /allow. Kept aligned with coold/src/firewall/rule.rs:
// namespace is required (defaults to "default" on the wire), src/dst are
// string IPs, proto/port/id are omitted when absent.
type cooldRulePayload struct {
	Namespace string `json:"namespace"`
	Src       string `json:"src"`
	Dst       string `json:"dst"`
	Proto     string `json:"proto,omitempty"`
	Port      uint16 `json:"port,omitempty"`
	ID        string `json:"id,omitempty"`
}

// toAllowRule converts a payload coming back from coold into the CLI's
// AllowRule. The host field is filled in by the caller (it is the mesh host
// the list came from, not part of the payload).
func (p cooldRulePayload) toAllowRule() (AllowRule, bool) {
	src := net.ParseIP(p.Src)
	dst := net.ParseIP(p.Dst)
	if src == nil || dst == nil {
		return AllowRule{}, false
	}
	ns := p.Namespace
	if ns == "" {
		ns = "default"
	}
	r := AllowRule{
		Namespace: ns,
		Src:       src,
		Dst:       dst,
		Proto:     p.Proto,
		Port:      int(p.Port),
	}
	if p.ID != "" {
		r.Comment = "cid:" + p.ID
	}
	return r, true
}

// allowRulePayload converts an AllowRule into the wire shape coold accepts.
// coold normalizes and computes the id itself, so we send only the tuple.
// Empty namespace is materialized as "default" on the wire so older coold
// builds with a default-only schema keep working.
func allowRulePayload(r AllowRule) cooldRulePayload {
	ns := r.Namespace
	if ns == "" {
		ns = "default"
	}
	p := cooldRulePayload{
		Namespace: ns,
		Src:       r.Src.String(),
		Dst:       r.Dst.String(),
		Proto:     r.Proto,
	}
	if r.Port > 0 {
		p.Port = uint16(r.Port)
	}
	return p
}

// CooldApply POSTs r to coold's /allow endpoint on host. coold is reached
// via SSH-bounce: SSH into host, curl localhost wg0 mgmt IP. This is the
// transport of choice for the alpha because the CLI runs on a laptop that
// isn't a mesh peer — only hosts inside the wg0 network can reach coold.
func CooldApply(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	sshPort, cooldPort int,
	iface, token string,
	r AllowRule,
) error {
	body, err := json.Marshal(allowRulePayload(r))
	if err != nil {
		return fmt.Errorf("marshal allow rule: %w", err)
	}
	cmd := buildCurlAllow(iface, token, cooldPort, string(body))
	if _, stderr, err := runner.Run(ctx, host, user, sshPort, cmd); err != nil {
		return fmt.Errorf("coold apply on %s: %w (stderr: %s)",
			host, err, strings.TrimSpace(stderr))
	}
	return nil
}

// CooldRevoke DELETEs rule id from coold on host. coold returns 204 even
// when the id is unknown, so missing rules are a silent no-op.
func CooldRevoke(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	sshPort, cooldPort int,
	iface, token, id string,
) error {
	if id == "" {
		return fmt.Errorf("coold revoke: empty id")
	}
	cmd := buildCurlRevoke(iface, token, cooldPort, id)
	if _, stderr, err := runner.Run(ctx, host, user, sshPort, cmd); err != nil {
		return fmt.Errorf("coold revoke on %s: %w (stderr: %s)",
			host, err, strings.TrimSpace(stderr))
	}
	return nil
}

// CooldList GETs coold's /allow endpoint on host and returns the parsed
// rules. An empty namespace means "all namespaces"; a non-empty value is
// forwarded to coold as `?namespace=<ns>`. Missing coold (no wg0 interface)
// is treated as an empty slice so a partially-deployed mesh doesn't break
// `firewall list`.
func CooldList(
	ctx context.Context,
	runner ssh.Runner,
	host, user string,
	sshPort, cooldPort int,
	iface, token, namespace string,
) ([]AllowRule, error) {
	cmd := buildCurlList(iface, token, cooldPort, namespace)
	stdout, stderr, err := runner.Run(ctx, host, user, sshPort, cmd)
	if err != nil {
		return nil, fmt.Errorf("coold list on %s: %w (stderr: %s)",
			host, err, strings.TrimSpace(stderr))
	}
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return nil, nil
	}
	var payloads []cooldRulePayload
	if err := json.Unmarshal([]byte(stdout), &payloads); err != nil {
		return nil, fmt.Errorf("parse coold list on %s: %w (body: %s)",
			host, err, stdout)
	}
	out := make([]AllowRule, 0, len(payloads))
	for _, p := range payloads {
		r, ok := p.toAllowRule()
		if !ok {
			continue
		}
		r.Host = host
		out = append(out, r)
	}
	return out, nil
}

// CooldListAll fans CooldList across every host in parallel and returns a
// stably-sorted flattened slice plus the per-host results. tokenFor is
// called once per host on its worker goroutine — fail here and the host
// surfaces as a ServerResult.Err instead of polluting the rule slice. An
// empty namespace forwards `?namespace=` omitted (coold returns all).
func CooldListAll(
	ctx context.Context,
	runner ssh.Runner,
	hosts []string,
	user string,
	sshPort, cooldPort int,
	iface string,
	tokenFor func(host string) (string, error),
	concurrency int,
	namespace string,
) ([]AllowRule, []ssh.ServerResult[[]AllowRule]) {
	results := ssh.ForEachServer(ctx, hosts, concurrency,
		func(ctx context.Context, host string) ([]AllowRule, error) {
			token, err := tokenFor(host)
			if err != nil {
				return nil, err
			}
			return CooldList(ctx, runner, host, user, sshPort, cooldPort, iface, token, namespace)
		})
	var all []AllowRule
	for _, r := range results {
		all = append(all, r.Result...)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Host != all[j].Host {
			return all[i].Host < all[j].Host
		}
		if all[i].Namespace != all[j].Namespace {
			return all[i].Namespace < all[j].Namespace
		}
		si, sj := all[i].Src.String(), all[j].Src.String()
		if si != sj {
			return si < sj
		}
		di, dj := all[i].Dst.String(), all[j].Dst.String()
		if di != dj {
			return di < dj
		}
		return all[i].Port < all[j].Port
	})
	return all, results
}

// shellSingleQuote wraps s in POSIX-shell single quotes, escaping any
// embedded single quotes. Used to embed JSON bodies and tokens into shell
// commands without breaking quoting.
func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// DefaultWGInterface is the WireGuard interface name the firewall CLI
// assumes when no override is supplied. Matches the default of
// `coolify init --wg-interface`.
const DefaultWGInterface = "wg0"

// mgmtIPScript discovers coold's bind IP on the remote host by reading the
// first IPv4 address on the host's WireGuard interface. Emitted as part of
// every curl command so the CLI doesn't need to track per-host mgmt IPs
// (they are already encoded in the host's own WG interface).
func mgmtIPScript(iface string) string {
	return fmt.Sprintf(
		`MGMT=$(ip -4 -o addr show %[1]s 2>/dev/null | awk '{print $4}' | cut -d/ -f1); `+
			`test -n "$MGMT" || { echo "coold mgmt IP (%[1]s) not found on $(hostname) — is coold installed?" >&2; exit 1; }; `,
		iface)
}

// mgmtIPScriptSoft is the same as mgmtIPScript but treats a missing WG
// interface as "no rules" rather than a failure. Used by list so a host
// without coold is simply absent from the output instead of aborting the
// whole fanout.
func mgmtIPScriptSoft(iface string) string {
	return fmt.Sprintf(
		`MGMT=$(ip -4 -o addr show %s 2>/dev/null | awk '{print $4}' | cut -d/ -f1); `+
			`if [ -z "$MGMT" ]; then echo '[]'; exit 0; fi; `,
		iface)
}

// buildCurlAllow returns the shell one-liner that POSTs body to coold.
// Token is embedded inline in the -H header; on the remote it is briefly
// visible in /proc/<curl-pid>/cmdline to root only, for the ~ms lifetime of
// the curl invocation. Acceptable for alpha; TLS + stdin-fed tokens are a
// follow-up.
func buildCurlAllow(iface, token string, port int, body string) string {
	return mgmtIPScript(iface) +
		`curl -fsS --max-time 10 ` +
		`-H ` + shellSingleQuote("Authorization: Bearer "+token) + ` ` +
		`-H 'Content-Type: application/json' ` +
		`-X POST -d ` + shellSingleQuote(body) + ` ` +
		fmt.Sprintf(`"http://$MGMT:%d%s/allow"`, port, CooldAPIBasePath)
}

// buildCurlRevoke returns the shell one-liner that DELETEs rule id.
func buildCurlRevoke(iface, token string, port int, id string) string {
	return mgmtIPScript(iface) +
		`curl -fsS --max-time 10 -o /dev/null ` +
		`-H ` + shellSingleQuote("Authorization: Bearer "+token) + ` ` +
		`-X DELETE ` +
		fmt.Sprintf(`"http://$MGMT:%d%s/allow/%s"`, port, CooldAPIBasePath, id)
}

// buildCurlList returns the shell one-liner that GETs /allow. A missing
// WG interface returns an empty JSON array so the caller sees "no rules"
// instead of a transport error. A non-empty namespace is forwarded as
// ?namespace=<ns>.
func buildCurlList(iface, token string, port int, namespace string) string {
	query := ""
	if namespace != "" {
		query = "?namespace=" + namespace
	}
	return mgmtIPScriptSoft(iface) +
		`curl -fsS --max-time 10 ` +
		`-H ` + shellSingleQuote("Authorization: Bearer "+token) + ` ` +
		fmt.Sprintf(`"http://$MGMT:%d%s/allow%s"`, port, CooldAPIBasePath, query)
}

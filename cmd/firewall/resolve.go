package firewall

import (
	"fmt"
	"net"
	"strings"

	ifw "github.com/coollabsio/coolify-cli/internal/firewall"
)

// resolveEndpoint turns a user-supplied reference (name, short-id, raw IP,
// or "host:name") into the container it points at. When ref is a raw IP
// that doesn't match any discovered container, it returns a synthetic
// entry with Host="" — the caller must derive Host some other way.
//
// Ambiguous names across hosts are rejected; the user must disambiguate
// with "host:name" or a short-ID.
func resolveEndpoint(ref string, all []ifw.Container) (ifw.Container, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ifw.Container{}, fmt.Errorf("empty container reference")
	}

	// "host:name" form — exact host disambiguator.
	if host, name, ok := splitHostName(ref); ok {
		for _, c := range all {
			if c.Host == host && c.Name == name {
				return c, nil
			}
		}
		return ifw.Container{}, fmt.Errorf("no container named %q on host %q", name, host)
	}

	// Raw IP form.
	if ip := net.ParseIP(ref); ip != nil {
		for _, c := range all {
			if c.IP.Equal(ip) {
				return c, nil
			}
		}
		// Synthetic: caller must decide on Host.
		return ifw.Container{IP: ip}, nil
	}

	// Name / short-id form. Collect matches, error on ambiguity.
	var matches []ifw.Container
	for _, c := range all {
		if c.Name == ref || strings.HasPrefix(c.ID, ref) {
			matches = append(matches, c)
		}
	}
	switch len(matches) {
	case 0:
		return ifw.Container{}, fmt.Errorf("no container matches %q", ref)
	case 1:
		return matches[0], nil
	default:
		return ifw.Container{}, fmt.Errorf(
			"reference %q is ambiguous across hosts (%s) — use host:name form",
			ref, hostList(matches))
	}
}

func splitHostName(ref string) (host, name string, ok bool) {
	i := strings.IndexByte(ref, ':')
	if i <= 0 || i == len(ref)-1 {
		return "", "", false
	}
	// Reject if the part after `:` looks like a port (all digits) — likely
	// an IP:port form the user didn't mean.
	name = ref[i+1:]
	host = ref[:i]
	if allDigits(name) {
		return "", "", false
	}
	return host, name, true
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func hostList(cs []ifw.Container) string {
	seen := map[string]bool{}
	var hosts []string
	for _, c := range cs {
		if !seen[c.Host] {
			hosts = append(hosts, c.Host)
			seen[c.Host] = true
		}
	}
	return strings.Join(hosts, ", ")
}

// findHostForIP returns the SSH host that owns ip (i.e. the host whose
// coolify-mesh bridge has ip assigned). Used when --to/--from is given as
// a raw IP not tied to a running container.
func findHostForIP(ip net.IP, all []ifw.Container) (string, bool) {
	for _, c := range all {
		if c.IP.Equal(ip) {
			return c.Host, true
		}
	}
	return "", false
}

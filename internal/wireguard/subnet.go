package wireguard

import (
	"encoding/binary"
	"fmt"
	"net"
	"sort"
)

// Warning describes a non-fatal conflict discovered during IP allocation.
type Warning struct {
	Host   string
	Reason string
}

// MachineIP returns the host address within a per-host subnet — the first
// usable IP (network address + 1).  For example, 10.210.5.0/24 → 10.210.5.1.
//
// Used for the Podman bridge gateway. WireGuard does NOT use this — wg0
// gets a separate /32 from the management pool (see AllocateMgmtIPs).
func MachineIP(subnet *net.IPNet) net.IP {
	return uint32ToIP(ipToUint32(subnet.IP.To4()) + 1)
}

// Allocate assigns a per-host subnet (of size hostPrefix) to every host in
// hosts, carving them from pool.
//
// Rules:
//   - Duplicate host names in hosts → error (user input bug).
//   - Existing subnet within pool with correct prefix → kept unchanged (stable).
//   - Existing subnet outside pool or wrong prefix → warning, reassign.
//   - Two existing hosts with the same subnet → first (alphabetical) kept,
//     second gets a warning and is reassigned.
//   - New hosts receive the lowest free subnet in pool.
//
// Returns (assignments, warnings, error).
func Allocate(
	pool *net.IPNet,
	hostPrefix int,
	existing map[string]*net.IPNet,
	hosts []string,
) (map[string]*net.IPNet, []Warning, error) {
	// 1. Dedup hosts.
	hostCount := make(map[string]int, len(hosts))
	for _, h := range hosts {
		hostCount[h]++
	}
	for h, n := range hostCount {
		if n > 1 {
			return nil, nil, fmt.Errorf("duplicate host in --servers: %s", h)
		}
	}

	pool4 := pool.IP.To4()
	if pool4 == nil {
		return nil, nil, fmt.Errorf("only IPv4 pools are supported")
	}

	result := make(map[string]*net.IPNet, len(hosts))
	usedNetworks := make(map[uint32]bool)
	var warnings []Warning

	subnetClaim := make(map[uint32]string)

	// 2. Seed from existing — sorted for deterministic conflict resolution.
	existingHosts := make([]string, 0, len(existing))
	for h := range existing {
		existingHosts = append(existingHosts, h)
	}
	sort.Strings(existingHosts)

	// Pool bounds (used for both validation and iteration).
	pool4Network := ipToUint32(pool4)
	poolOnes, poolBits := pool.Mask.Size()
	poolHostBits := poolBits - poolOnes
	pool4Broadcast := pool4Network | (uint32(1)<<uint(poolHostBits) - 1)

	for _, host := range existingHosts {
		subnet := existing[host]
		if subnet == nil {
			continue
		}
		subnet4 := subnet.IP.To4()
		ones, _ := subnet.Mask.Size()

		if subnet4 == nil || !pool.Contains(subnet4) || ones != hostPrefix {
			warnings = append(warnings, Warning{
				Host:   host,
				Reason: fmt.Sprintf("existing subnet %s is not a /%d inside pool %s, reassigning", subnet, hostPrefix, pool),
			})
			continue
		}

		networkU32 := ipToUint32(subnet4)

		// For /32 mgmt IPs, reject pool's network address (.0) and broadcast
		// (.255.255) — many tools refuse them as host addresses.
		if hostPrefix == 32 && (networkU32 == pool4Network || networkU32 == pool4Broadcast) {
			warnings = append(warnings, Warning{
				Host:   host,
				Reason: fmt.Sprintf("existing mgmt IP %s is the pool network or broadcast address, reassigning", subnet4),
			})
			continue
		}

		if claimant, exists := subnetClaim[networkU32]; exists {
			warnings = append(warnings, Warning{
				Host:   host,
				Reason: fmt.Sprintf("duplicate subnet %s (already claimed by %s), reassigning", subnet, claimant),
			})
			continue
		}

		subnetClaim[networkU32] = host
		usedNetworks[networkU32] = true
		result[host] = cloneIPNet(subnet)
	}

	// 3. Iterate the pool to assign new hosts.
	hostSubnetSize := 32 - hostPrefix
	step := uint32(1) << uint(hostSubnetSize)

	nextFreeSubnet := func() (*net.IPNet, error) {
		// For /32 allocations (mgmt IPs), skip both the pool network address
		// (.0) and the pool broadcast address (.255.255) since many tools
		// refuse them as host IPs. For larger subnets (e.g. /24), the bridge
		// inside the subnet handles its own .0/.broadcast — we only need to
		// not start the iterator at the broadcast itself.
		start := pool4Network
		end := pool4Broadcast
		if hostPrefix == 32 {
			start = pool4Network + 1
			// end stays at broadcast; loop is u < end so broadcast is excluded.
		}
		for u := start; u < end; u += step {
			if !usedNetworks[u] {
				mask := net.CIDRMask(hostPrefix, 32)
				return &net.IPNet{IP: uint32ToIP(u), Mask: mask}, nil
			}
		}
		return nil, fmt.Errorf("pool %s is exhausted (no free /%d subnets)", pool, hostPrefix)
	}

	for _, host := range hosts {
		if _, already := result[host]; already {
			continue
		}
		subnet, err := nextFreeSubnet()
		if err != nil {
			return nil, warnings, fmt.Errorf("allocating subnet for %s: %w", host, err)
		}
		usedNetworks[ipToUint32(subnet.IP.To4())] = true
		result[host] = subnet
	}

	return result, warnings, nil
}

// AllocateMgmtIPs assigns a /32 management IP to every host in hosts from pool.
// Wraps Allocate by promoting/demoting between net.IP and *net.IPNet.
func AllocateMgmtIPs(
	pool *net.IPNet,
	existing map[string]net.IP,
	hosts []string,
) (map[string]net.IP, []Warning, error) {
	wrapped := make(map[string]*net.IPNet, len(existing))
	for h, ip := range existing {
		ip4 := ip.To4()
		if ip4 == nil {
			continue
		}
		wrapped[h] = &net.IPNet{IP: ip4, Mask: net.CIDRMask(32, 32)}
	}

	subnets, warns, err := Allocate(pool, 32, wrapped, hosts)
	if err != nil {
		return nil, warns, err
	}

	out := make(map[string]net.IP, len(subnets))
	for h, n := range subnets {
		out[h] = cloneIP(n.IP.To4())
	}
	return out, warns, nil
}

// ipToUint32 converts a 4-byte IP to a uint32 for arithmetic.
func ipToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

// uint32ToIP converts a uint32 back to a net.IP.
func uint32ToIP(u uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, u)
	return ip
}

// cloneIP returns a copy of ip so that mutations don't affect the caller.
func cloneIP(ip net.IP) net.IP {
	c := make(net.IP, len(ip))
	copy(c, ip)
	return c
}

// cloneIPNet returns a deep copy of n.
func cloneIPNet(n *net.IPNet) *net.IPNet {
	return &net.IPNet{
		IP:   cloneIP(n.IP),
		Mask: append(net.IPMask(nil), n.Mask...),
	}
}

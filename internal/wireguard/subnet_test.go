package wireguard

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParseCIDR(s string) *net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return n
}

func TestMachineIP(t *testing.T) {
	tests := []struct {
		subnet string
		want   string
	}{
		{"10.210.0.0/24", "10.210.0.1"},
		{"10.210.5.0/24", "10.210.5.1"},
		{"10.210.255.0/24", "10.210.255.1"},
		{"192.168.0.0/24", "192.168.0.1"},
	}
	for _, tt := range tests {
		n := mustParseCIDR(tt.subnet)
		got := MachineIP(n)
		assert.Equal(t, tt.want, got.String(), "subnet=%s", tt.subnet)
	}
}

func TestAllocateMgmtIPs_Basic(t *testing.T) {
	pool := mustParseCIDR("100.64.0.0/16")
	hosts := []string{"h1", "h2", "h3"}

	got, warns, err := AllocateMgmtIPs(pool, nil, hosts)
	require.NoError(t, err)
	assert.Empty(t, warns)

	// Allocation skips pool network (.0.0) — starts at .0.1.
	assert.Equal(t, "100.64.0.1", got["h1"].String())
	assert.Equal(t, "100.64.0.2", got["h2"].String())
	assert.Equal(t, "100.64.0.3", got["h3"].String())
}

func TestAllocateMgmtIPs_StableReuse(t *testing.T) {
	pool := mustParseCIDR("100.64.0.0/16")
	existing := map[string]net.IP{
		"h1": net.ParseIP("100.64.0.42"),
	}
	hosts := []string{"h1", "h2"}

	got, warns, err := AllocateMgmtIPs(pool, existing, hosts)
	require.NoError(t, err)
	assert.Empty(t, warns)

	assert.Equal(t, "100.64.0.42", got["h1"].String())
	assert.Equal(t, "100.64.0.1", got["h2"].String())
}

func TestAllocateMgmtIPs_RejectsPoolNetworkAndBroadcast(t *testing.T) {
	pool := mustParseCIDR("100.64.0.0/16")
	existing := map[string]net.IP{
		"hN": net.ParseIP("100.64.0.0"),     // pool network
		"hB": net.ParseIP("100.64.255.255"), // pool broadcast
	}
	hosts := []string{"hN", "hB"}

	got, warns, err := AllocateMgmtIPs(pool, existing, hosts)
	require.NoError(t, err)
	assert.Len(t, warns, 2)

	for _, h := range hosts {
		ip := got[h].String()
		assert.NotEqual(t, "100.64.0.0", ip, h)
		assert.NotEqual(t, "100.64.255.255", ip, h)
	}
}

func TestAllocateMgmtIPs_OutOfPool_Warns(t *testing.T) {
	pool := mustParseCIDR("100.64.0.0/16")
	existing := map[string]net.IP{
		"h1": net.ParseIP("10.210.0.1"), // outside pool
	}
	hosts := []string{"h1"}

	got, warns, err := AllocateMgmtIPs(pool, existing, hosts)
	require.NoError(t, err)
	require.Len(t, warns, 1)
	assert.True(t, pool.Contains(got["h1"]), "reassigned IP must be inside pool")
}

func TestAllocate_PerHostSubnets(t *testing.T) {
	pool := mustParseCIDR("10.210.0.0/16")
	hosts := []string{"h1", "h2", "h3"}

	got, warns, err := Allocate(pool, 24, nil, hosts)
	require.NoError(t, err)
	assert.Empty(t, warns)

	assert.Equal(t, "10.210.0.0/24", got["h1"].String())
	assert.Equal(t, "10.210.1.0/24", got["h2"].String())
	assert.Equal(t, "10.210.2.0/24", got["h3"].String())
}

func TestAllocate_StableReuse(t *testing.T) {
	pool := mustParseCIDR("10.210.0.0/16")
	existing := map[string]*net.IPNet{
		"h1": mustParseCIDR("10.210.5.0/24"),
	}
	hosts := []string{"h1", "h2"}

	got, warns, err := Allocate(pool, 24, existing, hosts)
	require.NoError(t, err)
	assert.Empty(t, warns)

	// h1 keeps its existing subnet.
	assert.Equal(t, "10.210.5.0/24", got["h1"].String())
	// h2 gets the lowest free subnet (0 since 5 is taken).
	assert.Equal(t, "10.210.0.0/24", got["h2"].String())
}

func TestAllocate_FillsGaps(t *testing.T) {
	pool := mustParseCIDR("10.210.0.0/16")
	existing := map[string]*net.IPNet{
		"h1": mustParseCIDR("10.210.0.0/24"),
		"h2": mustParseCIDR("10.210.2.0/24"),
	}
	hosts := []string{"h1", "h2", "h3"}

	got, warns, err := Allocate(pool, 24, existing, hosts)
	require.NoError(t, err)
	assert.Empty(t, warns)

	assert.Equal(t, "10.210.0.0/24", got["h1"].String())
	assert.Equal(t, "10.210.2.0/24", got["h2"].String())
	// Gap at .1 is filled.
	assert.Equal(t, "10.210.1.0/24", got["h3"].String())
}

func TestAllocate_DuplicateSubnet_Warns(t *testing.T) {
	pool := mustParseCIDR("10.210.0.0/16")
	// Both ha and hb claim 10.210.5.0/24; ha wins (alphabetical).
	existing := map[string]*net.IPNet{
		"ha": mustParseCIDR("10.210.5.0/24"),
		"hb": mustParseCIDR("10.210.5.0/24"),
	}
	hosts := []string{"ha", "hb"}

	got, warns, err := Allocate(pool, 24, existing, hosts)
	require.NoError(t, err)
	require.Len(t, warns, 1)
	assert.Equal(t, "hb", warns[0].Host)
	assert.Contains(t, warns[0].Reason, "duplicate subnet")

	// ha keeps 10.210.5.0/24; hb is reassigned.
	assert.Equal(t, "10.210.5.0/24", got["ha"].String())
	assert.NotEqual(t, "10.210.5.0/24", got["hb"].String())
}

func TestAllocate_OutOfPool_Warns(t *testing.T) {
	pool := mustParseCIDR("10.210.0.0/16")
	existing := map[string]*net.IPNet{
		"h1": mustParseCIDR("192.168.0.0/24"), // outside pool
	}
	hosts := []string{"h1"}

	got, warns, err := Allocate(pool, 24, existing, hosts)
	require.NoError(t, err)
	require.Len(t, warns, 1)
	assert.Equal(t, "h1", warns[0].Host)
	assert.Contains(t, warns[0].Reason, "not a /24 inside pool")

	// h1 is reassigned to a pool address.
	assert.True(t, pool.Contains(got["h1"].IP), "reassigned IP must be inside pool")
}

func TestAllocate_WrongPrefix_Warns(t *testing.T) {
	pool := mustParseCIDR("10.210.0.0/16")
	existing := map[string]*net.IPNet{
		"h1": mustParseCIDR("10.210.0.0/16"), // wrong prefix (/16 instead of /24)
	}
	hosts := []string{"h1"}

	got, warns, err := Allocate(pool, 24, existing, hosts)
	require.NoError(t, err)
	require.Len(t, warns, 1)
	assert.Contains(t, warns[0].Reason, "not a /24 inside pool")

	ones, _ := got["h1"].Mask.Size()
	assert.Equal(t, 24, ones, "reassigned subnet must have /24 prefix")
}

func TestAllocate_DuplicateHost_Errors(t *testing.T) {
	pool := mustParseCIDR("10.210.0.0/16")
	hosts := []string{"1.1.1.1", "1.1.1.1"}

	_, _, err := Allocate(pool, 24, nil, hosts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate host")
}

func TestAllocate_PoolExhaustion(t *testing.T) {
	// /28 pool with /28 subnets — only one slot.
	pool := mustParseCIDR("10.0.0.0/28")
	hosts := []string{"h1", "h2"}

	_, _, err := Allocate(pool, 28, nil, hosts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exhausted")
}

func TestAllocate_EmptyHosts(t *testing.T) {
	pool := mustParseCIDR("10.210.0.0/16")
	got, warns, err := Allocate(pool, 24, nil, nil)
	require.NoError(t, err)
	assert.Empty(t, warns)
	assert.Empty(t, got)
}

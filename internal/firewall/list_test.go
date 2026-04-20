package firewall

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListAllow_Parses(t *testing.T) {
	r := &fakeRunner{responses: map[string]string{
		"iptables -S COOLIFY-ALLOW": `-N COOLIFY-ALLOW
-A COOLIFY-ALLOW -s 10.210.0.10/32 -d 10.210.1.10/32 -p tcp -m tcp --dport 80 -m comment --comment "cid:abc123def456" -j ACCEPT
-A COOLIFY-ALLOW -s 10.210.0.11 -d 10.210.1.11 -p udp -m udp --dport 53 -j ACCEPT
`,
	}}
	got, err := ListAllow(context.Background(), r, "h1", "root", 22)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "h1", got[0].Host)
	assert.Equal(t, "tcp", got[0].Proto)
	assert.Equal(t, 80, got[0].Port)
	assert.Equal(t, "udp", got[1].Proto)
	assert.Equal(t, 53, got[1].Port)
}

func TestListAllow_EmptyChainMissing(t *testing.T) {
	r := &fakeRunner{responses: map[string]string{}}
	got, err := ListAllow(context.Background(), r, "h1", "root", 22)
	assert.NoError(t, err)
	assert.Empty(t, got)
}

func TestListAll_Sorted(t *testing.T) {
	r := &fakeRunner{responses: map[string]string{
		"iptables -S COOLIFY-ALLOW": `-A COOLIFY-ALLOW -s 10.0.0.1 -d 10.0.0.2 -p tcp -m tcp --dport 80 -j ACCEPT
`,
	}}
	all, perHost := ListAll(context.Background(), r,
		[]string{"h2", "h1"}, "root", 22, 2)
	assert.Len(t, all, 2)
	assert.Equal(t, "h1", all[0].Host)
	assert.Equal(t, "h2", all[1].Host)
	assert.Len(t, perHost, 2)
}

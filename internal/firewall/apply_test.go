package firewall

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mkRule() AllowRule {
	return AllowRule{
		Host:  "h1",
		Src:   net.ParseIP("10.210.0.10"),
		Dst:   net.ParseIP("10.210.1.10"),
		Proto: "tcp", Port: 80,
		Comment: "cid:abc123def456",
	}
}

func TestBuildApplyCmd_Shape(t *testing.T) {
	cmd := buildApplyCmd(mkRule())
	// Ordering: ensure chain → persistence unit → -C||-A → save.
	ensureIdx := indexOf(cmd, "iptables -N COOLIFY-ALLOW")
	unitIdx := indexOf(cmd, "systemctl enable "+PersistUnitName)
	checkIdx := indexOf(cmd, "iptables -C COOLIFY-ALLOW")
	appendIdx := indexOf(cmd, "iptables -A COOLIFY-ALLOW")
	saveIdx := indexOf(cmd, "iptables -S "+ChainName)
	assert.True(t, ensureIdx >= 0)
	assert.True(t, unitIdx > ensureIdx)
	assert.True(t, checkIdx > unitIdx)
	assert.True(t, appendIdx > checkIdx)
	assert.True(t, saveIdx > appendIdx)
}

func TestBuildRevokeCmd_Shape(t *testing.T) {
	cmd := buildRevokeCmd(mkRule())
	assert.Contains(t, cmd, "iptables -C COOLIFY-ALLOW")
	assert.Contains(t, cmd, "iptables -D COOLIFY-ALLOW")
	assert.Contains(t, cmd, "exit 0") // no-op when chain missing
	assert.Contains(t, cmd, SaveRulesCommand())
}

func TestApplyAllow_MissingHost(t *testing.T) {
	r := mkRule()
	r.Host = ""
	err := ApplyAllow(context.Background(), &fakeRunner{}, r, "root", 22)
	assert.Error(t, err)
}

func TestApplyAllow_RunsAndCaptures(t *testing.T) {
	fr := &fakeRunner{responses: map[string]string{}}
	err := ApplyAllow(context.Background(), fr, mkRule(), "root", 22)
	assert.NoError(t, err)
	assert.Len(t, fr.calls, 1)
	assert.Contains(t, fr.calls[0], "iptables -A COOLIFY-ALLOW")
}

func TestRevokeAllow_RunsAndCaptures(t *testing.T) {
	fr := &fakeRunner{responses: map[string]string{}}
	err := RevokeAllow(context.Background(), fr, mkRule(), "root", 22)
	assert.NoError(t, err)
	assert.Len(t, fr.calls, 1)
	assert.Contains(t, fr.calls[0], "iptables -D COOLIFY-ALLOW")
}

package firewall

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeID_Stable(t *testing.T) {
	a := ComputeID(net.ParseIP("10.210.0.10"), net.ParseIP("10.210.1.10"), "tcp", 80)
	b := ComputeID(net.ParseIP("10.210.0.10"), net.ParseIP("10.210.1.10"), "tcp", 80)
	assert.Equal(t, a, b)
	assert.Len(t, a, 12)
}

func TestComputeID_CaseInsensitiveProto(t *testing.T) {
	a := ComputeID(net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "TCP", 80)
	b := ComputeID(net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "tcp", 80)
	assert.Equal(t, a, b)
}

func TestComputeID_DifferentInputsDifferent(t *testing.T) {
	a := ComputeID(net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "tcp", 80)
	b := ComputeID(net.ParseIP("1.1.1.1"), net.ParseIP("2.2.2.2"), "tcp", 443)
	assert.NotEqual(t, a, b)
}

func TestAllowRule_RenderAppend(t *testing.T) {
	tests := []struct {
		name string
		rule AllowRule
		want string
	}{
		{
			name: "tcp with port and comment",
			rule: AllowRule{
				Src: net.ParseIP("10.210.0.10"), Dst: net.ParseIP("10.210.1.10"),
				Proto: "tcp", Port: 80, Comment: "cid:abc123def456",
			},
			want: `iptables -A COOLIFY-ALLOW -s 10.210.0.10 -d 10.210.1.10 -p tcp --dport 80 -m comment --comment "cid:abc123def456" -j ACCEPT`,
		},
		{
			name: "udp with port",
			rule: AllowRule{
				Src: net.ParseIP("10.210.0.10"), Dst: net.ParseIP("10.210.1.10"),
				Proto: "udp", Port: 53,
			},
			want: `iptables -A COOLIFY-ALLOW -s 10.210.0.10 -d 10.210.1.10 -p udp --dport 53 -j ACCEPT`,
		},
		{
			name: "no proto no port",
			rule: AllowRule{
				Src: net.ParseIP("10.210.0.10"), Dst: net.ParseIP("10.210.1.10"),
			},
			want: `iptables -A COOLIFY-ALLOW -s 10.210.0.10 -d 10.210.1.10 -j ACCEPT`,
		},
		{
			name: "proto without port",
			rule: AllowRule{
				Src: net.ParseIP("10.210.0.10"), Dst: net.ParseIP("10.210.1.10"),
				Proto: "tcp",
			},
			want: `iptables -A COOLIFY-ALLOW -s 10.210.0.10 -d 10.210.1.10 -p tcp -j ACCEPT`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.rule.RenderAppend())
		})
	}
}

func TestAllowRule_RenderDelete_And_Check(t *testing.T) {
	r := AllowRule{
		Src: net.ParseIP("10.0.0.1"), Dst: net.ParseIP("10.0.0.2"),
		Proto: "tcp", Port: 8080, Comment: "cid:xyz000000001",
	}
	assert.Contains(t, r.RenderDelete(), "iptables -D COOLIFY-ALLOW")
	assert.Contains(t, r.RenderCheck(), "iptables -C COOLIFY-ALLOW")
	// Match args identical between -D and -A.
	_, after, _ := splitAtFirst(r.RenderAppend(), "COOLIFY-ALLOW ")
	_, afterDel, _ := splitAtFirst(r.RenderDelete(), "COOLIFY-ALLOW ")
	assert.Equal(t, after, afterDel)
}

func splitAtFirst(s, sep string) (before, after string, ok bool) {
	i := indexOf(s, sep)
	if i < 0 {
		return s, "", false
	}
	return s[:i], s[i+len(sep):], true
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestParseChainLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantOk  bool
		wantSrc string
		wantDst string
		wantPro string
		wantPrt int
		wantCmt string
	}{
		{
			name:    "full rule with comment",
			line:    `-A COOLIFY-ALLOW -s 10.210.0.10/32 -d 10.210.1.10/32 -p tcp -m tcp --dport 80 -m comment --comment "cid:abc123def456" -j ACCEPT`,
			wantOk:  true,
			wantSrc: "10.210.0.10",
			wantDst: "10.210.1.10",
			wantPro: "tcp",
			wantPrt: 80,
			wantCmt: "cid:abc123def456",
		},
		{
			name:    "no port, unquoted comment",
			line:    `-A COOLIFY-ALLOW -s 10.0.0.1 -d 10.0.0.2 -m comment --comment cid:simple123456 -j ACCEPT`,
			wantOk:  true,
			wantSrc: "10.0.0.1",
			wantDst: "10.0.0.2",
			wantPro: "",
			wantPrt: 0,
			wantCmt: "cid:simple123456",
		},
		{
			name:    "udp with port",
			line:    `-A COOLIFY-ALLOW -s 10.0.0.1 -d 10.0.0.2 -p udp -m udp --dport 53 -j ACCEPT`,
			wantOk:  true,
			wantSrc: "10.0.0.1",
			wantDst: "10.0.0.2",
			wantPro: "udp",
			wantPrt: 53,
		},
		{
			name:   "other chain",
			line:   `-A FORWARD -j DROP`,
			wantOk: false,
		},
		{
			name:   "empty",
			line:   ``,
			wantOk: false,
		},
		{
			name:   "missing src",
			line:   `-A COOLIFY-ALLOW -d 10.0.0.2 -j ACCEPT`,
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, ok := ParseChainLine(tt.line)
			assert.Equal(t, tt.wantOk, ok)
			if !tt.wantOk {
				return
			}
			assert.Equal(t, tt.wantSrc, r.Src.String())
			assert.Equal(t, tt.wantDst, r.Dst.String())
			assert.Equal(t, tt.wantPro, r.Proto)
			assert.Equal(t, tt.wantPrt, r.Port)
			if tt.wantCmt != "" {
				assert.Equal(t, tt.wantCmt, r.Comment)
			}
		})
	}
}

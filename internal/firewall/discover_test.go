package firewall

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDiscoverLine(t *testing.T) {
	tests := []struct {
		line   string
		wantOk bool
		wantID string
		wantNm string
		wantIP string
	}{
		{"abcdef123456|web|10.210.0.10", true, "abcdef123456", "web", "10.210.0.10"},
		{"abcdef1234567890|web|10.210.0.10", true, "abcdef123456", "web", "10.210.0.10"},
		{"|name|10.0.0.1", false, "", "", ""},
		{"id|name|", false, "", "", ""},
		{"id|name|not-an-ip", false, "", "", ""},
		{"", false, "", "", ""},
		{"a|b", false, "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			id, name, ip, ok := ParseDiscoverLine(tt.line)
			assert.Equal(t, tt.wantOk, ok)
			if !ok {
				return
			}
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantNm, name)
			assert.Equal(t, tt.wantIP, ip.String())
		})
	}
}

// fakeRunner is a deterministic ssh.Runner for firewall tests. Responses
// map a command substring to its canned stdout.
type fakeRunner struct {
	responses map[string]string
	calls     []string
}

func (f *fakeRunner) Run(_ context.Context, _, _ string, _ int, cmd string) (string, string, error) {
	f.calls = append(f.calls, cmd)
	for sub, resp := range f.responses {
		if strings.Contains(cmd, sub) {
			return resp, "", nil
		}
	}
	return "", "", nil
}

func TestDiscoverContainers(t *testing.T) {
	r := &fakeRunner{responses: map[string]string{
		"podman ps": "abc111111111|web|10.210.0.10\ndef222222222|api|10.210.0.11\n\n",
	}}
	got, err := DiscoverContainers(context.Background(), r, "h1", "root", 22, "coolify-mesh")
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "api", got[0].Name) // sorted by name
	assert.Equal(t, "web", got[1].Name)
	assert.Equal(t, "h1", got[0].Host)
	assert.Equal(t, "10.210.0.11", got[0].IP.String())
}

func TestDiscoverContainers_EmptyOutput(t *testing.T) {
	r := &fakeRunner{responses: map[string]string{}}
	got, err := DiscoverContainers(context.Background(), r, "h1", "root", 22, "coolify-mesh")
	assert.NoError(t, err)
	assert.Empty(t, got)
}

func TestDiscoverContainers_BadLinesSkipped(t *testing.T) {
	r := &fakeRunner{responses: map[string]string{
		"podman ps": "abc111111111|web|10.210.0.10\ngarbage\n|noid|1.1.1.1\n",
	}}
	got, err := DiscoverContainers(context.Background(), r, "h1", "root", 22, "coolify-mesh")
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "web", got[0].Name)
}

func TestDiscoverAll_Sorted(t *testing.T) {
	r := &fakeRunner{responses: map[string]string{
		"podman ps": "aaa111111111|x|10.210.0.10",
	}}
	all, perHost := DiscoverAll(context.Background(), r,
		[]string{"h2", "h1"}, "root", 22, "coolify-mesh", 2)
	assert.Len(t, all, 2)
	assert.Equal(t, "h1", all[0].Host)
	assert.Equal(t, "h2", all[1].Host)
	assert.Len(t, perHost, 2)
}

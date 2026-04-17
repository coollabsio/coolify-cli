package services

import (
	"net"
	"strings"
	"testing"
)

func TestCooldServiceUnit_EmbedsMgmtIP(t *testing.T) {
	got := CooldServiceUnit(net.ParseIP("100.64.0.5"))

	for _, want := range []string{
		"Environment=COOLD_HOST_MGMT_IP=100.64.0.5",
		"Wants=corrosion.service",
		"After=corrosion.service network-online.target podman.socket",
		"ExecStart=/usr/local/bin/coold",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("unit missing %q:\n%s", want, got)
		}
	}
}

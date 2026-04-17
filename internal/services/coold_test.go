package services

import (
	"net"
	"strings"
	"testing"
)

func TestCooldServiceUnit_EmbedsMgmtIP(t *testing.T) {
	got := CooldServiceUnit(net.ParseIP("100.64.0.5"), net.ParseIP("10.210.7.1"))

	for _, want := range []string{
		"Environment=COOLD_HOST_MGMT_IP=100.64.0.5",
		"Environment=COOLD_BRIDGE_GATEWAY_IP=10.210.7.1",
		"Environment=COOLD_DNS_ZONE=coolify.internal",
		"Wants=corrosion.service",
		"After=corrosion.service network-online.target podman.socket",
		"ExecStart=/usr/local/bin/coold",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("unit missing %q:\n%s", want, got)
		}
	}
}

func TestCooldServiceUnit_NilBridgeGatewaySkipsDNSEnv(t *testing.T) {
	got := CooldServiceUnit(net.ParseIP("100.64.0.5"), nil)

	if strings.Contains(got, "COOLD_BRIDGE_GATEWAY_IP") {
		t.Errorf("expected no bridge env when nil, got:\n%s", got)
	}
	if strings.Contains(got, "COOLD_DNS_ZONE") {
		t.Errorf("expected no DNS zone env when nil, got:\n%s", got)
	}
	if !strings.Contains(got, "Environment=COOLD_HOST_MGMT_IP=100.64.0.5") {
		t.Errorf("expected mgmt IP env, got:\n%s", got)
	}
}

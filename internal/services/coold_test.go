package services

import (
	"net"
	"strings"
	"testing"
)

func TestCooldServiceUnit_EmbedsMgmtIPAndNamespaces(t *testing.T) {
	namespaces := []CooldNamespace{
		{Name: "default", Network: "coolify-default-mesh", BridgeGateway: net.ParseIP("10.210.7.1")},
		{Name: "alpha", Network: "coolify-alpha-mesh", BridgeGateway: net.ParseIP("10.210.8.1")},
	}
	got := CooldServiceUnit(net.ParseIP("100.64.0.5"), namespaces)

	for _, want := range []string{
		"Environment=COOLD_HOST_MGMT_IP=100.64.0.5",
		"Environment=COOLD_NAMESPACES=default:coolify-default-mesh:10.210.7.1,alpha:coolify-alpha-mesh:10.210.8.1",
		"Environment=COOLD_DNS_ZONE=coolify.internal",
		"Environment=COOLD_API_BIND=100.64.0.5:8443",
		"Environment=COOLD_API_TOKEN_FILE=/etc/coolify/api-token",
		"AmbientCapabilities=CAP_NET_BIND_SERVICE CAP_NET_ADMIN CAP_NET_RAW",
		"Wants=corrosion.service",
		"After=corrosion.service network-online.target podman.socket",
		"ExecStart=/usr/local/bin/coold",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("unit missing %q:\n%s", want, got)
		}
	}
}

func TestCooldServiceUnit_EmptyNamespacesSkipsNamespaceEnv(t *testing.T) {
	got := CooldServiceUnit(net.ParseIP("100.64.0.5"), nil)

	if strings.Contains(got, "COOLD_NAMESPACES") {
		t.Errorf("expected no namespace env when nil, got:\n%s", got)
	}
	if strings.Contains(got, "COOLD_DNS_ZONE") {
		t.Errorf("expected no DNS zone env when nil, got:\n%s", got)
	}
	if !strings.Contains(got, "Environment=COOLD_HOST_MGMT_IP=100.64.0.5") {
		t.Errorf("expected mgmt IP env, got:\n%s", got)
	}
}

func TestCooldNamespacesEnvValue_Triples(t *testing.T) {
	ns := []CooldNamespace{
		{Name: "default", Network: "coolify-default-mesh", BridgeGateway: net.ParseIP("10.210.0.1")},
		{Name: "alpha", Network: "coolify-alpha-mesh", BridgeGateway: net.ParseIP("10.220.0.1")},
	}
	got := CooldNamespacesEnvValue(ns)
	want := "default:coolify-default-mesh:10.210.0.1,alpha:coolify-alpha-mesh:10.220.0.1"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if CooldNamespacesEnvValue(nil) != "" {
		t.Errorf("expected empty string for nil slice")
	}
}

package services

import (
	"net"
	"strings"
	"testing"
)

func TestCooldInstallCommand_SubstitutesVersion(t *testing.T) {
	for _, version := range []string{"nightly", "v1.2.3"} {
		cmd := CooldInstallCommand(version)
		if !strings.Contains(cmd, version) {
			t.Errorf("version %q not found in install command", version)
		}
		if !strings.Contains(cmd, "coollabsio/coold/releases/download/"+version) {
			t.Errorf("release URL missing version %q in:\n%s", version, cmd)
		}
		if !strings.Contains(cmd, "/usr/local/bin/coold.version") {
			t.Errorf("version marker write missing from install command")
		}
	}
}

func TestCooldInstallCommand_ArchDetection(t *testing.T) {
	cmd := CooldInstallCommand("nightly")
	for _, want := range []string{
		"x86_64)  ARCH=amd64",
		"aarch64) ARCH=arm64",
		"coold-linux-${ARCH}.tar.gz",
		"install -m 0755",
	} {
		if !strings.Contains(cmd, want) {
			t.Errorf("expected %q in install command:\n%s", want, cmd)
		}
	}
}

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

func TestCooldServiceUnit_EmitsBuilderEnvWhenConfigured(t *testing.T) {
	builder := &BuilderConfig{
		Capacity:    4,
		CPUQuota:    "400%",
		MemoryMax:   "4G",
		TimeoutSecs: 900,
		DenyNets:    []string{"100.64.0.0/16", "10.210.0.0/16"},
	}
	got := CooldServiceUnitWithScheduler(
		net.ParseIP("100.64.0.5"),
		nil,
		&SchedulerConfig{URL: "http://100.64.0.1:6443", JWTPath: "/etc/coolify/host-jwt"},
		builder,
	)

	for _, want := range []string{
		"Environment=COOLD_BUILDER_ENABLED=true",
		"Environment=COOLD_BUILDER_CAPACITY=4",
		"Environment=COOLD_BUILDER_CPU_QUOTA=400%",
		"Environment=COOLD_BUILDER_MEMORY_MAX=4G",
		"Environment=COOLD_BUILDER_TIMEOUT_SECS=900",
		"Environment=COOLD_BUILDER_DENY_NETS=100.64.0.0/16,10.210.0.0/16",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("unit missing %q:\n%s", want, got)
		}
	}
}

func TestCooldServiceUnit_BuilderDefaultsWhenZero(t *testing.T) {
	builder := &BuilderConfig{} // all zero values
	got := CooldServiceUnitWithScheduler(
		net.ParseIP("100.64.0.5"),
		nil,
		&SchedulerConfig{URL: "http://100.64.0.1:6443", JWTPath: "/etc/coolify/host-jwt"},
		builder,
	)

	for _, want := range []string{
		"Environment=COOLD_BUILDER_CAPACITY=2",
		"Environment=COOLD_BUILDER_CPU_QUOTA=200%",
		"Environment=COOLD_BUILDER_MEMORY_MAX=2G",
		"Environment=COOLD_BUILDER_TIMEOUT_SECS=1800",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("unit missing default %q:\n%s", want, got)
		}
	}
}

func TestCooldServiceUnit_OmitsBuilderEnvWhenNil(t *testing.T) {
	got := CooldServiceUnitWithScheduler(
		net.ParseIP("100.64.0.5"),
		nil,
		&SchedulerConfig{URL: "http://100.64.0.1:6443", JWTPath: "/etc/coolify/host-jwt"},
		nil,
	)
	if strings.Contains(got, "COOLD_BUILDER_") {
		t.Errorf("expected no builder env when nil, got:\n%s", got)
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

package services

import (
	"strings"
	"testing"
)

func TestBrokerInstallCommand_ContainsNewAssetName(t *testing.T) {
	cmd := BrokerInstallCommand("nightly")

	for _, want := range []string{
		"broker-linux-${ARCH}.tar.gz",
		"/usr/local/bin/broker",
		"nightly",
	} {
		if !strings.Contains(cmd, want) {
			t.Errorf("BrokerInstallCommand missing %q", want)
		}
	}
	if strings.Contains(cmd, "coolify-broker") {
		t.Error("BrokerInstallCommand still contains old name 'coolify-broker'")
	}
}

func TestBrokerInstallCommand_VersionTagEmbedded(t *testing.T) {
	cmd := BrokerInstallCommand("v1.2.3")
	if !strings.Contains(cmd, "v1.2.3") {
		t.Error("BrokerInstallCommand missing version tag in URL and version file write")
	}
}

func TestBrokerServiceUnit_ExecStartPath(t *testing.T) {
	unit := BrokerServiceUnit("100.64.0.1:6443", BrokerJWTPubPath)

	if !strings.Contains(unit, "ExecStart=/usr/local/bin/broker") {
		t.Error("BrokerServiceUnit ExecStart does not point to /usr/local/bin/broker")
	}
	if strings.Contains(unit, "coolify-broker") {
		t.Error("BrokerServiceUnit still contains old name 'coolify-broker'")
	}
	if strings.Contains(unit, "BUILDER_GRPC_BIND") {
		t.Error("BrokerServiceUnit still emits BROKER_BUILDER_GRPC_BIND; builder port was removed")
	}
	if strings.Contains(unit, "BROKER_REDIS_URL") || strings.Contains(unit, "redis") {
		t.Error("BrokerServiceUnit still references Redis; UDS migration should have dropped it")
	}
	for _, want := range []string{
		"BROKER_GRPC_BIND=100.64.0.1:6443",
		"BROKER_UNIX_SOCKET_PATH=" + BrokerUnixSocketPath,
		"RuntimeDirectory=coolify",
		BrokerJWTPubPath,
	} {
		if !strings.Contains(unit, want) {
			t.Errorf("BrokerServiceUnit missing %q", want)
		}
	}
}

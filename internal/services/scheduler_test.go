package services

import (
	"strings"
	"testing"
)

func TestSchedulerInstallCommand_ContainsNewAssetName(t *testing.T) {
	cmd := SchedulerInstallCommand("nightly")

	for _, want := range []string{
		"scheduler-linux-${ARCH}.tar.gz",
		"/usr/local/bin/scheduler",
		"nightly",
	} {
		if !strings.Contains(cmd, want) {
			t.Errorf("SchedulerInstallCommand missing %q", want)
		}
	}
	if strings.Contains(cmd, "coolify-scheduler") {
		t.Error("SchedulerInstallCommand still contains old name 'coolify-scheduler'")
	}
}

func TestSchedulerInstallCommand_VersionTagEmbedded(t *testing.T) {
	cmd := SchedulerInstallCommand("v1.2.3")
	if !strings.Contains(cmd, "v1.2.3") {
		t.Error("SchedulerInstallCommand missing version tag in URL and version file write")
	}
}

func TestSchedulerServiceUnit_ExecStartPath(t *testing.T) {
	unit := SchedulerServiceUnit("100.64.0.1:6443", SchedulerJWTPubPath)

	if !strings.Contains(unit, "ExecStart=/usr/local/bin/scheduler") {
		t.Error("SchedulerServiceUnit ExecStart does not point to /usr/local/bin/scheduler")
	}
	if strings.Contains(unit, "coolify-scheduler") {
		t.Error("SchedulerServiceUnit still contains old name 'coolify-scheduler'")
	}
	if strings.Contains(unit, "BUILDER_GRPC_BIND") {
		t.Error("SchedulerServiceUnit still emits SCHEDULER_BUILDER_GRPC_BIND; builder port was removed")
	}
	if strings.Contains(unit, "SCHEDULER_REDIS_URL") || strings.Contains(unit, "redis") {
		t.Error("SchedulerServiceUnit still references Redis; UDS migration should have dropped it")
	}
	for _, want := range []string{
		"SCHEDULER_GRPC_BIND=100.64.0.1:6443",
		"SCHEDULER_UNIX_SOCKET_PATH=" + SchedulerUnixSocketPath,
		"RuntimeDirectory=coolify",
		SchedulerJWTPubPath,
	} {
		if !strings.Contains(unit, want) {
			t.Errorf("SchedulerServiceUnit missing %q", want)
		}
	}
}

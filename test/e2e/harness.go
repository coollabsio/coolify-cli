//go:build e2e

// Package e2e is a live-server test harness for the coold/broker/builder
// stack. It assumes a mesh already deployed by
//
//	coolify init apply --servers $BUILDER_HOST,$COOLD_ONLY_HOST \
//	                    --central $CENTRAL_HOST \
//	                    --namespaces default \
//	                    --builder-hosts $BUILDER_HOST \
//	                    --yes
//
// and talks to Redis on the central host via ssh + `redis-cli`. No mocks.
//
// Run with:
//
//	BUILDER_HOST=78.47.80.33 \
//	COOLD_ONLY_HOST=159.69.186.231 \
//	BUILDER_MGMT=100.64.0.1 \
//	COOLD_ONLY_MGMT=100.64.0.2 \
//	CENTRAL_HOST=78.47.80.33 \
//	SSH_KEY=~/.ssh/id_ed25519-no-pass \
//	go test -tags e2e -v -timeout 15m ./test/e2e/...
package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type env struct {
	BuilderHost   string // SSH address of the host with builder capability
	CooldOnlyHost string // SSH address of the host without builder capability
	BuilderMgmt   string // wg0 mgmt IP for BuilderHost (host_id in envelopes)
	CooldOnlyMgmt string // wg0 mgmt IP for CooldOnlyHost
	CentralHost   string // SSH address of the central host (Redis + broker)
	SSHKey        string // private key path
	SSHUser       string // ssh user (default "root")
}

func load(t *testing.T) env {
	t.Helper()
	e := env{
		BuilderHost:   must(t, "BUILDER_HOST"),
		CooldOnlyHost: must(t, "COOLD_ONLY_HOST"),
		BuilderMgmt:   must(t, "BUILDER_MGMT"),
		CooldOnlyMgmt: must(t, "COOLD_ONLY_MGMT"),
		CentralHost:   must(t, "CENTRAL_HOST"),
		SSHKey:        must(t, "SSH_KEY"),
		SSHUser:       getenv("SSH_USER", "root"),
	}
	return e
}

func must(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Fatalf("env %s required", key)
	}
	return v
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// sshRun executes cmd on host. Returns stdout + err. stderr folded into err.
func (e env) sshRun(host, cmd string) (string, error) {
	args := []string{
		"-i", e.SSHKey,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
		fmt.Sprintf("%s@%s", e.SSHUser, host),
		cmd,
	}
	out, err := exec.Command("ssh", args...).CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("ssh %s: %w: %s", host, err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// redisXadd writes a build:cmd envelope on central's Redis.
func (e env) redisXadd(payload string) error {
	cmd := fmt.Sprintf("redis-cli XADD build:cmd '*' payload %q", payload)
	_, err := e.sshRun(e.CentralHost, cmd)
	return err
}

// redisLpop reads one entry from build:resp:<requestId>, or "" if empty.
func (e env) redisLpop(requestId string) (string, error) {
	out, err := e.sshRun(e.CentralHost, fmt.Sprintf("redis-cli LPOP build:resp:%s", requestId))
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	if out == "(nil)" || out == "" {
		return "", nil
	}
	return out, nil
}

type buildResponse struct {
	RequestID   string `json:"request_id"`
	Status      string `json:"status"`
	Digest      string `json:"digest,omitempty"`
	RegistryRef string `json:"registry_ref,omitempty"`
	DurationMs  uint64 `json:"duration_ms,omitempty"`
	Code        uint32 `json:"code,omitempty"`
	Message     string `json:"message,omitempty"`
	Stage       string `json:"stage,omitempty"`
}

// waitBuildResp polls build:resp:<requestId> until a value appears or timeout.
func (e env) waitBuildResp(t *testing.T, requestId string, timeout time.Duration) buildResponse {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		line, err := e.redisLpop(requestId)
		if err != nil {
			t.Fatalf("LPOP build:resp:%s failed: %v", requestId, err)
		}
		if line != "" {
			var r buildResponse
			if err := json.Unmarshal([]byte(line), &r); err != nil {
				t.Fatalf("parse response %q: %v", line, err)
			}
			return r
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("no build:resp:%s within %s", requestId, timeout)
	return buildResponse{}
}

// hasImage reports whether `buildah images` on host lists any row whose
// name contains tag.
func (e env) hasImage(host, tag string) bool {
	out, _ := e.sshRun(host, fmt.Sprintf("buildah images 2>/dev/null | grep -q %q && echo Y", tag))
	return strings.Contains(out, "Y")
}

// unitActive reports whether the transient build service for requestId is
// active on host.
func (e env) unitActive(host, requestId string) bool {
	out, _ := e.sshRun(host, fmt.Sprintf("systemctl is-active coolify-build-%s.service 2>&1", requestId))
	return strings.TrimSpace(out) == "active"
}

// restartCoold blocks until `systemctl restart coold` returns on host.
func (e env) restartCoold(host string) error {
	_, err := e.sshRun(host, "systemctl restart coold")
	return err
}

// cleanImage removes a leftover image tag so repeated tests don't collide.
func (e env) cleanImage(host, tag string) {
	_, _ = e.sshRun(host, fmt.Sprintf("buildah rmi -f %q 2>/dev/null || true", "localhost/"+tag))
}

// uniqReqID returns a lowercase request_id suitable for use as an OCI image
// tag (OCI rejects uppercase).
func uniqReqID(prefix string) string {
	return fmt.Sprintf("%s-%d", strings.ToLower(prefix), time.Now().UnixNano())
}

// buildEnvelope assembles the JSON payload Laravel would send.
func buildEnvelope(requestId, hostId, repoURL, gitRef, target, outputDir string) string {
	body := map[string]interface{}{
		"request_id": requestId,
		"command": map[string]interface{}{
			"type":         "static_build",
			"repo_url":     repoURL,
			"git_ref":      gitRef,
			"target_image": target,
			"output_dir":   outputDir,
		},
	}
	if hostId != "" {
		body["host_id"] = hostId
	}
	b, _ := json.Marshal(body)
	return string(b)
}

func cancelEnvelope(requestId string) string {
	b, _ := json.Marshal(map[string]interface{}{
		"request_id": requestId,
		"command":    map[string]interface{}{"type": "cancel"},
	})
	return string(b)
}

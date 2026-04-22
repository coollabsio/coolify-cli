//go:build e2e

package e2e

import (
	"strings"
	"testing"
	"time"
)

// Small static-site repo used as the happy-path fixture. It has HTML at
// the repo root so we point output_dir="." and skip the default "dist".
const smallRepo = "https://github.com/mdn/beginner-html-site"

// Large repo used to keep a build "in flight" for cancel / restart tests.
// git clone --depth=1 is still slow enough to give us multi-second windows.
const slowRepo = "https://github.com/torvalds/linux"

func TestPinToBuilderHost(t *testing.T) {
	e := load(t)
	req := uniqReqID("e2e-pin-builder")
	defer e.cleanImage(e.BuilderHost, req)

	payload := buildEnvelope(req, e.BuilderMgmt, smallRepo, "main", "localhost/"+req, ".")
	if err := e.redisXadd(payload); err != nil {
		t.Fatalf("XADD: %v", err)
	}

	resp := e.waitBuildResp(t, req, 3*time.Minute)
	if resp.Status != "ok" {
		t.Fatalf("want ok, got %+v", resp)
	}
	if !strings.HasPrefix(resp.Digest, "sha256:") {
		t.Fatalf("expected sha256 digest, got %q", resp.Digest)
	}
	if !e.hasImage(e.BuilderHost, req) {
		t.Fatalf("image %s missing on builder host", req)
	}
	if e.hasImage(e.CooldOnlyHost, req) {
		t.Fatalf("image %s appeared on coold-only host — pinning leaked", req)
	}
}

func TestPinToCooldOnlyHostReturns503(t *testing.T) {
	e := load(t)
	req := uniqReqID("e2e-pin-coold-only")

	payload := buildEnvelope(req, e.CooldOnlyMgmt, smallRepo, "main", "localhost/"+req, ".")
	if err := e.redisXadd(payload); err != nil {
		t.Fatalf("XADD: %v", err)
	}

	resp := e.waitBuildResp(t, req, 30*time.Second)
	if resp.Status != "error" {
		t.Fatalf("want error, got %+v", resp)
	}
	if resp.Code != 503 {
		t.Fatalf("want code 503, got %d (%s)", resp.Code, resp.Message)
	}
	if !strings.Contains(resp.Message, "host has no builder capability") {
		t.Fatalf("want cap-missing message, got %q", resp.Message)
	}
}

func TestUnknownHostIdReturns503(t *testing.T) {
	e := load(t)
	req := uniqReqID("e2e-unknown-host")

	payload := buildEnvelope(req, "100.64.99.99", smallRepo, "main", "localhost/"+req, ".")
	if err := e.redisXadd(payload); err != nil {
		t.Fatalf("XADD: %v", err)
	}

	resp := e.waitBuildResp(t, req, 30*time.Second)
	if resp.Status != "error" || resp.Code != 503 {
		t.Fatalf("want error 503, got %+v", resp)
	}
}

func TestLoadBalancePicksBuilderHost(t *testing.T) {
	e := load(t)
	req := uniqReqID("e2e-lb")
	defer e.cleanImage(e.BuilderHost, req)

	payload := buildEnvelope(req, "" /*no host_id*/, smallRepo, "main", "localhost/"+req, ".")
	if err := e.redisXadd(payload); err != nil {
		t.Fatalf("XADD: %v", err)
	}

	resp := e.waitBuildResp(t, req, 3*time.Minute)
	if resp.Status != "ok" {
		t.Fatalf("want ok, got %+v", resp)
	}
	// Only builder host has the capability, so it MUST end up there.
	if !e.hasImage(e.BuilderHost, req) {
		t.Fatalf("image missing on builder host; load-balance should have picked it")
	}
}

func TestBuildCancelEmitsStageCancel(t *testing.T) {
	e := load(t)
	req := uniqReqID("e2e-cancel")

	// Slow repo so the build is still mid-clone when we cancel.
	payload := buildEnvelope(req, e.BuilderMgmt, slowRepo, "master", "localhost/"+req, ".")
	if err := e.redisXadd(payload); err != nil {
		t.Fatalf("XADD: %v", err)
	}

	// Wait for the transient unit to appear.
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) && !e.unitActive(e.BuilderHost, req) {
		time.Sleep(1 * time.Second)
	}
	if !e.unitActive(e.BuilderHost, req) {
		t.Fatalf("transient unit never activated")
	}

	// Cancel.
	if err := e.redisXadd(cancelEnvelope(req)); err != nil {
		t.Fatalf("cancel XADD: %v", err)
	}

	resp := e.waitBuildResp(t, req, 30*time.Second)
	if resp.Status != "error" || resp.Code != 499 || resp.Stage != "cancel" {
		t.Fatalf("want error/499/cancel, got %+v", resp)
	}

	// Unit should be gone.
	time.Sleep(2 * time.Second)
	if e.unitActive(e.BuilderHost, req) {
		t.Fatalf("unit still active after cancel")
	}
}

func TestCooldRestartAdoptsInFlightBuild(t *testing.T) {
	e := load(t)
	req := uniqReqID("e2e-restart")

	// Slow build so we can interrupt coold mid-flight.
	payload := buildEnvelope(req, e.BuilderMgmt, slowRepo, "master", "localhost/"+req, ".")
	if err := e.redisXadd(payload); err != nil {
		t.Fatalf("XADD: %v", err)
	}

	// Wait until the transient unit is active.
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) && !e.unitActive(e.BuilderHost, req) {
		time.Sleep(1 * time.Second)
	}
	if !e.unitActive(e.BuilderHost, req) {
		t.Fatalf("transient unit never activated")
	}

	// Restart coold. The transient unit lives in system.slice and must
	// survive. resume_or_reap should adopt it on the new coold.
	if err := e.restartCoold(e.BuilderHost); err != nil {
		t.Fatalf("restart coold: %v", err)
	}
	time.Sleep(3 * time.Second)

	if !e.unitActive(e.BuilderHost, req) {
		t.Fatalf("transient unit did not survive coold restart")
	}

	// Cancel so the test doesn't drag on for 30+ minutes cloning kernel.
	if err := e.redisXadd(cancelEnvelope(req)); err != nil {
		t.Fatalf("cancel XADD: %v", err)
	}

	resp := e.waitBuildResp(t, req, 60*time.Second)
	if resp.Status != "error" || resp.Code != 499 {
		t.Fatalf("want cancel, got %+v", resp)
	}

	// Post-cancel: unit gone, workdir cleaned.
	time.Sleep(2 * time.Second)
	if e.unitActive(e.BuilderHost, req) {
		t.Fatalf("unit still active after adopted-cancel")
	}
	out, _ := e.sshRun(e.BuilderHost,
		"test -d /var/lib/coolify-builder/work/"+req+" && echo STILL || echo CLEANED")
	if !strings.Contains(out, "CLEANED") {
		t.Fatalf("workdir not cleaned: %q", out)
	}
}

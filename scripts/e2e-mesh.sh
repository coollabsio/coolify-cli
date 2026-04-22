#!/usr/bin/env bash
# End-to-end sanity test for the coolify mesh + firewall stack.
#
#   1. `coolify init apply` on two servers with two namespaces (default, alpha).
#   2. Start one nginx ("web-*") on SERVER_A and one alpine client ("client-*")
#      on SERVER_B inside each namespace — static --ip, --dns <bridge-gw>,
#      --restart=always so they survive reboot.
#      Also start client2-default on SERVER_A (same bridge as web-default) to
#      test intra-host nft bridge-family deny.
#   3. Verify cross-host traffic is DROPped by default (wget times out).
#   4. Verify intra-host same-bridge traffic is DROPped by default (nft plane).
#   5. Verify nft bridge table coolify_bridge present on both hosts.
#   6. `coolify firewall allow` per namespace (cross-host + intra-host).
#   7. Verify wget succeeds in both planes.
#   8. Re-run init apply to verify nft scaffold idempotency.
#
# Usage:
#   SERVERS=1.2.3.4,5.6.7.8 scripts/e2e-mesh.sh
#
# Required env:
#   SERVERS                   — exactly two SSH-reachable IPs, comma-separated.
#                               First = "host A" (web-* containers).
#                               Second = "host B" (client-* containers).
# Optional env:
#   SSH_KEY                   — default ~/.ssh/id_ed25519-no-pass (no passphrase)
#   SSH_USER                  — default root
#   COOLIFY_SSH_PASSPHRASE    — only if SSH_KEY is passphrase-protected;
#                               requires `sshpass` on PATH
#
# The script assumes `--container-pool` defaults (10.210.0.0/16, /24). With two
# hosts + two namespaces the allocator hands out 10.210.{0,1,2,3}.0/24; gateway
# is always .1, container IPs below are pinned to .10.

set -euo pipefail

SSH_KEY="${SSH_KEY:-$HOME/.ssh/id_ed25519-no-pass}"
SSH_USER="${SSH_USER:-root}"
SERVERS="${SERVERS:?set SERVERS=<host-a>,<host-b>}"

IFS=',' read -r SERVER_A SERVER_B EXTRA <<<"$SERVERS"
SERVER_A="${SERVER_A// /}"
SERVER_B="${SERVER_B// /}"
if [[ -z "$SERVER_A" || -z "$SERVER_B" || -n "${EXTRA:-}" ]]; then
  echo "SERVERS must contain exactly two comma-separated IPs (got: $SERVERS)" >&2
  exit 1
fi

: "${COOLIFY_SSH_PASSPHRASE:=}"
export COOLIFY_SSH_PASSPHRASE

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# Namespace → gateway IP on each host (matches allocator output).
GW_A_DEFAULT=10.210.0.1
GW_B_DEFAULT=10.210.1.1
GW_A_ALPHA=10.210.2.1
GW_B_ALPHA=10.210.3.1

# Container IPs (all pinned to .10 in each /24).
IP_WEB_DEFAULT=10.210.0.10      # host A, namespace default
IP_CLIENT_DEFAULT=10.210.1.10   # host B, namespace default
IP_WEB_ALPHA=10.210.2.10        # host A, namespace alpha
IP_CLIENT_ALPHA=10.210.3.10     # host B, namespace alpha
# Intra-host client on same bridge as web-default (host A, namespace default).
IP_CLIENT2_DEFAULT=10.210.0.11  # host A, namespace default

NGINX_IMAGE=docker.io/library/nginx:alpine
ALPINE_IMAGE=docker.io/library/alpine

SSH_OPTS=(-i "$SSH_KEY" -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10 -o BatchMode=yes)

say()  { printf '\n\033[1;36m==> %s\033[0m\n' "$*"; }
warn() { printf '\033[1;33m%s\033[0m\n' "$*" >&2; }
fail() { printf '\033[1;31m%s\033[0m\n' "$*" >&2; exit 1; }

# Use sshpass if passphrase was supplied; otherwise lean on ssh-agent / keyless.
ssh_exec() {
  local host="$1"; shift
  if [[ -n "$COOLIFY_SSH_PASSPHRASE" ]]; then
    SSHPASS="$COOLIFY_SSH_PASSPHRASE" sshpass -P "passphrase" -e \
      ssh "${SSH_OPTS[@]}" "$SSH_USER@$host" "$@"
  else
    ssh "${SSH_OPTS[@]}" "$SSH_USER@$host" "$@"
  fi
}

cli() {
  if [[ -n "$COOLIFY_SSH_PASSPHRASE" ]]; then
    go run ./coolify "$@" --ssh-key "$SSH_KEY" --ssh-user "$SSH_USER"
  else
    go run ./coolify "$@" --ssh-key "$SSH_KEY" --ssh-user "$SSH_USER"
  fi
}

# assert_blocked <host> <container> <target-ip-or-hostname>
assert_blocked() {
  local host="$1" client="$2" target="$3"
  if ssh_exec "$host" "podman exec $client wget -T 4 -qO- http://$target" >/dev/null 2>&1; then
    fail "expected timeout for $client@$host → $target but request succeeded"
  fi
  printf '  blocked: %s@%s → %s ✓\n' "$client" "$host" "$target"
}

# assert_flows <host> <container> <target-ip-or-hostname>
assert_flows() {
  local host="$1" client="$2" target="$3"
  if ! ssh_exec "$host" "podman exec $client wget -T 5 -qO- http://$target" | grep -q 'nginx'; then
    fail "$client@$host → $target failed to reach nginx"
  fi
  printf '  OK: %s@%s → %s ✓\n' "$client" "$host" "$target"
}

# ─── 1. init apply ────────────────────────────────────────────────────────────
say "1/8 coolify init apply on $SERVERS (namespaces: default, alpha)"
cli init apply \
  --servers "$SERVERS" \
  --namespaces default,alpha \
  --yes

# ─── 2. containers ────────────────────────────────────────────────────────────
say "2/8 creating containers with --ip / --dns / --restart=always"

run_container() {
  local host="$1" name="$2" network="$3" ip="$4" gw="$5" image="$6"; shift 6
  ssh_exec "$host" "podman rm -f $name >/dev/null 2>&1 || true"
  ssh_exec "$host" "podman run -d --name $name \
    --network $network --ip $ip --dns $gw --restart=always \
    $image $*"
}

# host A: nginx servers
run_container "$SERVER_A" web-default coolify-default-mesh "$IP_WEB_DEFAULT" "$GW_A_DEFAULT" "$NGINX_IMAGE"
run_container "$SERVER_A" web-alpha   coolify-alpha-mesh   "$IP_WEB_ALPHA"   "$GW_A_ALPHA"   "$NGINX_IMAGE"

# host B: alpine clients (sleep forever so we can exec into them)
run_container "$SERVER_B" client-default coolify-default-mesh "$IP_CLIENT_DEFAULT" "$GW_B_DEFAULT" "$ALPINE_IMAGE" sleep infinity
run_container "$SERVER_B" client-alpha   coolify-alpha-mesh   "$IP_CLIENT_ALPHA"   "$GW_B_ALPHA"   "$ALPINE_IMAGE" sleep infinity

# host A: 2nd client on same bridge as web-default — tests intra-host nft plane
run_container "$SERVER_A" client2-default coolify-default-mesh "$IP_CLIENT2_DEFAULT" "$GW_A_DEFAULT" "$ALPINE_IMAGE" sleep infinity

# ─── 3. cross-host default-deny ───────────────────────────────────────────────
say "3/8 confirming default-deny blocks cross-host traffic (expect timeouts)"

assert_blocked "$SERVER_B" client-default web-default.default.coolify.internal
assert_blocked "$SERVER_B" client-alpha   web-alpha.alpha.coolify.internal

# ─── 4. intra-host same-bridge default-deny (nft bridge plane) ────────────────
say "4/8 confirming intra-host same-bridge traffic blocked (nft bridge plane)"
# Raw IP intentional — DNS via bridge gateway also crosses the nft bridge hook;
# using raw IP isolates the firewall check from DNS-path correctness.
assert_blocked "$SERVER_A" client2-default "$IP_WEB_DEFAULT"

# ─── 5. nft table present on both hosts ───────────────────────────────────────
say "5/8 verifying nft bridge table coolify_bridge present on both hosts"
for host in "$SERVER_A" "$SERVER_B"; do
  ssh_exec "$host" "nft list table bridge coolify_bridge" >/dev/null \
    || fail "nft table coolify_bridge missing on $host"
  printf '  present: %s ✓\n' "$host"
done

# ─── 6. allow rules ───────────────────────────────────────────────────────────
say "6/8 adding allow rules (cross-host + intra-host)"

cli firewall allow \
  --servers "$SERVERS" \
  --namespace default \
  --from client-default --to web-default --port 80

cli firewall allow \
  --servers "$SERVERS" \
  --namespace alpha \
  --from client-alpha --to web-alpha --port 80

# Intra-host allow: client2-default → web-default on host A.
# Rule lands on host A (destination-host ownership); passing both servers is
# idempotent on the non-owner side.
cli firewall allow \
  --servers "$SERVERS" \
  --namespace default \
  --from client2-default --to web-default --port 80

# ─── 7. verify flow ───────────────────────────────────────────────────────────
say "7/8 verifying HTTP flows in both planes"

# Cross-host (iptables FORWARD plane)
assert_flows "$SERVER_B" client-default web-default.default.coolify.internal
assert_flows "$SERVER_B" client-alpha   web-alpha.alpha.coolify.internal

# Intra-host (nft bridge plane) — raw IP, same rationale as step 4
assert_flows "$SERVER_A" client2-default "$IP_WEB_DEFAULT"

# ─── 8. re-apply idempotency ──────────────────────────────────────────────────
say "8/10 re-running init apply — verifies nft scaffold idempotency (chain already exists regression)"
cli init apply \
  --servers "$SERVERS" \
  --namespaces default,alpha \
  --yes

# ─── 9. builder smoke test (static build) ─────────────────────────────────────
# Requires --central to have been passed to init apply. The script above does
# not pass --central, so builder capability may be disabled — gate on a marker
# file or just skip when /etc/coolify/jwt.priv is absent.
if ssh_exec "$SERVER_A" "test -f /etc/coolify/jwt.priv" >/dev/null 2>&1; then
  say "9/10 builder smoke test — push build:cmd, expect localhost image on central"

  REQ_ID="e2e-$(date +%s)"
  BUILD_PAYLOAD="{\"request_id\":\"$REQ_ID\",\"command\":{\"type\":\"static_build\",\"repo_url\":\"https://github.com/coollabsio/static-test-site\",\"git_ref\":\"main\",\"target_image\":\"localhost/e2e-$REQ_ID\"}}"

  ssh_exec "$SERVER_A" "redis-cli XADD build:cmd '*' payload '$BUILD_PAYLOAD'" >/dev/null

  DEADLINE=$(($(date +%s)+180))
  RESP=""
  while :; do
    RESP=$(ssh_exec "$SERVER_A" "redis-cli LPOP build:resp:$REQ_ID" 2>/dev/null || true)
    [[ -n "$RESP" && "$RESP" != "(nil)" ]] && break
    [[ $(date +%s) -ge $DEADLINE ]] && fail "builder smoke timed out after 180s"
    sleep 3
  done
  echo "$RESP" | grep -q '"status":"ok"' || fail "builder smoke returned error: $RESP"

  IMG_HOST=""
  for host in "$SERVER_A" "$SERVER_B"; do
    if ssh_exec "$host" "buildah images 2>/dev/null | grep -q localhost/e2e-$REQ_ID"; then
      IMG_HOST="$host"; break
    fi
  done
  [[ -n "$IMG_HOST" ]] || fail "image localhost/e2e-$REQ_ID not found on any host"
  printf '  OK: build succeeded; image on %s ✓\n' "$IMG_HOST"

  # ─── 10. cancel test ────────────────────────────────────────────────────────
  say "10/10 cancel test — dispatch then cancel; expect scope killed and cancel response"

  CAN_ID="e2e-cancel-$(date +%s)"
  CAN_BUILD="{\"request_id\":\"$CAN_ID\",\"command\":{\"type\":\"static_build\",\"repo_url\":\"https://github.com/torvalds/linux\",\"git_ref\":\"master\",\"target_image\":\"localhost/$CAN_ID\"}}"
  CAN_MSG="{\"request_id\":\"$CAN_ID\",\"command\":{\"type\":\"cancel\"}}"

  ssh_exec "$SERVER_A" "redis-cli XADD build:cmd '*' payload '$CAN_BUILD'" >/dev/null

  SCOPE_HOST=""
  for _ in 1 2 3 4 5 6 7 8 9 10; do
    sleep 2
    for host in "$SERVER_A" "$SERVER_B"; do
      if ssh_exec "$host" "systemctl list-units --no-legend --plain 'coolify-build-*.scope' 2>/dev/null | grep -q $CAN_ID"; then
        SCOPE_HOST="$host"; break 2
      fi
    done
  done
  [[ -n "$SCOPE_HOST" ]] || fail "scope coolify-build-$CAN_ID.scope never appeared"
  printf '  scope running on %s ✓\n' "$SCOPE_HOST"

  ssh_exec "$SERVER_A" "redis-cli XADD build:cmd '*' payload '$CAN_MSG'" >/dev/null

  DEADLINE=$(($(date +%s)+30))
  RESP=""
  while :; do
    RESP=$(ssh_exec "$SERVER_A" "redis-cli LPOP build:resp:$CAN_ID" 2>/dev/null || true)
    [[ -n "$RESP" && "$RESP" != "(nil)" ]] && break
    [[ $(date +%s) -ge $DEADLINE ]] && fail "cancel response timed out"
    sleep 2
  done
  echo "$RESP" | grep -q '"stage":"cancel"' || fail "expected stage=cancel in response, got: $RESP"

  if ssh_exec "$SCOPE_HOST" "systemctl is-active coolify-build-$CAN_ID.scope >/dev/null 2>&1"; then
    fail "scope still active after cancel: coolify-build-$CAN_ID.scope"
  fi
  printf '  OK: cancel SIGTERM killed cgroup; stage=cancel ✓\n'
else
  warn "skipping steps 9/10 (builder smoke + cancel): --central was not passed to init apply, so builder capability is not enabled"
fi

say "all checks passed"

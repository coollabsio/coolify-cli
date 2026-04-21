#!/usr/bin/env bash
# End-to-end sanity test for the coolify mesh + firewall stack.
#
#   1. `coolify init apply` on two servers with two namespaces (default, alpha).
#   2. Start one nginx ("web-*") on SERVER_A and one alpine client ("client-*")
#      on SERVER_B inside each namespace — static --ip, --dns <bridge-gw>,
#      --restart=always so they survive reboot.
#   3. Verify cross-host traffic is DROPped by default (wget times out).
#   4. `coolify firewall allow` per namespace.
#   5. Verify wget succeeds in both namespaces.
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

# ─── 1. init apply ────────────────────────────────────────────────────────────
say "1/5 coolify init apply on $SERVERS (namespaces: default, alpha)"
cli init apply \
  --servers "$SERVERS" \
  --namespaces default,alpha \
  --yes

# ─── 2. containers ────────────────────────────────────────────────────────────
say "2/5 creating containers with --ip / --dns / --restart=always"

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

# ─── 3. verify default-deny ───────────────────────────────────────────────────
say "3/5 confirming default-deny blocks cross-host traffic (expect timeouts)"

assert_blocked() {
  local client="$1" target="$2"
  if ssh_exec "$SERVER_B" "podman exec $client wget -T 4 -qO- http://$target" >/dev/null 2>&1; then
    fail "expected timeout for $client → $target but request succeeded"
  fi
  printf '  blocked: %s → %s ✓\n' "$client" "$target"
}

assert_blocked client-default web-default.default.coolify.internal
assert_blocked client-alpha   web-alpha.alpha.coolify.internal

# ─── 4. allow rules ───────────────────────────────────────────────────────────
say "4/5 adding allow rules (namespace-scoped)"

cli firewall allow \
  --servers "$SERVERS" \
  --namespace default \
  --from client-default --to web-default --port 80

cli firewall allow \
  --servers "$SERVERS" \
  --namespace alpha \
  --from client-alpha --to web-alpha --port 80

# ─── 5. verify flow ───────────────────────────────────────────────────────────
say "5/5 verifying cross-host HTTP now flows"

assert_flows() {
  local client="$1" target="$2"
  if ! ssh_exec "$SERVER_B" "podman exec $client wget -T 5 -qO- http://$target" | grep -q 'nginx'; then
    fail "$client → $target failed to reach nginx"
  fi
  printf '  OK: %s → %s ✓\n' "$client" "$target"
}

assert_flows client-default web-default.default.coolify.internal
assert_flows client-alpha   web-alpha.alpha.coolify.internal

say "all checks passed"

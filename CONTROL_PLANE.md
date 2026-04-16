# Coolify v5 Control Plane — Server Management Spec

This document lists everything the Coolify v5 control plane must implement on top of the bootstrap performed by `coolify init apply` to fully manage a fleet of mesh-connected hosts.

## What `coolify init apply --podman --default-deny` already provides

| Layer | Component | State |
|---|---|---|
| L3 mesh | WireGuard `wg0` per host with mgmt `/32` from `--wg-mgmt-pool` (default `100.64.0.0/16`) | Installed, configured, active |
| L3 mesh | Peer `AllowedIPs = <peer-mgmt>/32, <peer-container>/24` | Configured |
| Container runtime | Podman (distro apt) | Installed |
| Container runtime | `podman.socket` (rootful, `/run/podman/podman.sock`) | Enabled, active |
| Container network | `coolify-mesh` bridge per host with `/24` from `--container-pool` (default `10.210.0.0/16`), gateway `.1` | Created |
| Routing | `net.ipv4.ip_forward=1` (persisted via `/etc/sysctl.d/99-coolify-mesh.conf`) | Enabled |
| Firewall (mode A — `--podman` only) | `coolify-mesh-fw.service` with FORWARD ACCEPT for container subnet + POSTROUTING RETURN to skip podman MASQUERADE on wg0 | Active |
| Firewall (mode B — `--default-deny`) | `COOLIFY-INTRA` chain (ESTABLISHED/RELATED accept → COOLIFY-ALLOW → DROP), FORWARD jumps for `-s/-d <container-subnet>`, blanket ACCEPT removed | Active when set |
| Allow chain | `COOLIFY-ALLOW` (empty filter chain) | Created, ready for runtime rules |

Each host has a stable `(mgmt-ip, container-subnet)` pair. The bootstrap is idempotent — re-running `apply` only changes what drifted.

---

## What v5 control plane MUST implement

### 1. Inventory & state sync

- **Discovery**: query each host's `podman.socket` (over wg0 mgmt IP) for: containers, networks, volumes, images, system stats.
- **Drift detection**: periodically reconcile desired state (Coolify DB) against actual (podman API). Re-converge or alert.
- **Mesh join/leave**: when a host is added or removed from the cluster:
  - Add → invoke `coolify init apply` with the new `--servers` list (extends mesh, generates new mgmt IP + container `/24`, regenerates wg0 config on all peers).
  - Remove → invoke apply with reduced list; on the removed host, optionally tear down (out of scope for v1).

### 2. Container lifecycle

Talk to `podman.socket` REST API at `http://<mgmt-ip>/v5.0.0/libpod/...` (over wg0).

- Create container with `--network coolify-mesh` and explicit `--ip` from the host's `/24`.
  - Reserve container IPs in the control plane DB. Allocator skips `.1` (bridge gateway), reserves `.2-.254` for containers.
- Start, stop, restart, remove.
- Stream logs via `/containers/{id}/logs?follow=true` over the WG tunnel.
- Health checks via `/containers/{id}/healthcheck/run`.
- Resource limits, env vars, mounts, volumes, secrets — all standard podman API.

### 3. Network policy (firewall)

When host has `--default-deny` enabled, **all cross-host container traffic is dropped by default**. The control plane decides who talks to whom.

#### Allow-rule lifecycle

For an allow `(srcIP, dstIP)`:
- Add ACCEPT to `COOLIFY-ALLOW` on the host that **owns dstIP** (where DROP would otherwise fire).
- For bidirectional traffic (e.g. TCP), add the reverse `(dstIP, srcIP)` on the host that owns srcIP (so reply traffic is accepted on its way back through that host's FORWARD when it's the destination of the reply).
- In short: **a unidirectional allow is one rule on the destination host. A bidirectional allow is two rules on two hosts.**

#### Persistence model — recommended: systemd dropin

Don't use `iptables -A` directly (lost on reboot). Write a dropin per allow:

`/etc/systemd/system/coolify-mesh-fw.service.d/allow-<allow-uuid>.conf`:
```
[Service]
ExecStart=/bin/sh -c "/usr/sbin/iptables -C COOLIFY-ALLOW -s <SRC> -d <DST> -j ACCEPT 2>/dev/null || /usr/sbin/iptables -A COOLIFY-ALLOW -s <SRC> -d <DST> -j ACCEPT"
```

Then:
```bash
systemctl daemon-reload
systemctl restart coolify-mesh-fw.service
```

The base unit's `Type=oneshot, RemainAfterExit=yes` re-runs all `ExecStart=` lines on restart, including dropins. Survives reboots without `iptables-persistent`.

To remove an allow: delete the dropin file → `daemon-reload` → `restart`. Then optionally `iptables -D COOLIFY-ALLOW ...` for immediate effect (otherwise next packet hits DROP after restart).

Alternatively, `iptables -A` runtime + `iptables-save > /etc/iptables/rules.v4` + `netfilter-persistent` — but adds package dependency.

#### Allow API surface (control plane → CLI/agent)

```
POST   /api/v5/firewall/allow          {srcIP, dstIP, ports?, proto?}
DELETE /api/v5/firewall/allow/{id}
GET    /api/v5/firewall/allow          (list)
GET    /api/v5/firewall/allow/{id}     (status, last-applied)
```

Per-port allows: extend `ExecStart=` with `-p tcp --dport <PORT>` etc. — no scaffold change needed.

#### Intra-host isolation (NOT enforced by `--default-deny`)

Linux + netavark + Ubuntu 24.04: bridge L2 traffic bypasses iptables FORWARD even with `bridge-nf-call-iptables=1`. **Containers on the same host's `coolify-mesh` bridge can always reach each other.**

Two paths for v5 to enforce intra-host isolation:

- **(Recommended) Per-app podman networks**: each Coolify service = own podman network with `--opt isolate=true`. Different networks can't talk by default; use `podman network connect` for cross-app.
  - Trade-off: each network needs its own `/24` from container pool → wastes pool. Or carve `/27`s (allocator extension needed).
- **(Alternative) ebtables L2 filter**: `ebtables --logical-in podman1 --logical-out podman1 --ip-src X --ip-dst Y -j ACCEPT/DROP`. Independent toolchain, separate persistence. Bridge name discovery needed.

v1 ships without intra-host enforcement. v5 picks one path.

### 4. Container IP allocation per host

The bootstrap gives each host a `/24` (e.g. `10.210.0.0/24`). The control plane:
- Reserves `.1` (bridge gateway, skip).
- Allocates `.2-.254` for containers, deduplicated against running `podman ps` IPs.
- Pins IP via `podman run --ip <IP>` so DNS/firewall rules stay stable.
- Detects exhaustion early; alerts user to grow `--container-pool` or `--container-prefix`.

For `/24` per host: 253 containers max. For higher density: re-bootstrap with `--container-prefix 23` or larger pool.

### 5. Service discovery

- Containers don't get cross-host DNS automatically. v5 should run a DNS server (e.g. CoreDNS) per host or one cluster-wide:
  - Listening on each host's wg0 mgmt IP (`100.64.0.X:53`).
  - Records: `<service>.<namespace>.coolify.local A 10.210.X.Y`.
  - Configure containers via `--dns 100.64.0.X` (host's own mgmt IP).
- Alternative: rely on Coolify-injected env vars (e.g. `DB_HOST=10.210.0.5`). Simpler for v1.

### 6. Ingress (public traffic → containers)

`coolify init` doesn't manage public ingress. v5 deploys a reverse proxy (Traefik/Caddy) per host or HA pair:
- Listens on host public IP `:80/:443`.
- Routes `Host: app.example.com` → container IP (over container bridge or wg0 if cross-host).
- Cert management via ACME.
- Coolify generates proxy config from app routing rules.

Important: ingress proxy needs its own podman network OR can share `coolify-mesh`. Sharing means proxy can reach all containers — fine since it's the entrypoint.

### 7. Deployment workflows

- Image pull on target host(s) via `/images/pull`.
- Rolling deploy: create new container with new tag, healthcheck, swap proxy upstream, remove old.
- Volume + secrets mounted before start.
- Build pipeline: per-host build runner (separate podman socket call) or central builder + image push to registry.

### 8. Storage & volumes

- Local podman volumes per host (`/var/lib/containers/storage/volumes`).
- Cross-host: distributed FS (out of scope) OR pin stateful services to a host (anti-affinity rules in scheduler).
- Backup: `podman volume export` + scp to backup target. Coolify orchestrates schedule.

### 9. Scheduling

When user creates an app, control plane decides which host runs it:
- Round-robin / least-loaded / explicit pin.
- Pinned services (DB, persistent volumes) tracked in DB.
- Re-schedule on host failure (wg0 down, last-handshake stale).

Failure detection: poll `wg show wg0 latest-handshakes` on every host, parse seconds-since-handshake; alert if > N seconds.

### 10. Observability

Per host metrics (over wg0):
- `podman info` → version, storage driver, free space.
- `podman ps -a --format json` → container state.
- `podman stats --no-stream --format json` → CPU/mem per container.
- `wg show wg0 dump` → peer state, transfer bytes, latest handshake.
- `cat /proc/meminfo /proc/loadavg` → host load.
- `iptables -nvL COOLIFY-ALLOW` → allow rules + match counters (for audit).

Stream into central time-series store (Prometheus / VictoriaMetrics).

### 11. Updates

- Coolify runtime image self-updates (container restart with new image).
- WireGuard / Podman package updates: `coolify init apply` re-runs idempotently and picks up newer packages from apt. Schedule periodic re-apply (weekly?).
- Mesh config changes (new host, removed host) trigger re-apply on all hosts; control plane orchestrates.

### 12. Security posture

- **Private keys never leave hosts**: WG private key generated on remote, never transits SSH (already done by bootstrap).
- **Podman socket access**: `/run/podman/podman.sock` is rootful, exposed via `unix://`. Control plane connects via SSH tunnel OR via wg0 + a thin proxy (`podman system service tcp://100.64.0.X:2375`). Latter is simpler but exposes API on management network — acceptable since wg0 is trusted, but add TLS + auth for defense-in-depth.
- **SSH access**: bootstrap uses key-based SSH. Control plane should rotate SSH keys per agent install, store in encrypted DB.
- **Host firewall (iptables INPUT chain)**: bootstrap doesn't lock down INPUT. v5 should drop public access to ports other than `:51820/udp` (WG), `:22/tcp` (SSH), `:80/:443` (ingress).
- **Audit**: log every COOLIFY-ALLOW change with who-when-why metadata.

### 13. Failure modes & recovery

| Failure | Detection | Recovery |
|---|---|---|
| Host SSH unreachable | bootstrap apply error | Manual investigation; node marked unhealthy in DB |
| WG peer offline (`latest_handshake > 180s`) | `wg show` poll | Mark unhealthy; re-schedule containers if pinning permits |
| Podman socket unreachable | API call timeout | Restart `podman.socket`; if persistent, re-bootstrap |
| Firewall service failed | `systemctl is-active != active` | Re-run `coolify init apply`; service is idempotent |
| Container OOM/crash | `podman events` watcher | Restart per restart policy; alert after N crashes |
| Container subnet exhausted | allocator returns error | Alert; offer apply with bigger `--container-prefix` |
| Mgmt IP exhausted | allocator returns error | Alert; rare for /16 |
| `coolify-mesh` bridge missing | probe `podman network exists` returns no | Re-run apply |
| User manually deletes COOLIFY-ALLOW chain | runtime check | Re-run apply (recreates chain via service restart) |

### 14. Multi-tenancy (deferred)

If Coolify ever supports tenant isolation:
- Tenant = own podman network namespace per host.
- Allows always scoped within tenant; cross-tenant requires explicit allow.
- Pool subdivided per tenant. Allocator extension.

Not in v1 or v5 initial.

---

## Out of scope (now and likely v5)

- Rootless containers (would need user namespace mapping, separate sockets per user).
- IPv6 mesh (`fdcc::` style, ip6tables mirror).
- Hardware-level isolation (SELinux profiles, AppArmor).
- Live migration (qemu/criu).
- Distributed storage (Ceph/Longhorn).
- macvlan / SR-IOV networking.
- Autoscaling.
- BGP / external network announcements.

---

## Quick reference — operations the agent CLI should expose

(Future `coolify-cli` subcommands beyond `init`)

```
coolify deploy <app>           # build + push + run
coolify scale <app> --replicas N
coolify firewall allow <src> <dst> [--port N --proto tcp]
coolify firewall deny <id>
coolify firewall list
coolify host list              # show mesh state, last-handshake, container count
coolify host add <ip> --ssh-key K
coolify host remove <ip>
coolify logs <container>
coolify exec <container> -- sh
```

These all wrap podman API calls + mesh state queries over wg0.

---

## Summary

`coolify init apply` does the **one-shot host bootstrap**: WG mesh, podman runtime, bridge network, default-deny scaffold. After that, **everything dynamic is the v5 control plane's job**: container lifecycle, allow rules in COOLIFY-ALLOW (via systemd dropins for persistence), scheduling, observability, ingress, updates.

The two pieces communicate via:
1. **SSH** for bootstrap + re-converge (idempotent re-runs).
2. **Podman REST API** over wg0 mgmt IPs for runtime ops.
3. **Filesystem dropins** in `/etc/systemd/system/coolify-mesh-fw.service.d/` for persistent firewall state.

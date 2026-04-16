# Coolify v5 Control Plane — Server Management Spec

This document lists everything the Coolify v5 control plane must implement on top of the bootstrap performed by `coolify init apply` to fully manage a fleet of mesh-connected hosts.

## Architecture overview

```
┌────────────────────────────┐
│  Coolify central UI / API  │
│  (single instance / HA)    │
└──────────────┬─────────────┘
               │ HTTPS over wg0 (TLS + bearer token)
               │ to 100.64.0.X:8443
               ▼
┌────────────────────────────┐  ┌─────────────────────────┐
│  coold (per-host agent)    │  │  /run/podman/podman.sock│
│  - REST API on wg0 :8443   │──┤  bind-mount, host-only  │
│  - RBAC, audit, rate limit │  │  (NEVER on network)     │
│  - Talks ONLY to local sock│  └─────────────┬───────────┘
└────────────────────────────┘                │
                                              ▼
                              ┌─────────────────────────────┐
                              │  podmand (containers, nets) │
                              └─────────────────────────────┘
```

**Key principle**: `/run/podman/podman.sock` is **never exposed on TCP**. coold (per-host agent container or systemd service) bind-mounts the socket and proxies a curated REST API over wg0. Central Coolify never touches the raw podman API directly.

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

Talk to **coold** REST API at `https://<mgmt-ip>:8443/...` (over wg0). coold proxies to the local `/run/podman/podman.sock` Unix socket.

- Create container with `--network coolify-mesh` and explicit `--ip` from the host's `/24`.
  - Reserve container IPs in the control plane DB. Allocator skips `.1` (bridge gateway), reserves `.2` for coold itself, `.3-.254` for app containers.
- Start, stop, restart, remove.
- Stream logs via coold's `/containers/{id}/logs?follow=true` (which proxies podman API over the wg0 tunnel).
- Health checks via `/containers/{id}/healthcheck/run`.
- Resource limits, env vars, mounts, volumes, secrets — all standard podman API surfaced through coold.

#### coold deployment

coold runs as a privileged container on each host (or as a host systemd service). v5 control plane installs it via `coolify init apply` after the mesh + podman + bridge are up — OR `coolify init` could grow a `--coold` flag that installs it as part of bootstrap (out of scope for v1, but trivial extension).

Reference container spec:
```bash
podman run -d --name coold --restart=always \
  --network coolify-mesh --ip 10.210.X.2 \
  -v /run/podman/podman.sock:/run/podman/podman.sock \
  -v /etc/coolify/coold:/etc/coolify/coold:ro \
  --security-opt label=disable \
  -p 100.64.0.X:8443:8443 \
  ghcr.io/coollabs/coold:latest
```

- Listens on host's WG mgmt IP only (`100.64.0.X:8443`) — unreachable from public internet.
- TLS cert + bearer token auth on every request.
- Allow rule needed in `COOLIFY-ALLOW`: central Coolify host's mgmt IP → this host's `100.64.0.X:8443`. (Alternatively: skip default-deny for mgmt subnet — see §3.)

### 3. Network policy (firewall)

When host has `--default-deny` enabled, **all cross-host container traffic is dropped by default**. The control plane decides who talks to whom.

#### Division of labour: bootstrap vs coold

| Layer | Owner | Responsibility |
|---|---|---|
| Chain scaffold (COOLIFY-INTRA, COOLIFY-ALLOW, FORWARD jumps, conntrack early-accept, POSTROUTING RETURN) | `coolify init apply` (one-shot) | Install + idempotently re-converge on flag change. Never touches individual allow rules. |
| Allow rules inside `COOLIFY-ALLOW` | **coold** | Sole owner. Persists in coold's own DB. Applies in batch. Survives reboot via coold autostart. |

**`coolify init` is intentionally not the rule store.** Bootstrap creates the empty allow chain; coold fills it.

#### Allow-rule lifecycle

For an allow `(srcIP, dstIP)`:
- Add ACCEPT to `COOLIFY-ALLOW` on the host that **owns dstIP** (where DROP would otherwise fire).
- For bidirectional traffic (e.g. TCP, ICMP echo+reply), add the reverse `(dstIP, srcIP)` on the host that owns srcIP. (Reply packets traverse THAT host's FORWARD chain when arriving back, and dst-side check fires there.)
- **One unidirectional allow = one rule on one host. One bidirectional allow = two rules on two hosts.**
- Conntrack ESTABLISHED early-accept (installed by bootstrap) handles in-flow follow-up packets — no need to add per-packet rules.

#### Persistence + scale model — coold owns rules

Per-rule systemd dropins do NOT scale (1000 rules × `daemon-reload` + restart = minutes, fs clutter, audit nightmare). Instead:

```
coold service (per host)
  ├─ DB:  /var/lib/coold/allows.db   (sqlite, source of truth)
  ├─ Boot:        load DB → emit nft/iptables-restore batch
  ├─ API mutate:  insert/delete row + apply incremental iptables -A/-D
  └─ Reconcile:   periodic full reload from DB to detect drift
```

Apply paths:

| Backend | Bulk apply (1000 rules) | Atomicity |
|---|---|---|
| `iptables -A` per rule | ~5s | per-rule |
| `iptables-restore --noflush` (preferred for iptables-legacy) | ~50ms | per-batch |
| `nft -f /tmp/rules.nft` (preferred when host uses nftables backend) | ~10ms | atomic transaction |

coold detects backend (`iptables --version` or presence of nftables socket) and picks. Bootstrap doesn't care.

For **systemctl restart coolify-mesh-fw.service** (e.g. `coolify init apply` re-runs after a flag flip): the unit flushes COOLIFY-INTRA but **never flushes COOLIFY-ALLOW** — coold's rules survive. If somehow lost (manual `iptables -F COOLIFY-ALLOW`), coold's reconcile loop replays from DB within seconds.

#### Allow API surface (central Coolify → coold REST)

```
POST   /api/v1/firewall/allow          {src, dst, proto?, port?, comment?}    → returns id
DELETE /api/v1/firewall/allow/{id}
GET    /api/v1/firewall/allow                                                  list
GET    /api/v1/firewall/allow/{id}                                             show + match counters
POST   /api/v1/firewall/allow/bulk     {add: [...], remove: [...]}             atomic batch
POST   /api/v1/firewall/reconcile                                              force full reload
```

coold translates each row into the right iptables/nft fragment. Per-port: `-p tcp --dport <N>`. Source/dest IP, CIDR, or set reference (for grouping like "all-frontend-ips").

For very large rule sets: use **nftables sets** so a rule references a set name, and the set membership changes are O(1):

```
nft add element ip filter coolify_allowed_pairs { 10.210.0.10 . 10.210.1.10 }
```

One static rule like `ct state new ip saddr . ip daddr @coolify_allowed_pairs accept` evaluates in O(log n) regardless of set size. coold maintains the set rather than thousands of rules. Optional optimization for v5+.

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

**Pattern**: embedded DNS server in coold, backed by [Corrosion](https://github.com/superfly/corrosion) (CRDT sqlite gossiped via SWIM across the mesh). No env injection. No container restarts on backend movement.

#### Why DNS-via-coold over alternatives

| Approach | Stable target? | Backend move = restart? | Complexity |
|---|---|---|---|
| Env injection (`DB_HOST=10.210.5.42`) | no — IP changes | yes (rolling redeploy on every change) | medium (template engine + dep graph) |
| **Embedded DNS in coold** | **yes (hostname)** | **no** | **low (~200 LoC)** |
| VIP per service | yes (IP) | no | high (keepalived/BGP/IPVS) |
| Per-host HTTP/TCP proxy | yes (port) | no | medium (proxy config) |

DNS chosen: smallest moving parts, works for any protocol, standard `getaddrinfo()` path, ubiquitous client support.

#### Corrosion schema (replicated sqlite)

```sql
CREATE TABLE services (
    id              TEXT PRIMARY KEY,         -- "myapp.db"
    coolify_app_id  TEXT NOT NULL,
    name            TEXT NOT NULL,            -- "db"
    namespace       TEXT NOT NULL,            -- "myapp"
    port            INTEGER,                  -- canonical port (informational)
    updated_at      INTEGER NOT NULL          -- ms epoch (CRDT clock)
);

CREATE TABLE service_endpoints (
    service_id      TEXT NOT NULL,
    container_id    TEXT NOT NULL,
    host_mgmt_ip    TEXT NOT NULL,            -- 100.64.0.X (host running the container)
    container_ip    TEXT NOT NULL,            -- 10.210.X.Y
    healthy         INTEGER NOT NULL,
    updated_at      INTEGER NOT NULL,
    PRIMARY KEY (service_id, container_id)
);
```

Each coold writes its own host's container facts. Reads are local sqlite (sub-ms). Gossip handles distribution; convergence ~1s in small clusters.

#### Embedded DNS server

```go
// pseudocode — ~200 LoC total
func (c *Coold) serveDNS() {
    pc, _ := net.ListenPacket("udp", "10.210.X.1:53")  // bridge gateway IP
    for {
        buf := make([]byte, 512)
        n, addr, _ := pc.ReadFrom(buf)
        go c.handle(buf[:n], addr, pc)
    }
}

func (c *Coold) handle(query []byte, src net.Addr, pc net.PacketConn) {
    msg := dns.Unpack(query)
    name := msg.Questions[0].Name  // "myapp.db.coolify.internal."

    if !strings.HasSuffix(name, ".coolify.internal.") {
        // Forward to upstream (configurable; default 1.1.1.1).
        pc.WriteTo(c.upstream.Query(msg), src)
        return
    }

    serviceID := strings.TrimSuffix(name, ".coolify.internal.")
    var ips []string
    c.corrosion.Query(`
        SELECT container_ip FROM service_endpoints
        WHERE service_id = ? AND healthy = 1
    `, serviceID).Scan(&ips)

    if len(ips) == 0 {
        pc.WriteTo(dns.NXDOMAIN(msg), src); return
    }
    pc.WriteTo(dns.AnswerA(msg, ips, ttl=5), src)
}
```

Listens on **bridge gateway IP** (`10.210.X.1:53`) of the host's `coolify-mesh` bridge — reachable from every container in the host's `/24` via standard kernel routing.

#### Container creation hook

Every container coold creates gets:
```
podman run --dns 10.210.X.1 --dns-search coolify.internal ...
```

App code uses short names: `getaddrinfo("myapp.db", ...)` → libc appends search suffix → `myapp.db.coolify.internal` → coold answers from local Corrosion.

#### Resolution flow

```
1. App in container A on host-1 (10.210.0.10) calls getaddrinfo("myapp.db")
2. libc reads /etc/resolv.conf:
     nameserver 10.210.0.1
     search coolify.internal
3. UDP query "myapp.db.coolify.internal" → 10.210.0.1:53
4. coold@host-1 reads local Corrosion → 10.210.5.42 (running on host-3)
5. Reply: A 10.210.5.42, TTL=5
6. App opens TCP to 10.210.5.42:5432
7. Routed via wg0 (peer host-3's AllowedIPs covers 10.210.5.0/24)
   → bridge → container
8. (If --default-deny is on, COOLIFY-ALLOW on host-3 must permit
    10.210.0.10 → 10.210.5.42.)
```

#### Backend movement (zero restart on dependents)

```
T+0:   myapp.db @ 10.210.5.42 on host-3. Endpoint row gossiped.
T+10s: User redeploys myapp.db on host-3.
       coold@host-3:
         - new container at 10.210.5.43
         - INSERT new endpoint row (10.210.5.43)
         - DELETE old endpoint row (10.210.5.42)
         - kill old container
       Corrosion gossips delta.
T+11s: All hosts have updated state.
T+15s: App on host-1 has stale TCP to 10.210.5.42 — broken when old container died.
       App's reconnect logic re-resolves myapp.db → 10.210.5.43 → reconnects.
       App container NEVER restarted, env NEVER changed.
```

App must have reconnect logic (every reasonable DB/cache client does). DNS provides the new IP transparently.

#### TTL

5s. Trade-off:
- Lower = faster failover, more queries.
- Higher = quieter DNS, slower failover.

Apps with infinite-cache resolvers (Java's `networkaddress.cache.ttl=-1`) won't see updates. Document for users; not coold's problem.

#### Multi-replica services

Resolver returns ALL healthy A records. Apps with proper conn pools (postgres, redis clients) handle multi-target naturally. No client-side LB protocol needed.

#### Health & staleness

- coold marks `healthy=0` on healthcheck fail. DNS stops returning that IP within next query.
- Stale-row TTL: rows older than 60s without heartbeat are pruned (owning coold heartbeats every 15s).

#### TLD

`.coolify.internal` — `.internal` is RFC 6761 reserved for private use. Won't collide with public TLDs. Configurable per-cluster.

#### Failure modes

| Failure | Behaviour |
|---|---|
| coold dies | Cluster DNS resolution stops. systemd restarts coold (~3s). Existing connections survive. Same profile as k8s losing CoreDNS. |
| Corrosion split-brain | Each partition serves local view; CRDT merges cleanly when partition heals. May serve stale IPs during partition. |
| Backend healthy in DB but unreachable | DNS returns IP → app connection fails → app retries. If multi-replica, may pick different one on retry. |
| Container has no `--dns` (created outside coold) | No cluster resolution. Document: only coold-managed containers get discovery. |
| Cross-region high latency | Slower convergence; stale DNS for 10–30s. Acceptable v1. |

#### API surface (central Coolify → coold REST)

```
POST   /api/v1/services/register      {service_id, app_id, name, namespace, port, container_id, container_ip, host_mgmt_ip}
DELETE /api/v1/services/{service_id}/endpoints/{container_id}
GET    /api/v1/services/{service_id}/endpoints
GET    /api/v1/services?namespace=myapp
GET    /api/v1/dns/lookup/{name}      (debug — what coold would answer)
GET    /api/v1/dns/stats              (qps, hit/miss/forward counts)
```

Most ops are automatic side effects of deploy/scale/health-check. Central rarely calls `/services/register` directly — coold registers on container create, deregisters on remove.

#### Bootstrap impact

Zero. `coolify init apply` doesn't change. Bridge gateway `10.210.X.1` was always reserved by `MachineIP()`; coold binds port 53 on it when deployed.

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
- **Podman socket access**: `/run/podman/podman.sock` stays as a rootful Unix socket on each host — **NEVER exposed on TCP**. Only **coold** (per-host agent, see §2) has access via bind-mount. coold surfaces a curated REST API over wg0 with TLS + bearer auth. This means:
  - Compromise of a non-coold container does NOT grant podman API access.
  - All container ops are auditable at the coold API layer (RBAC, rate limit, deny dangerous flags like `--privileged`).
  - No `podman system service tcp://...` listener; no need for socket-level TLS.
  - Central Coolify only knows the coold endpoint, not the underlying socket.
- **SSH access**: bootstrap uses key-based SSH. Control plane should rotate SSH keys per agent install, store in encrypted DB. After bootstrap, day-to-day ops go via coold REST — SSH is for re-bootstrap only.
- **Host firewall (iptables INPUT chain)**: bootstrap doesn't lock down INPUT. v5 should drop public access to ports other than `:51820/udp` (WG), `:22/tcp` (SSH), `:80/:443` (ingress). coold's `:8443` binds to the wg0 IP only, so it's already not on the public interface.
- **coold port reachability**: with `--default-deny`, central Coolify needs an allow in COOLIFY-ALLOW on each managed host: `-s <central-mgmt-ip>/32 -d <host-mgmt-ip>/32 -p tcp --dport 8443 -j ACCEPT`. v5 installs this allow as part of "host join" workflow.
- **Audit**: log every COOLIFY-ALLOW change with who-when-why metadata; coold mirrors with API-level audit log.

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

The pieces communicate via:
1. **SSH** for bootstrap + re-converge (idempotent re-runs).
2. **coold REST API** over wg0 mgmt IPs (`https://100.64.0.X:8443`) for runtime ops. coold is the *only* process with access to the local podman socket AND the sole owner of allow rules in COOLIFY-ALLOW.

Persistence model:
- Bootstrap state (chains, jumps, conntrack accept) → idempotent `coolify init apply` re-runs.
- Allow rules → coold's own DB, applied via `iptables-restore --noflush` or `nft -f`. No per-rule systemd dropin (doesn't scale to 1000s of rules).

The podman socket is host-local. There is no TCP podman API. coold is the security/audit boundary AND the firewall state authority between the central Coolify control plane and the host.

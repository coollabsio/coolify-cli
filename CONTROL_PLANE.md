# Coolify v5 Control Plane — Server Management Spec

This document lists everything the Coolify v5 control plane must implement on top of the host provisioning performed by the `coolify init` subcommand tree (`bootstrap` for first install, `extend` for adding hosts, `upgrade` for bumping agent versions) to fully manage a fleet of mesh-connected hosts.

## Architecture overview

```
┌─────────────────────────────────────┐
│  Coolify central UI / API           │
│  - Multi-tenant (cloud) or 1-tenant │
│    (self-hosted); same binary       │
│  - WSS / gRPC bidi stream listener  │
│    on :443 (public)                 │
│  - Routes commands by host_id       │
└────────────────────▲────────────────┘
                     │ outbound TLS :443 (WSS / gRPC bidi)
                     │ long-lived, resumable, jittered reconnect
                     │ per-host JWT (issued at enroll)
                     │
   ┌─────────────────┴──────────────────┐
   │      (per-customer gateway,        │
   │       OPTIONAL — one mesh host     │
   │       proxies N coolds → 1 stream) │
   └─────────────────▲──────────────────┘
                     │ same stream protocol, over wg0
                     │
┌────────────────────┴────────────────┐  ┌─────────────────────────┐
│  coold (per-host agent)             │  │  /run/podman/podman.sock│
│  - Dials central (or gateway) out   │──┤  bind-mount, host-only  │
│  - Local REST on wg0 :8443          │  │  (NEVER on network)     │
│    (intra-mesh callers: CLI, peers) │  └─────────────┬───────────┘
│  - Bearer-token authn (both paths)  │                │
│  - Talks ONLY to local podman sock  │                ▼
└─────────────────────────────────────┘  ┌─────────────────────────────┐
                                         │  podmand (containers, nets) │
                                         └─────────────────────────────┘
```

**Key principles**:

1. **`/run/podman/podman.sock` is never exposed on TCP.** coold bind-mounts it and proxies a curated API. Central Coolify never touches the raw podman socket directly.
2. **coold always dials outbound — never accepts inbound from central or public internet.** One topology for self-hosted and cloud SaaS. Works through any NAT/corp firewall, scales to thousands of hosts per central region (10k+ idle streams are cheap). No "add central to every customer's wg0" — central never joins any mesh.
3. **coold still exposes a local REST API on wg0 mgmt IP** for intra-mesh callers only (the `coolify firewall` CLI via SSH-bounce, other coolds in the same mesh, a per-customer gateway if deployed). Never reachable from public internet; wg0 is the only L3 boundary that can hit it.
4. **Per-customer gateway (optional)**: for large customers, one host in the mesh runs a stream aggregator that dials central once and proxies commands to the other coolds over wg0. Reduces stream fan-out at central from N-per-customer to 1-per-customer; adds one hop of latency. Transparent to both ends — same protocol each side.

## What `coolify init bootstrap` already provides

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
  - Add → invoke `coolify init extend --servers <full list> --new-hosts <new host>` (installs the new host end-to-end, regenerates wg0 config on every existing peer with the new mgmt IP + namespace `/24`s, leaves agent binaries on existing hosts untouched).
  - Remove → not supported by a first-class subcommand today. Documented workaround for alpha: tear the host out-of-band (stop services, drop it from DNS) and re-run `coolify init bootstrap` with the reduced `--servers` list on a maintenance window; a dedicated `remove-host` flow is a follow-up.

### 2. Container lifecycle

Every container op is a command sent over coold's outbound stream (central → coold) or a local REST call on coold's wg0 listener (intra-mesh → coold). coold executes the command against the local `/run/podman/podman.sock` Unix socket and streams results back.

- Create container with `--network coolify-mesh` and explicit `--ip` from the host's `/24`.
  - Reserve container IPs in the control plane DB. Allocator skips `.1` (bridge gateway), reserves `.2` for coold itself, `.3-.254` for app containers.
- Start, stop, restart, remove.
- Stream logs via `/containers/{id}/logs?follow=true` (coold relays podman API frames over the open control stream).
- Health checks via `/containers/{id}/healthcheck/run`.
- Resource limits, env vars, mounts, volumes, secrets — all standard podman API surfaced through coold.

#### coold is a primitive proxy, not an app brain

coold follows the **kubelet analogue**: it knows containers, images, volumes, networks, iptables, and Corrosion writes. It does **not** know apps, compose, Dockerfiles, buildpacks, or Nixpacks. Central Coolify is the apiserver+controllers: it parses app-level config and compiles it into a sequence of primitive ops streamed to coold.

Test for "should this live in coold?": could a second orchestrator (a Nomad-style competitor) reuse this coold with a different app model? If yes → coold. If no → central.

#### Wire surface (enumerable)

Same endpoint set on both transports (outbound stream from central, local REST on wg0 for intra-mesh callers). New verbs require a coold release — there is no `/podman/raw` passthrough.

```
# Images
POST   /api/v1/images/pull           {ref, auth?}            -> {digest}
GET    /api/v1/images                                         -> [{ref, digest, size}]
DELETE /api/v1/images/{ref}

# Containers (filtered podman surface)
POST   /api/v1/containers            <create spec>            -> {id}
POST   /api/v1/containers/{id}/start
POST   /api/v1/containers/{id}/stop          {timeout?}
POST   /api/v1/containers/{id}/restart
DELETE /api/v1/containers/{id}                {force?}
GET    /api/v1/containers/{id}                (inspect)
GET    /api/v1/containers/{id}/logs?follow=true               (streamed)
POST   /api/v1/containers/{id}/exec           {cmd, tty?}     (streamed)
POST   /api/v1/containers/{id}/healthcheck/run

# Volumes
POST   /api/v1/volumes               {name, driver, labels}
DELETE /api/v1/volumes/{name}
GET    /api/v1/volumes/{name}

# Networks (bootstrap creates coolify-mesh; extra per-app nets created here)
POST   /api/v1/networks              {name, driver, options, labels}
DELETE /api/v1/networks/{name}
GET    /api/v1/networks

# Firewall (coold = sole writer)
POST   /api/v1/firewall/allow        {src, dst, proto?, port?}  -> {id}
DELETE /api/v1/firewall/allow/{id}
GET    /api/v1/firewall/allow

# Service endpoints (Corrosion writer; used by central to register deploys)
POST   /api/v1/services/register
DELETE /api/v1/services/{id}/endpoints/{container_id}
GET    /api/v1/services/{id}/endpoints

# DNS (diagnostics)
GET    /api/v1/dns/lookup/{name}
GET    /api/v1/dns/stats

# Host facts (read-only; central scrapes these for observability + scheduling)
GET    /api/v1/host/info             (podman info, kernel, wg state, load)
GET    /api/v1/host/containers       (podman ps -a)
GET    /api/v1/host/stats            (podman stats snapshot)
```

**Deny filter on `POST /containers`** (defense-in-depth even though central is trusted):
- Block `--privileged`, `--cap-add=SYS_ADMIN/NET_ADMIN` unless host is marked `allow_privileged=true`.
- Block host-path bind mounts outside a configurable allowlist (default: none).
- Block host netns (`--net=host`) unless the container is coold itself.

Anything not above is not coold's job. No `/apps`, `/deployments`, `/compose`, `/build`, `/podman/raw`. coold does not parse compose, Dockerfiles, buildpacks, or any app-level config — central compiles these into sequences of the primitive ops above and streams them down.

#### Networks

Default = shared `coolify-mesh` bridge. Containers get `.coolify.internal` DNS + flat L3 across the mesh. Users may define extra podman networks per app (docker-compose `networks:` style) via `POST /networks` + container attach on create. Central compiles compose into network-create + container-attach primitives.

#### coold deployment

coold runs as a privileged container on each host (or as a host systemd service). `coolify init bootstrap` puts it in place at install time (and `coolify init upgrade` bumps its version later): binary, systemd unit with `COOLD_API_BIND=<wg0-mgmt-ip>:8443`, random per-host bearer token at `/etc/coolify/api-token` (mode 0600), outbound stream config written atomically to `/etc/coolify/coold.env`.

Reference container spec (equivalent to systemd-service deployment):
```bash
podman run -d --name coold --restart=always \
  --network coolify-mesh --ip 10.210.X.2 \
  -v /run/podman/podman.sock:/run/podman/podman.sock \
  -v /etc/coolify/coold:/etc/coolify/coold:ro \
  --security-opt label=disable \
  -p 100.64.0.X:8443:8443 \
  ghcr.io/coollabs/coold:latest
```

- **Outbound stream**: coold dials `wss://<central-host>/v1/agent` (or gRPC bidi) on start, presenting its per-host JWT. Central routes commands to it by host id over the open stream. Stream is the primary control channel for both self-hosted and cloud SaaS — same code path, same binary.
- **Local REST on wg0 mgmt IP (`100.64.0.X:8443`)**: accepts intra-mesh callers only (the `coolify firewall` CLI via SSH-bounce, other coolds in the same mesh, a per-customer gateway). Not reachable from public internet — wg0 is the L3 boundary. Bearer-token auth on every request.
- **No inbound from central**: central never dials coold. All mutations arrive over the coold-initiated stream; no `COOLIFY-ALLOW` rule for "central → host:8443" needed. Works through NAT/corp firewalls.

#### Control channel transport (stream)

Two candidates; spec-time decision, not per-host:

| Option | Pros | Cons |
|---|---|---|
| **gRPC bidi stream over HTTP/2** *(chosen)* | typed Protobuf schemas, native server-streaming for logs/exec, versionable wire | stricter proxy requirements (some corp proxies still mangle HTTP/2); larger runtime |
| WebSocket (WSS over :443) *(fallback)* | traverses every proxy, tiny overhead, libs everywhere | framing is custom-on-top; manual request/response correlation |

**Decision: gRPC bidi + Protobuf.** Typed schemas + native server-streaming for logs and exec outweigh the proxy risk; WSS remains the documented fallback if gRPC-through-proxy issues show up in the field. Both run on :443, so customer-side egress rules stay unchanged either way.

#### Enrollment

coold registers once at install using a one-time token from central:

```bash
coolify init bootstrap \
  --central-url https://cloud.coolify.io \
  --enroll-token <one-time-hex>
```

1. coold POSTs `(host_id, wg0_mgmt_ip, container_subnet, enroll_token)` to `https://<central>/v1/enroll`.
2. Central validates the enroll token (scoped to a tenant, single-use, short TTL) and issues a long-lived per-host JWT + TLS-pinned central cert. Response stored in `/etc/coolify/coold.env` (mode 0600).
3. coold burns the enroll token and switches to JWT for the persistent stream.
4. Central revokes by invalidating the JWT in its own DB; next stream reconnect fails auth and the host is quarantined until re-enrolled.

#### Reconnect + fleet-restart storms

Single-central-restart would otherwise trigger simultaneous reconnects from every host. Mitigations:

- **Jittered backoff**: exponential from 1s up to 60s with full jitter. 10k hosts reconnecting spread across ~minutes, not seconds.
- **Resumable streams**: stream carries a monotonic `last_seq` per host so central can replay missed commands after reconnect without central-side queueing beyond an in-memory ring buffer.
- **Region sharding**: DNS round-robin or geo-steering across multiple central stream gateways; each gateway holds O(10k) streams. Stateful routing via consistent-hashing on host_id so a host lands on the same gateway across reconnects (cache affinity).

#### Per-customer gateway (optional)

For customers with 50+ hosts, one designated mesh host runs a **gateway mode coold** (same binary, different role):

- Dials central like any other coold.
- Accepts incoming streams from its peer coolds over wg0 (they dial `wss://<gateway-mgmt-ip>:8443/v1/agent-peer` instead of central).
- Relays commands down, responses up. Maintains O(hosts-in-mesh) inbound streams + 1 outbound to central.

Saves N-1 WAN streams at central per customer; costs one hop of latency + one more thing to keep alive. Opt-in via `coolify init bootstrap --gateway-for-mesh` on the chosen host; peers get `--via-gateway <gateway-mgmt-ip>` at install.

### 3. Network policy (firewall)

When host has `--default-deny` enabled, **all cross-host container traffic is dropped by default**. The control plane decides who talks to whom.

#### Division of labour: bootstrap vs coold vs central

| Layer | Owner | Responsibility |
|---|---|---|
| Chain scaffold (COOLIFY-INTRA, COOLIFY-ALLOW, FORWARD jumps, conntrack early-accept, POSTROUTING RETURN) | `coolify init bootstrap` (also reconverges on `extend`) | Install + idempotently re-converge on flag change. Never touches individual allow rules. |
| Rule metadata (who/when/why, audit log, RBAC, tenant scoping, app→rule mapping) | **Coolify central DB** | Authoritative store. All rich queries, audit trails, and access control live here. |
| Raw rule tuples `(src, dst, proto, port)` on the host | **coold** (single writer) | Apply to kernel + snapshot to `/etc/coolify/allow.rules` for reboot. Stateless-ish — just a cache of what the caller (central Coolify or `coolify firewall` CLI) told it to apply. No metadata, no DB. |

**Key split**: central Coolify owns rich state (metadata, audit, RBAC). Per-host coold owns only the raw rules needed to program the kernel + survive reboot. This keeps coold small and lets a single central DB be the source of truth for all cross-cutting concerns.

**App-topology compilation happens in central.** coold applies the rule tuples it is told to apply; it does not generate rules from app intent (e.g. "allow service `web` → `db`"). Central compiles that from the app model and sends individual `POST /firewall/allow` frames.

**`coolify init` is intentionally not the rule store.** Bootstrap creates the empty allow chain. coold is the sole writer into it. Callers reach coold via two paths: (a) central Coolify over the coold-initiated outbound stream, (b) intra-mesh callers (`coolify firewall` CLI via SSH-bounce, other coolds, optional per-customer gateway) via coold's local REST API on wg0 mgmt IP.

#### Reboot persistence

Works the same pre- and post-coold because both use the same file format:

- `/etc/coolify/allow.rules` — filter-table fragment, `:COOLIFY-ALLOW` + `-A COOLIFY-ALLOW` lines only. Written atomically (`.tmp` + `mv`) on every rule change.
- `/etc/systemd/system/coolify-mesh-allow.service` — `Type=oneshot`, `After=coolify-mesh-fw.service`, `Wants=coolify-mesh-fw.service`. `ExecStart=iptables-restore --noflush /etc/coolify/allow.rules`. `--noflush` means only `COOLIFY-ALLOW` is populated; nothing else is disturbed.

coold owns the file: it rewrites `/etc/coolify/allow.rules` on every successful API mutate, keeping it in sync with the live kernel. The `coolify firewall` CLI never touches the file — it POSTs/DELETEs through coold and coold handles persistence + systemd unit install. One writer, one format.

#### Allow-rule lifecycle

For an allow `(srcIP, dstIP)`:
- Add ACCEPT to `COOLIFY-ALLOW` on the host that **owns dstIP** (where DROP would otherwise fire).
- For bidirectional traffic (e.g. TCP, ICMP echo+reply), add the reverse `(dstIP, srcIP)` on the host that owns srcIP. (Reply packets traverse THAT host's FORWARD chain when arriving back, and dst-side check fires there.)
- **One unidirectional allow = one rule on one host. One bidirectional allow = two rules on two hosts.**
- Conntrack ESTABLISHED early-accept (installed by bootstrap) handles in-flow follow-up packets — no need to add per-packet rules.

#### Persistence + scale model

Per-rule systemd dropins do NOT scale (1000 rules × `daemon-reload` + restart = minutes, fs clutter, audit nightmare). Instead, coold is a thin rule-applier backed by central:

```
coold service (per host)
  ├─ Snapshot file:  /etc/coolify/allow.rules   (flat iptables-save fragment)
  ├─ Boot:           systemd unit runs iptables-restore --noflush from file
  ├─ API mutate:     apply iptables -A/-D  →  regen snapshot via iptables-save
  └─ Reconcile:      central periodically diffs its DB vs coold's live
                     `iptables -S COOLIFY-ALLOW`; pushes deltas to re-converge
```

Source of truth for **the set of rules that should exist** = central Coolify DB. Source of truth for **what's programmed in the kernel right now** = kernel itself, mirrored to `/etc/coolify/allow.rules` for reboot. coold does not keep its own DB.

#### Write ordering (crash/reboot safety)

Every mutating call from central → coold follows this sequence:

1. **Central writes to its own DB first** (with its own audit/tenant metadata). Durable with the rest of Coolify's state.
2. **Central sends command over the open stream** to coold with just `(src, dst, proto, port)`. No inbound connection to coold — the stream was already established by coold at boot.
3. **coold applies `iptables -A/-D`** to kernel.
4. **coold regenerates `/etc/coolify/allow.rules`** via `iptables-save` (atomic `.tmp` + `mv`).
5. **coold returns success to central** over the same stream (response carries the request id).
6. **On any failure in 3–5**, central marks the row "pending" in its DB and retries / surfaces to operator. Nothing is lost because step 1 is already durable.

Consequences:
- **Crash between steps 3 and 4** → kernel has the rule, file doesn't. Reboot loses the rule. Central's reconcile loop detects divergence (its DB has the rule, live kernel doesn't after boot) and re-pushes. Safe, with a small drift window bounded by reconcile cadence.
- **Crash between steps 4 and 5** → kernel + file both updated, but central didn't get the ack. Central retries; `iptables -C` guard makes the retry a no-op. Safe.
- **coold down when central wants to mutate** → central queues the change and retries on reconnect. No state loss on either side.
- **Central DB is authoritative** — a reboot can only *shrink* the live rule set compared to central's view, never grow it.

Bulk ops (`/bulk`) ship the whole batch in one REST call. coold applies via `iptables-restore --noflush` / `nft -f` (atomic transaction), then regens snapshot once.

Apply paths:

| Backend | Bulk apply (1000 rules) | Atomicity |
|---|---|---|
| `iptables -A` per rule | ~5s | per-rule |
| `iptables-restore --noflush` (preferred for iptables-legacy) | ~50ms | per-batch |
| `nft -f /tmp/rules.nft` (preferred when host uses nftables backend) | ~10ms | atomic transaction |

coold detects backend (`iptables --version` or presence of nftables socket) and picks. Bootstrap doesn't care.

For **systemctl restart coolify-mesh-fw.service** (e.g. a `coolify init bootstrap` re-run after a flag flip, or `coolify init extend` reinstalling the unit because the namespace list changed): the unit flushes COOLIFY-INTRA but **never flushes COOLIFY-ALLOW** — existing rules survive. If somehow lost (manual `iptables -F COOLIFY-ALLOW`, crash mid-write), central's reconcile loop compares its own DB against `iptables -S COOLIFY-ALLOW` from each host and re-pushes any missing tuples within the reconcile interval.

#### Allow API surface

Same method/path set is served on both transports — stream (central → coold) and local REST (intra-mesh → coold). Stream = JSON-RPC frames carrying the same `(method, path, body)` tuple; REST = plain HTTP on wg0 mgmt IP :8443.

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

#### API surface

Same dual-transport model as the firewall API — stream from central, REST from intra-mesh callers.

```
POST   /api/v1/services/register      {service_id, app_id, name, namespace, port, container_id, container_ip, host_mgmt_ip}
DELETE /api/v1/services/{service_id}/endpoints/{container_id}
GET    /api/v1/services/{service_id}/endpoints
GET    /api/v1/services?namespace=myapp
GET    /api/v1/dns/lookup/{name}      (debug — what coold would answer)
GET    /api/v1/dns/stats              (qps, hit/miss/forward counts)
```

Most ops are automatic side effects of deploy/scale/health-check. Central rarely calls `/services/register` directly — coold registers on container create, deregisters on remove.

coold writes Corrosion rows on behalf of central (explicit `POST /services/register` frames); it does not infer service identity from container labels. Central supplies `service_id` explicitly so naming policy stays in one place.

#### Bootstrap impact

Minimal. `coolify init bootstrap` creates every `coolify-<ns>-mesh` Podman network with `--disable-dns` so netavark never starts aardvark-dns on the bridge gateway `:53`. coold owns that socket. Bridge gateway IP was always reserved by `MachineIP()`.

Pre-alpha deployments that created the network without `--disable-dns` are detected at plan-time (probe reads `podman network inspect .DNSEnabled`). A `recreate-podman-network` action drops and recreates the network — same subnet, same gateway, but with DNS disabled. Any attached containers are disconnected via `podman network rm -f`.

#### Port 53 conflict handling

Three layers protect coold's `10.210.X.1:53` socket:

| Layer | Mechanism | Covers |
|---|---|---|
| 1. Bootstrap | `podman network create --disable-dns` (+ drift recreate) | aardvark-dns squat |
| 2. Bind target | coold binds **bridge gateway IP only**, not `0.0.0.0` and not wg0 mgmt IP | host wildcard DNS daemons (dnsmasq/pihole on `0.0.0.0:53`) and wg0 bloat |
| 3. Preflight | `net.Listen("tcp", gateway+":53")` probe before `ListenPacket` | clear actionable error + systemd `Restart=on-failure` retry |

systemd-resolved on Ubuntu binds `127.0.0.53:53` — no conflict with bridge gateway.

Bind rule: coold DNS is container-facing only (listen on bridge gateway IP). coold REST API is operator-facing (listen on wg0 mgmt IP, port 8443). Separate concerns, separate sockets.

### 6. Ingress (public traffic → containers)

`coolify init` doesn't manage public ingress. v5 deploys a reverse proxy (Traefik/Caddy) per host or HA pair:
- Listens on host public IP `:80/:443`.
- Routes `Host: app.example.com` → container IP (over container bridge or wg0 if cross-host).
- Cert management via ACME.
- Coolify generates proxy config from app routing rules.

Important: ingress proxy needs its own podman network OR can share `coolify-mesh`. Sharing means proxy can reach all containers — fine since it's the entrypoint.

### 7. Deployment workflows

Deploy is a **central-side state machine** that compiles app intent (compose / Dockerfile / buildpack / Nixpacks / raw image) into a sequence of coold primitives (see §2 wire surface). coold does not participate in planning — it executes one primitive per frame.

#### Build pipeline (not in coold)

```
git push
   │
   ▼
Central receives webhook
   │
   ▼
Builder (BuildKit / Buildpacks / Nixpacks)             ← coold NOT involved
  - Self-hosted: first mesh host by default;
    central may pin via target_host_id per build.
  - Cloud: central-run.
   │
   ▼
Push to registry (registry.coolify.io or customer's)   ← coold NOT involved
   │
   ▼
Central deploy controller → primitive op stream → coold on target host
```

coold's only role in the build path: `POST /images/pull` once the tag exists in the registry.

#### Deploy flow (T0–T10 — every frame = one §2 primitive)

```
T0  Central builder clones source, invokes BuildKit / buildpack / nixpacks.
    Output: OCI image @ registry.coolify.io/tenant/web:v2.

T1  Central deploy controller picks target host H (scheduler = least-loaded / pin).

T2  Frame: POST /images/pull {ref: "registry.coolify.io/tenant/web:v2"}
    coold@H calls podman.sock /images/create, streams progress back.

T3  Frame: POST /volumes {name: "web-data", driver: "local"}
    coold@H idempotent; no-op if exists.

T4  Frame: POST /containers  (central templates from compose + resolved secrets)
    body:
      {
        "image": "registry.coolify.io/tenant/web:v2",
        "name": "web-v2-a3f91",
        "network": "coolify-mesh",
        "ip": "10.210.H.42",
        "dns": ["10.210.H.1"],
        "dns_search": ["coolify.internal"],
        "env": {"DATABASE_URL": "postgres://…"},
        "mounts": [{"volume": "web-data", "target": "/data"}],
        "healthcheck": {"test": ["CMD","curl","-f","http://localhost/"], "interval": "5s"},
        "labels": {"coolify.app": "web", "coolify.version": "v2"}
      }
    coold checks deny filter → calls podman.sock /containers/create → returns id.

T5  Frame: POST /containers/{id}/start
    coold starts container.

T6  Central polls GET /containers/{id} or subscribes to events.
    Wait for healthy; abort + rollback on timeout.

T7  Frame: POST /services/register
    coold writes Corrosion row. Gossip distributes; DNS now answers new IP.

T8  Frame: POST /firewall/allow  (on dst host — coold = sole kernel writer)
    {src: proxy-ip, dst: 10.210.H.42, proto: "tcp", port: 80}

T9  Central ingress controller regenerates proxy config (Caddy/Traefik/nginx)
    → upstreams point to new container IP.
    Frame: POST /containers/{proxy-id}/exec (reload)  or proxy-specific reload.

T10 Cutover complete. Central retires the old container:
      POST /containers/{old-id}/stop {timeout: 10}
      DELETE /containers/{old-id}
      DELETE /services/web/endpoints/{old-container-id}
      DELETE /firewall/allow/{old-rule-id}
```

Every T-frame is one of the narrow primitives in §2. coold never runs compose, never builds, never picks hosts, never reads app config. If a future verb is needed, it gets added to §2 and the coold release, not smuggled through a passthrough.

**coold non-goals for deploy**: no compose parser, no buildpacks, no Dockerfile handler, no Nixpacks, no scheduler, no ingress templating, no rollback orchestration, no secrets store.

### 8. Storage & volumes

- Local podman volumes per host (`/var/lib/containers/storage/volumes`).
- Cross-host: distributed FS (out of scope) OR pin stateful services to a host (anti-affinity rules in scheduler).
- Backup: `podman volume export` + scp to backup target. Coolify orchestrates schedule.
- **v5 alpha decision**: stateful services **pin to host**. Cross-host volume movement / distributed FS is post-alpha.

### 9. Scheduling

**Placement lives in central.** coold provides facts (`GET /host/info`, `/host/stats`, `/host/containers`); central consumes them, picks the target host, and sends the resulting primitives. coold has no placement logic.

When user creates an app, central decides which host runs it:
- Round-robin / least-loaded / explicit pin.
- Pinned services (DB, persistent volumes) tracked in central DB.
- Re-schedule on host failure (wg0 down, last-handshake stale).

Failure detection: central polls `wg show wg0 latest-handshakes` via `GET /host/info` on every host, parses seconds-since-handshake; alerts if > N seconds.

### 10. Observability

coold exposes read-only `/host/*` endpoints surfacing the facts below. Central (or a central-side scraper) pulls from each host and feeds Prometheus / VictoriaMetrics. coold does **not** push metrics.

Per host metrics (over wg0 via coold endpoints):
- `GET /host/info` → podman info (version, storage driver, free space), kernel, wg state, load.
- `GET /host/containers` → `podman ps -a --format json` state.
- `GET /host/stats` → `podman stats --no-stream --format json` CPU/mem per container.
- Wg handshake + transfer bytes via `GET /host/info` (`wg show wg0 dump` internally).
- `iptables -nvL COOLIFY-ALLOW` match counters (for audit) exposed through `GET /firewall/allow` with counters.

Stream into central time-series store (Prometheus / VictoriaMetrics).

### 11. Updates

- Coolify runtime image self-updates (container restart with new image).
- WireGuard / Podman package updates: `coolify init bootstrap` re-runs idempotently and picks up newer packages from apt. Agent (coold/corrosion/scheduler/builder) bumps go through `coolify init upgrade --coold-version vX.Y.Z` etc. Schedule periodic re-apply (weekly?).
- Mesh config changes (new host, removed host) trigger re-apply on all hosts; control plane orchestrates.

### 12. Security posture

- **Private keys never leave hosts**: WG private key generated on remote, never transits SSH (already done by bootstrap).
- **Podman socket access**: `/run/podman/podman.sock` stays as a rootful Unix socket on each host — **NEVER exposed on TCP**. Only **coold** (per-host agent, see §2) has access via bind-mount. coold surfaces a curated REST API over wg0 with TLS + bearer auth. This means:
  - Compromise of a non-coold container does NOT grant podman API access.
  - coold enforces bearer-token authn and can deny dangerous flags (e.g. `--privileged`) at the API surface. RBAC, per-user/tenant scoping, and business audit live **only** in central Coolify (see §3 split).
  - No `podman system service tcp://...` listener; no need for socket-level TLS.
  - Central Coolify only knows the coold endpoint, not the underlying socket.
- **SSH access**: bootstrap uses key-based SSH. Control plane should rotate SSH keys per agent install, store in encrypted DB. After bootstrap, day-to-day ops go via coold REST — SSH is for re-bootstrap only.
- **Host firewall (iptables INPUT chain)**: bootstrap doesn't lock down INPUT. v5 should drop public access to ports other than `:51820/udp` (WG), `:22/tcp` (SSH), `:80/:443` (ingress). coold's `:8443` binds to the wg0 IP only, so it's already not on the public interface.
- **coold port reachability**: central never dials in — coold's outbound stream is the control path — so no `COOLIFY-ALLOW` rule for central is needed. coold's local REST on wg0 mgmt IP (`:8443`) is reachable only from inside the mesh, and is used by (a) the `coolify firewall` CLI via SSH-bounce, (b) other coolds in the same mesh, (c) an optional per-customer gateway. Nothing on the public internet reaches coold. Outbound TLS :443 to central must be permitted by the customer's egress firewall — standard for any SaaS agent.
- **Audit**: central Coolify is the sole authoritative audit log — who-when-why metadata for every COOLIFY-ALLOW change. coold writes only an ops/debug request log (request id, endpoint, status, duration) for troubleshooting; it never sees the identity of the human caller, only the bearer token used to reach it.

### 13. Failure modes & recovery

| Failure | Detection | Recovery |
|---|---|---|
| Host SSH unreachable | bootstrap apply error | Manual investigation; node marked unhealthy in DB |
| WG peer offline (`latest_handshake > 180s`) | `wg show` poll | Mark unhealthy; re-schedule containers if pinning permits |
| Podman socket unreachable | API call timeout | Restart `podman.socket`; if persistent, re-bootstrap |
| Firewall service failed | `systemctl is-active != active` | Re-run `coolify init bootstrap`; service is idempotent |
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
coolify deploy <app>                                      # build + push + run
coolify scale <app> --replicas N
coolify firewall containers --servers A,B ...             # discover mesh containers (SSH+podman)
coolify firewall list --servers A,B ...                   # list allow rules across hosts (coold GET /allow, SSH-bounced)
coolify firewall allow --from <ref> --to <ref> --port N   # add allow rule (coold POST /allow, SSH-bounced)
coolify firewall revoke --from <ref> --to <ref> --port N  # remove allow rule (coold DELETE /allow/{id})
coolify host list                                         # show mesh state, last-handshake, container count
coolify host add <ip> --ssh-key K
coolify host remove <ip>
coolify logs <container>
coolify exec <container> -- sh
```

`coolify firewall` is implemented today as a thin SSH-bounced REST client of coold (§3 above). The laptop running the CLI isn't a mesh peer, so every call SSHes into the target host and runs `curl "http://<wg0-mgmt-ip>:8443/api/v1/firewall/..."` against coold locally. Per-host bearer tokens are fetched from `/etc/coolify/api-token` on demand (with `--coold-token` as an override for homogeneous test clusters).

Everything else on the roadmap (`coolify deploy`, `coolify scale`, `coolify logs`, `coolify exec`) targets the **central** API (SaaS or self-hosted central), not coold directly. Central compiles the request into the primitive-op sequence in §7 and streams it to coold. Only `coolify firewall` currently bypasses central and hits coold directly — legacy + test harness until central wires up `/firewall/*` itself.

---

## Summary

`coolify init bootstrap` does the **first-time host install**: WG mesh, podman runtime, bridge network, default-deny scaffold, coold/corrosion/scheduler/builder agents. `coolify init extend` adds hosts to an existing mesh without disturbing converged ones; `coolify init upgrade` bumps agent versions across the fleet. After that, **everything dynamic is the v5 control plane's job**: container lifecycle, allow rules in COOLIFY-ALLOW (via systemd dropins for persistence), scheduling, observability, ingress, updates.

The pieces communicate via:
1. **SSH** for host provisioning + re-converge (idempotent `coolify init bootstrap` / `extend` / `upgrade` re-runs). SSH is the installer channel only, not a steady-state control path.
2. **coold → central outbound stream** (WSS / gRPC bidi on :443) for day-to-day runtime ops from central. One topology for self-hosted and cloud SaaS; central never dials coold, never joins any mesh. Per-customer gateway (optional) collapses N streams into 1 per mesh.
3. **coold local REST API** on wg0 mgmt IP (`http://100.64.0.X:8443`) for intra-mesh callers: the `coolify firewall` CLI via SSH-bounce, other coolds, the per-customer gateway. Never reachable from the public internet.

coold is the *only* process with access to the local podman socket AND the sole writer of allow rules in COOLIFY-ALLOW. Both transports hit the same API surface.

Persistence model:
- Bootstrap state (chains, jumps, conntrack accept) → idempotent `coolify init bootstrap` re-runs (and `extend` when a namespace is added).
- Rule metadata (who/when/why, audit, RBAC, tenant scoping) → central Coolify DB only. coold does not duplicate this.
- Kernel rules → programmed by coold on every API call (from either central Coolify or the `coolify firewall` CLI); mirrored to `/etc/coolify/allow.rules` for reboot via `coolify-mesh-allow.service` (oneshot `iptables-restore --noflush`).
- Today the `coolify firewall` CLI is the primary caller of coold (SSH-bounced REST client with per-host `/etc/coolify/api-token` resolution). Central Coolify will call the same API once wired.

The podman socket is host-local. There is no TCP podman API. coold is the **authn + privilege boundary** between any caller (central Coolify over the outbound stream, or the `coolify firewall` CLI via SSH-bounced local REST) and the host, AND the kernel-rule applier. Central Coolify owns RBAC, tenant scoping, and the business audit log (who/when/why). coold only verifies a bearer token (per-host static for local REST; per-host JWT for the stream), applies the rule, and keeps an ops/debug request log. `coolify firewall` exercises the local REST surface today; central will exercise the stream surface — same code path end-to-end, different transport.

**coold stays small.** All app-aware logic (compose, Dockerfile, buildpacks, Nixpacks, scheduling, rollback, ingress templating, RBAC, audit) lives in central. coold's wire surface is enumerable (§2); new verbs require a coold release, not a `/podman/raw` passthrough. If coold ever grows a `/apps` or `/compose` endpoint, that is the wrong layer.

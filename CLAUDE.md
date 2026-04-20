# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a CLI tool for interacting with the Coolify API, built with Go using the Cobra framework. The CLI allows users to manage Coolify instances (both cloud and self-hosted), servers, projects, resources, deployments, domains, and private keys.

### API Specification
This CLI is a client for the Coolify API. The API specification is defined in the OpenAPI schema:
- **Source**: https://github.com/coollabsio/coolify/blob/v4.x/openapi.json
- **Raw JSON**: https://raw.githubusercontent.com/coollabsio/coolify/refs/heads/v4.x/openapi.json
- **Base Path**: `/api/v1/`
- **Authentication**: Bearer token (API tokens from Coolify dashboard at `/security/api-tokens`)

All commands in this CLI are wrappers around API endpoints defined in the OpenAPI specification. When adding new features or endpoints:
1. Check the OpenAPI spec for available endpoints and their request/response schemas
2. Ensure the CLI command structure follows the API resource hierarchy
3. Match the API's data types and validation rules

## Architecture

### Command Structure
The codebase follows Cobra's command pattern with a root command and subcommands:
- Entry point: `coolify/main.go` calls `cmd.Execute()`
- Root command: `cmd/root.go` - contains core utilities (HTTP client, authentication, version checking, config management)
- Subcommands: Each command is in its own file in `cmd/`:
  - `context.go` - manage Coolify context (add, remove, list, set default/token)
  - `servers.go` - list and get server information
  - `projects.go` - list projects with environments and applications
  - `resources.go` - list resources
  - `deploy.go` - deploy resources
  - `domains.go` - manage domains
  - `privatekeys.go` - manage SSH keys
  - `update.go` - self-update CLI
  - `version.go` - show CLI version

### Configuration Management
- Uses Viper for configuration management
- Config file location: `~/.config/coolify/config.json` (via xdg package)
- Config stores multiple instances with tokens, default instance selection
- Global flags available: `--token`, `--host`, `--format`, `--show-sensitive`, `--force`, `--debug`

### API Communication
Core API functions in `cmd/root.go`:
- `Fetch(url string)` - GET requests
- `Post(url, input)` - POST requests
- `Delete(url)` - DELETE requests
All API calls use `Fqdn + "/api/v1/" + url` pattern with Bearer token authentication

### Version Management
- CLI version tracking with auto-update check (10 minute interval)
- API version checking and minimum version enforcement via `CheckMinimumVersion()`
- Self-update capability using `go-selfupdate` library

### Output Formatting
Three output modes supported via `--format` flag:
- `table` (default) - tabwriter formatted output
- `json` - compact JSON
- `pretty` - indented JSON

## Development Commands

### Build
```bash
go build -o coolify ./coolify
```

### Run locally
```bash
go run ./coolify [command]
```

### Test a command
```bash
go run ./coolify context list
go run ./coolify servers list --debug
```

### Install locally
```bash
go install ./coolify
```

### Run tests
```bash
# Run all tests (tests are in internal/ directory)
go test ./internal/...

# Run with coverage
go test ./internal/... -cover

# Run with verbose output
go test ./internal/... -v

# Run specific package
go test ./internal/api/... -v
go test ./internal/service/... -v

# Run specific test
go test ./internal/api -run TestClient_Get_Success -v
```

### Before committing
```bash
# 1. Run tests
go test ./internal/...

# 2. Check coverage
go test ./internal/... -cover

# 3. Run linter (if available)
golangci-lint run

# 4. Format code
go fmt ./...
```

## Release Process

- Uses GoReleaser for multi-platform builds (Linux, Darwin, Windows on amd64/arm64)
- Release workflow: `.github/workflows/release-cli.yml` triggers on GitHub releases
- GoReleaser config: `.goreleaser.yml`
- Install script: `scripts/install.sh` downloads from GitHub releases

## Key Patterns

### Adding a New Command
1. Create new file in `cmd/` (e.g., `cmd/newfeature.go`)
2. Define command struct with cobra.Command
3. Implement Run function with:
   - Call `CheckDefaultThings(nil)` to validate version and format
   - Use `Fetch()`, `Post()`, or `Delete()` helpers
   - Handle JSON unmarshaling into typed structs
   - Support all three output formats
4. Register command in `init()` function: `rootCmd.AddCommand(yourCmd)`

### API Version Requirements
If a command requires a specific Coolify API version, pass it to `CheckDefaultThings()`:
```go
minimumVersion := "4.0.0"
CheckDefaultThings(&minimumVersion)
```

### Handling Sensitive Data
- Use `ShowSensitive` flag to control display of tokens/secrets
- Default overlay: `SensitiveInformationOverlay = "********"`

### UUID vs ID Pattern
**CRITICAL: Always use UUIDs for user-facing interactions, never internal database IDs.**

When adding new commands or models:
1. **Command Arguments**: Always accept UUIDs as string arguments (e.g., `<resource_uuid>`), never integer IDs
2. **API Endpoints**: Construct API paths using UUIDs (e.g., `resources/{uuid}`), not IDs
3. **Service Layer**: Methods should accept `uuid string` parameters, not `id int`
4. **Table Output**: Hide internal IDs from table output using `table:"-"` struct tags
5. **Model Fields**:
   - Keep `ID int` field with `json:"id" table:"-"` (for API responses, hidden from users)
   - Always include `UUID string` field with `json:"uuid"` (visible to users)

**Example model:**
```go
type Resource struct {
    ID   int    `json:"id" table:"-"`     // Hidden from table output
    UUID string `json:"uuid"`              // Shown in table output
    Name string `json:"name"`
    // ... other fields
}
```

**Why UUIDs?**
- UUIDs are stable across environments (dev, staging, prod)
- IDs are internal implementation details that can change
- UUIDs are more secure (don't expose database sequencing)
- Coolify API uses UUIDs as the primary resource identifier

## `coolify init` — WireGuard mesh + Podman bootstrap (alpha, v5)

**This subcommand is an outlier**: it does NOT talk to the Coolify API. It SSHes into remote hosts and installs/configures WireGuard, Podman, the bridge network, and a firewall scaffold. It's a one-shot infrastructure bootstrap consumed by the future v5 control plane (coold).

### What it does

- Establishes a full-mesh WireGuard overlay across N hosts.
- Each host gets a mgmt IP `/32` from `--wg-mgmt-pool` (default `100.64.0.0/16`, RFC 6598 CGNAT) on `wg0`.
- Each host gets a container subnet `/<container-prefix>` from `--container-pool` (default `10.210.0.0/16`, default prefix `/24`) owned by a Podman bridge named `coolify-mesh`.
- Optionally installs Podman + enables `podman.socket` + creates the bridge + installs `coolify-mesh-fw.service` (`--podman` flag).
- Optionally installs default-deny firewall scaffold (`--default-deny` flag) — adds `COOLIFY-INTRA` and empty `COOLIFY-ALLOW` chains.

### Architecture (why this layout)

The mgmt pool and container pool are **separate** so the Podman bridge can own the full container `/24` without conflicting with `wg0`. Pattern adopted from uncloud (psviderski/uncloud).

WG config per host (e.g. host A):
```
[Interface]
Address    = 100.64.0.1/32      # mgmt IP, NOT in container pool
ListenPort = 51820
PrivateKey = <gen on host>

[Peer]                          # one per other host
PublicKey  = <peer pubkey>
AllowedIPs = 100.64.0.2/32, 10.210.1.0/24    # peer mgmt + peer container subnet
Endpoint   = <peer SSH ip>:51820
```

Critical: `AllowedIPs` lists the peer's full `/24` so kernel routes `10.210.<peer>.0/24 dev wg0`. This is what makes cross-host container traffic work.

Podman network `coolify-mesh` is created with `--disable-dns` — bridge gateway `10.210.X.1:53` is reserved for coold's embedded cluster DNS (see CONTROL_PLANE.md §5). Pre-alpha networks with `dns_enabled=true` are detected on re-run and recreated.

Firewall service (`coolify-mesh-fw.service`) installed by `--podman`:
- POSTROUTING `RETURN` rule prevents Podman MASQUERADE from rewriting container egress source on `wg0` (would break reverse routing because wg0 has no IP in the container subnet).
- Mode A (no `--default-deny`): blanket FORWARD ACCEPT for container subnet.
- Mode B (`--default-deny`): COOLIFY-INTRA chain (ESTABLISHED accept → COOLIFY-ALLOW → DROP), FORWARD jumps for `-s/-d <container-subnet>`. v5 control plane fills `COOLIFY-ALLOW`.

### Cross-host vs intra-host firewall

- **Cross-host default-deny WORKS** — those packets cross interfaces (wg0 ↔ bridge) and traverse iptables FORWARD. Empirically verified.
- **Intra-host (same bridge) is NOT enforced** — Linux + netavark + Ubuntu 24.04 quirk: bridge L2 traffic bypasses iptables FORWARD even with `bridge-nf-call-iptables=1`. v5 control plane handles intra-host isolation via per-app podman networks (`--opt isolate=true`), not iptables.

### Subcommands

```bash
coolify init plan   --servers IP1,IP2 --ssh-key ~/.ssh/id_ed25519 [--podman --default-deny]
coolify init apply  --servers IP1,IP2 --ssh-key ~/.ssh/id_ed25519 [--podman --default-deny] [--yes]
```

- `plan` is read-only: SSH-probes each host, reconstructs current state, shows what `apply` would do. Idempotent.
- `apply` is the same plus execution. 2-phase parallel: phase 1 = install + keygen + podman + socket + IP forward. Re-probe to collect fresh public keys. Phase 2 = write WG config + enable/reload service + create podman network + install firewall.

### Flags (defined in `cmd/init/flags.go`)

| Flag | Default | Purpose |
|---|---|---|
| `--servers` | required | comma-separated SSH IPs |
| `--ssh-key` | required | path to SSH private key |
| `--ssh-passphrase-prompt` | false | prompt for key passphrase (also reads `COOLIFY_SSH_PASSPHRASE` env) |
| `--ssh-user` | `root` | SSH user |
| `--ssh-port` | `22` | SSH port |
| `--wg-mgmt-pool` | `100.64.0.0/16` | mgmt IP pool, /32 per host on wg0 |
| `--container-pool` | `10.210.0.0/16` | container pool, carved per host |
| `--container-prefix` | `24` | per-host container subnet prefix |
| `--wg-interface` | `wg0` | WG iface name on remote |
| `--wg-listen-port` | `51820` | WG UDP port |
| `--podman` | false | install podman + bridge + firewall service |
| `--podman-network` | `coolify-mesh` | bridge network name |
| `--default-deny` | false | requires `--podman`. Install COOLIFY-INTRA + empty COOLIFY-ALLOW chains for cross-host deny |
| `--concurrency` | `10` | parallel SSH connections |
| `--ssh-timeout` | `30s` | SSH connect timeout |
| `--yes`, `-y` | false | skip alpha confirmation prompt |

### Code layout

- `cmd/init/` — Cobra subcommands (`init`, `init plan`, `init apply`).
  - `flags.go` — `InitFlags` struct + bindings + SSH client builder.
  - `plan.go` — `runPlan`: parse pools, build SSH client, probe, plan, render.
  - `apply.go` — `runApply`: alpha gate, probe, plan, execute, verify.
  - `init.go` — registers subcommands; package is `initcmd` (not `init` — Go reserved keyword).
- `internal/wireguard/` — pure Go logic (no SSH, no I/O — `apply.go` is the SSH boundary).
  - `state.go` — `ServerState`, `MeshState`, `DesiredMesh` types.
  - `subnet.go` — `Allocate` (per-host subnets) + `AllocateMgmtIPs` (per-host /32) + conflict detection. Skips `.0` and broadcast for /32. Stable reuse + dedup-host check + warn-on-conflict.
  - `config.go` — `RenderConfig` + `WriteConfigCommand` for `wg0.conf` (Address /32, AllowedIPs mgmt + container).
  - `reconstruct.go` — `Probe` (SSH probes) + `Reconstruct` (parallel) + `parseConfigFile`.
  - `plan.go` — `BuildPlan` (pure function: desired - actual = actions). `ActionType` enum.
  - `apply.go` — `ApplyMesh` (2-phase fanout via `internal/ssh/fanout.go`). `runStep` helper.
  - `firewall.go` — `coolify-mesh-fw.service` unit generator (two-mode: blanket allow vs default-deny).
- `internal/ssh/` — generic SSH runner + parallel `ForEachServer[T]`.
- `test/fixtures/wg/wg0.conf` — fixture for parser tests.

### Key invariants

- **Reconstructed-only state**: no local state file. Every `plan`/`apply` re-probes via SSH. State lives on the hosts.
- **Idempotent**: re-running with no changes produces empty plan. State drift triggers re-converge (e.g. flipping `--default-deny` reinstalls the firewall service).
- **Private key never leaves host**: WG private key generated on remote via `wg genkey`; config written using `$PRIVKEY=$(cat /etc/wireguard/privatekey)` shell expansion.
- **Atomic config writes**: write to `.conf.tmp`, `mv` to `.conf`.
- **Stable subnet assignment**: existing valid assignments are preserved across re-runs; only invalid (out-of-pool, wrong prefix, duplicate, network/broadcast IP) trigger reassignment with warning.

### Future control plane (v5 / coold)

`coolify init` only does the **one-shot host bootstrap**. Day-to-day container/firewall ops are the v5 control plane's job. See `CONTROL_PLANE.md` for the full spec, including:

- coold per-host agent (REST API on wg0, bind-mounts `/run/podman/podman.sock`, NEVER exposes socket on TCP).
- Service discovery via embedded DNS in coold + Corrosion-replicated sqlite (no env injection, no container restart on backend movement).
- Allow-rule persistence via coold's own DB + `iptables-restore --noflush` or `nft -f` batch (NOT systemd dropins per rule — doesn't scale).
- Cross-host allow rules go on the **destination host** (where DROP would otherwise fire).

When extending `coolify init`, defer dynamic responsibilities to coold. Bootstrap should stay narrow: scaffold the mesh, install runtime, prep firewall chains. coold owns everything that changes at runtime.

### Testing init

Tests live in `internal/wireguard/*_test.go` and `cmd/init/*_test.go`:

```bash
go test ./internal/wireguard/... ./cmd/init/... -v
```

Use the SSH `Runner` interface for mocking — never open real SSH connections in unit tests. `internal/ssh/fanout.go` is generic; reuse for any per-server fanout.

## `coolify firewall` — cross-host allow-rule client (alpha, v5)

**This subcommand is the second outlier** (alongside `coolify init`): it does NOT talk to the Coolify API. It is a thin REST client of the **coold** per-host agent installed by `coolify init --install-coold`. `allow` / `revoke` / `list` all go through coold's REST API (`/api/v1/firewall/allow`). `containers` stays SSH+podman because coold has no container surface yet. Transport is **SSH-bounce**: the laptop running the CLI is not a mesh peer, so it SSHes into the target host and the shell there runs `curl "http://$(wg0-mgmt-ip):8443/api/v1/firewall/..."` against coold on localhost.

coold owns all kernel-rule + persistence logic (iptables/nft backend detection, `/etc/coolify/allow.rules` snapshot, `coolify-mesh-allow.service`). The CLI never writes iptables or systemd units directly.

### What it does

- Discovers containers on the `coolify-mesh` bridge across all listed hosts (SSH + `podman ps`).
- `POST /api/v1/firewall/allow` / `DELETE /api/v1/firewall/allow/{id}` / `GET /api/v1/firewall/allow` against coold on the host that **owns the destination IP** (per `CONTROL_PLANE.md §3`: rules go on dst host).
- Per-host bearer tokens fetched on demand from `/etc/coolify/api-token` (see `EnsureCooldAPITokenCommand` in `internal/services/coold.go` — each host generates its own random 32-byte hex token at install time).
- Idempotent at the coold level: POST of an identical tuple returns the existing id; DELETE of an unknown id returns 204.

### Subcommands

```bash
coolify firewall containers   # discover containers on coolify-mesh across hosts (SSH+podman)
coolify firewall list         # GET /allow on every host and merge
coolify firewall allow --from <ref> --to <ref> [--port N] [--proto tcp|udp] [--bidirectional]
coolify firewall revoke --from <ref> --to <ref> [--port N] [--proto tcp|udp] [--bidirectional]
```

`<ref>` accepts: container name (unique across mesh), `host:name`, short 12-char podman ID, or raw IP.

### Flags

Persistent (inherited from `cmd/common/sshmesh.go` — shared with `coolify init`):

| Flag | Default | Purpose |
|---|---|---|
| `--servers` | required | comma-separated SSH IPs |
| `--ssh-key` | required | SSH private key path |
| `--ssh-passphrase-prompt` | false | prompt for passphrase (also `COOLIFY_SSH_PASSPHRASE` env) |
| `--ssh-user` | `root` | SSH user |
| `--ssh-port` | `22` | SSH port |
| `--concurrency` | `10` | parallel SSH connections |
| `--ssh-timeout` | `30s` | SSH connect timeout |

Firewall-specific persistent:

| Flag | Default | Purpose |
|---|---|---|
| `--podman-network` | `coolify-mesh` | bridge network name (must match `coolify init --podman-network`) — used by `containers` discovery |
| `--coold-port` | `8443` | TCP port coold's REST API listens on (wg0 mgmt IP). Must match `COOLD_API_BIND` emitted by `internal/services/coold.go` |
| `--coold-token` | `""` | **optional** bearer-token override (also reads `COOLIFY_COOLD_TOKEN` env). When empty (the default), the CLI SSHes each host and reads `/etc/coolify/api-token` — tokens are per-host, not centrally shared |

Allow/revoke local:

| Flag | Default | Purpose |
|---|---|---|
| `--from` | required | source container ref or raw IP |
| `--to` | required | destination container ref or raw IP |
| `--port` | `0` | dst port (0 = any) |
| `--proto` | `tcp` | `tcp`, `udp`, or `""` (any — requires `--port=0`) |
| `--bidirectional` | false | also install reverse rule on src host (needed for server-initiated flows; conntrack ESTABLISHED handles client-initiated replies) |

### Rule identity

`cid = sha256(src|dst|proto|port)[:12]`. coold computes it server-side on POST and returns it in the body; the CLI surfaces it as the user-facing rule ID in `firewall list` output and uses it for DELETE. Stable across calls: `revoke --from … --to …` rebuilds the same cid and matches.

### SSH-bounce transport

Every coold call is wrapped in a single SSH command that first discovers the host's own wg0 mgmt IP and then curls coold on localhost:

```sh
# emitted for POST / DELETE (hard-fails if wg0 missing — no coold means nothing to apply to)
MGMT=$(ip -4 -o addr show wg0 2>/dev/null | awk '{print $4}' | cut -d/ -f1)
test -n "$MGMT" || { echo "coold mgmt IP (wg0) not found on $(hostname)" >&2; exit 1; }
curl -fsS --max-time 10 \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -X POST -d '{"src":"...","dst":"...","proto":"tcp","port":80}' \
  "http://$MGMT:8443/api/v1/firewall/allow"
```

`list` uses the **soft** variant: missing wg0 emits `[]` and exits 0 so a partially-deployed mesh doesn't abort the whole fanout.

### Per-host token resolution

`cmd/firewall/helpers.go::tokenResolver` hands out tokens per host with a sync.Mutex-guarded cache:

- `--coold-token` (or `COOLIFY_COOLD_TOKEN` env) set → closure returns the override for every host; no SSH fetch.
- Otherwise → first access per host SSHes `cat /etc/coolify/api-token`, caches the result for the rest of the run. Token-fetch failures surface as a `ServerResult.Err` on the owning host (won't poison others).

The cache is scoped to one CLI invocation — no on-disk caching.

### Persistence across reboots

**coold owns this now.** On every API mutate, coold regenerates `/etc/coolify/allow.rules` (flat `iptables-save` fragment) and the companion `coolify-mesh-allow.service` restores it on boot via `iptables-restore --noflush`. Pre-coold persistence scaffolding was removed from the CLI when it migrated to REST — same file format, different writer.

### Code layout

- `cmd/common/sshmesh.go` — shared SSH/mesh flag struct `SSHMeshFlags` + `BindSSHMeshFlags`, `BuildSSHClient`, `ParseSSHTimeout`, `ResolvePassphrase`, `Validate`. Embedded by both `cmd/init/InitFlags` and `cmd/firewall/FirewallFlags`.
- `cmd/firewall/` — Cobra layer.
  - `firewall.go` — `NewFirewallCommand()` parent + subcommand registration.
  - `flags.go` — `FirewallFlags` embeds `common.SSHMeshFlags` + `PodmanNetworkName` + `CooldToken` + `CooldPort`. `ResolveCooldToken()` returns the override or `""` (meaning "fetch per host").
  - `allow.go` — `allowRevokeFlags`, `emitAllowRevoke` (discover → resolve → build rule → coold POST/DELETE per rule, resolving token per host).
  - `list.go` — `emitList` fans out `CooldList` via `CooldListAll` using the per-host token resolver.
  - `containers.go` — `containers` subcommand (still SSH+podman; no coold dependency).
  - `resolve.go` — `resolveEndpoint(ref, []Container)` (name / host:name / short-id / raw IP).
  - `helpers.go` — `discoverAllViaPkg`, `tokenResolver` (per-host cached bearer-token closure).
- `internal/firewall/` — REST client + discovery.
  - `coold_client.go` — `FetchCooldToken`, `CooldApply`, `CooldRevoke`, `CooldList`, `CooldListAll`. `buildCurlAllow/Revoke/List`, `shellSingleQuote`, `mgmtIPScript` / `mgmtIPScriptSoft`. Consts `CooldAPIBasePath = "/api/v1/firewall"`, `CooldAPITokenPath = "/etc/coolify/api-token"`.
  - `discover.go` — `Container`, `discoverScript`, `DiscoverContainers`, `DiscoverAll` (parallel).
  - `rule.go` — `AllowRule`, `ComputeID`. Chain-rendering helpers were removed along with `apply.go` / `list.go` / `persist.go` (coold owns kernel + snapshot now).
- `internal/models/firewall.go` — table/JSON row types (`ContainerRow`, `AllowRuleRow`).
- `internal/services/coold.go` — `EnsureCooldAPITokenCommand` (installer writes `/etc/coolify/api-token`, mode 0600), `CooldServiceUnit` emits `COOLD_API_BIND=<mgmt-ip>:8443` + `COOLD_API_TOKEN_FILE=/etc/coolify/api-token`.

### Key invariants

- **Destination-host ownership**: every rule lives on exactly one host — the one whose `/24` contains the destination IP. `--bidirectional` adds the reverse rule on the src host.
- **coold is the only kernel writer**: the CLI never runs `iptables` or touches `/etc/coolify/allow.rules` directly. Everything flows through coold's REST API.
- **Per-host tokens by default**: each coold generates its own random token at install. `--coold-token` is an escape hatch for homogeneous test / CI environments, not the common path.
- **Bidirectional is opt-in**: conntrack ESTABLISHED accept (installed by `coolify-mesh-fw.service`) handles reply packets for client-initiated flows. Only set `--bidirectional` for protocols that actually open new connections in both directions.
- **Rule identity is hash, not UUID**: coold computes it server-side so CLI and any future writer agree on the same id for the same tuple.
- **Transient token exposure on remote `/proc`**: `curl -H "Authorization: Bearer $TOKEN"` is visible in `/proc/<curl-pid>/cmdline` for the ~ms lifetime of the call, root-only. Acceptable for alpha; TLS + stdin-fed tokens are a follow-up.

### Testing firewall

```bash
go test ./internal/firewall/... ./cmd/firewall/... ./cmd/common/... -v
```

Uses `fakeCooldRunner` / `cmdFakeRunner` pattern (substring → canned stdout map) — same as `cmd/init/plan_test.go`. All SSH calls mocked at the `ssh.Runner` boundary; no real SSH in unit tests. Token-fetch, mgmt-IP script, curl shape, JSON payload, and error propagation are all covered.

### End-to-end flow (verified on real hosts)

After `coolify init apply --podman --default-deny --install-coold --servers A,B ...` ran (coold must be up):

1. Baseline cross-host traffic DROPped by `COOLIFY-INTRA`.
2. `coolify firewall containers --servers A,B --ssh-key KEY` → discovery table.
3. `coolify firewall allow --servers A,B --ssh-key KEY --from client --to web --port 80` → CLI SSH-fetches each host's token, POSTs to coold on the dst host, traffic flows.
4. `coolify firewall list --servers A,B --ssh-key KEY` → merged rules from every host with their coold-assigned `cid:…` ID.
5. `coolify firewall revoke …` → coold DELETE, rule gone, traffic DROPped again.
6. Reboot → `coolify-mesh-allow.service` (installed by coold) restores from `/etc/coolify/allow.rules`.

Add `--coold-token <hex>` only when every host was bootstrapped with the same token (CI fixtures, homogeneous test clusters).

## Testing Requirements

**CRITICAL: All code changes MUST include tests. This is non-negotiable.**

### Test Coverage Requirements
- **Minimum coverage**: 70% for all packages
- **New features**: Must have 80%+ coverage
- **Bug fixes**: Must include regression tests
- **Refactoring**: Must maintain or improve existing coverage

### Testing Structure
```
test/
├── fixtures/           # Test data, mock API responses
├── mocks/             # Mock implementations of interfaces
└── integration/       # Integration tests with test server
```

### Test Requirements by Package Type

#### 1. Command Tests (`cmd/*_test.go`)
- Test command parsing and flag handling
- Test output formatting (table, json, pretty)
- Use mock API client to avoid real API calls
- Test error handling and validation
- Example:
```go
func TestServersListCmd(t *testing.T) {
    // Test with mock client
    // Verify output format
    // Test error cases
}
```

#### 2. API Client Tests (`internal/api/*_test.go`)
- Test request building
- Test response parsing
- Test error handling (4xx, 5xx status codes)
- Test retry logic
- Test timeout behavior
- **IMPORTANT**: Use `httptest.NewServer()` for mock HTTP responses (NOT real APIs)
- All API tests must use local mock servers, never call real Coolify cloud or external APIs

#### 3. Service Tests (`internal/service/*_test.go`)
- Test business logic
- Mock API client
- Test complex workflows
- Test error propagation

#### 4. Model Tests (`internal/models/*_test.go`)
- Test JSON marshaling/unmarshaling
- Test validation logic
- Test helper methods

#### 5. Integration Tests (`test/integration/*_test.go`)
- Test full command execution
- Test with real HTTP server (httptest)
- Test config file operations
- Test version checking
- Can be run with `-short` flag to skip

### Running Tests

```bash
# Run all tests (tests are in internal/ directory)
go test ./internal/...

# Run with coverage
go test ./internal/... -cover

# Generate coverage report
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run with verbose output
go test ./internal/... -v

# Run only unit tests (skip integration)
go test ./internal/... -short

# Run specific package
go test ./internal/api/... -v
go test ./internal/service/... -v
```

### Test Guidelines

1. **Table-driven tests**: Use for testing multiple scenarios
2. **Test naming**: `TestFunctionName_Scenario_ExpectedBehavior`
3. **Subtests**: Use `t.Run()` for related test cases
4. **Setup/Teardown**: Use `TestMain()` for package-level setup
5. **Parallel tests**: Use `t.Parallel()` when tests are independent
6. **Mock dependencies**: Never call real APIs in unit tests
7. **Test fixtures**: Store mock API responses in `test/fixtures/`

### Example Test Structure

```go
func TestServersList(t *testing.T) {
    tests := []struct {
        name       string
        response   string
        wantErr    bool
        wantCount  int
    }{
        {
            name:      "successful list",
            response:  readFixture("servers_list.json"),
            wantErr:   false,
            wantCount: 3,
        },
        {
            name:      "empty list",
            response:  "[]",
            wantErr:   false,
            wantCount: 0,
        },
        {
            name:      "api error",
            response:  `{"error":"unauthorized"}`,
            wantErr:   true,
            wantCount: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### When Adding a New Command

**CHECKLIST** (must complete ALL items):
- [ ] Create command implementation in `cmd/`
- [ ] Create corresponding test file in `internal/service/*_test.go` or `internal/api/*_test.go`
- [ ] Test all flags and arguments
- [ ] Test all output formats (table, json, pretty)
- [ ] Test error cases (missing args, API errors, invalid input)
- [ ] Add integration test if command has complex workflow
- [ ] Update README.md with command documentation
- [ ] Run `go test ./internal/...` and ensure all tests pass
- [ ] Verify coverage: `go test ./internal/... -cover`

### CI/CD Integration

Tests run automatically on:
- Every pull request
- Every commit to main branch
- Before releases

**Pull requests will be blocked if:**
- Any test fails
- Coverage drops below 70%
- New code has no tests

## .cursorrules Context

The project follows Go 1.22+ idioms with standard library preference:
- Use `net/http` standard library (no external HTTP frameworks)
- Leverage Go 1.22 ServeMux features for any routing needs
- Follow RESTful patterns for API interactions
- Implement proper error handling with custom types when needed
- Use Go's concurrency features appropriately
- Write secure, efficient, and maintainable code
- **ALWAYS write tests** - see Testing Requirements section above
# Proxmox MCP Server — Project Plan

## Summary

Build an MCP server in Go that exposes Proxmox VE cluster operations as MCP tools.
Uses the official `modelcontextprotocol/go-sdk`, a custom Proxmox HTTP client (no third-party
Proxmox library), API token auth, and strict linting enforced via `golangci-lint`, `gosec`,
`govulncheck`, and `go fix`.

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| MCP SDK | `github.com/modelcontextprotocol/go-sdk` | Official Go SDK |
| Proxmox client | Custom `net/http` wrapper | Full control, no external dep, easy to extend |
| Auth | API token only (`PVEAPIToken=`) | Stateless, no CSRF, no ticket renewal |
| Transports | Both stdio + HTTP (flag-selected) | stdio for local clients, HTTP for remote/shared |
| Linter | `golangci-lint` + `gosec` + `govulncheck` | Security-first, idiomatic Go |
| Formatter | `gofumpt` (stricter than `gofmt`) | Consistent style |
| TLS | `PROXMOX_INSECURE=true` opt-in for skip-verify | Proxmox uses self-signed certs by default |
| Error handling | Always wrap with `fmt.Errorf("doing X: %w", err)` | Idiomatic, stack-traceable |
| Global state | None — client injected, no `init()` | Testable, explicit |

## Project Structure

```
proxmox_mcp/
├── cmd/
│   └── proxmox-mcp/
│       └── main.go           # entrypoint, CLI flags, transport selection
├── internal/
│   └── proxmox/
│       ├── client.go         # custom HTTP client (auth, TLS, base URL)
│       ├── client_test.go    # httptest-based unit tests
│       ├── types.go          # shared Proxmox response/request structs
│       ├── cluster.go        # cluster-wide API calls
│       ├── nodes.go          # node-related API calls
│       ├── vms.go            # QEMU VM API calls
│       ├── containers.go     # LXC container API calls
│       ├── snapshots.go      # snapshot API calls
│       ├── tasks.go          # task polling (UPID → status)
│       ├── storage.go        # storage content API calls (Phase 3)
│       └── backup.go         # backup/vzdump API calls (Phase 3)
├── tools/
│   ├── register.go           # RegisterAll(cfg Config) wires all tools onto the MCP server
│   ├── cluster.go            # cluster-wide MCP tools
│   ├── nodes.go              # node MCP tools
│   ├── vms.go                # VM MCP tools
│   ├── containers.go         # container MCP tools
│   ├── snapshots.go          # snapshot MCP tools
│   ├── destructive.go        # delete_vm, delete_container (opt-in via Config.AllowDestructive)
│   ├── storage.go            # storage content MCP tools (Phase 3)
│   └── backup.go             # backup MCP tools (Phase 3)
├── .golangci.yml             # linter config
├── Makefile                  # quality gate targets
├── go.mod
├── go.sum
└── PLAN.md                   # this file
```

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `PROXMOX_API_URL` | yes | Base URL, e.g. `https://pve:8006/api2/json` |
| `PROXMOX_TOKEN_ID` | yes | Token ID, e.g. `root@pam!mcp` |
| `PROXMOX_TOKEN_SECRET` | yes | Token UUID secret |
| `PROXMOX_INSECURE` | no | Set `true` to skip TLS verification (self-signed certs) |
| `PROXMOX_ALLOW_DESTRUCTIVE` | no | Set `true` to register `delete_vm` and `delete_container` tools (default: disabled) |

## CLI Flags

| Flag | Default | Description |
|---|---|---|
| `--transport` | `stdio` | `stdio` or `http` |
| `--addr` | `localhost:8080` | Listen address when `--transport=http` |

## Makefile Targets

| Target | What it does |
|---|---|
| `make install-tools` | Installs `golangci-lint`, `gosec`, `govulncheck` |
| `make fmt` | `gofumpt -w .` |
| `make fix` | `go fix ./...` |
| `make vet` | `go vet ./...` |
| `make lint` | `golangci-lint run ./...` |
| `make sec` | `gosec ./...` (standalone) |
| `make vulncheck` | `govulncheck ./...` |
| `make test` | `go test -race -count=1 ./...` |
| `make build` | `go build ./cmd/proxmox-mcp/` |
| `make check` | **All of the above in order** — pre-commit gate |

## Linters Enabled (golangci-lint)

- `gosec` — hardcoded creds, TLS misconfig, weak crypto
- `govet` — struct alignment, printf mismatches
- `staticcheck` — broad static analysis
- `errcheck` — no silently dropped errors
- `ineffassign` — catch useless assignments
- `unused` — dead code
- `bodyclose` — HTTP response bodies must be closed
- `noctx` — HTTP requests must use context
- `gocritic` — opinionated style checks
- `revive` — drop-in `golint` replacement
- `misspell` — typos in comments/strings
- `prealloc` — slice preallocation hints
- `unconvert` — unnecessary type conversions
- `unparam` — unused function parameters
- `gofumpt` — formatting

`InsecureSkipVerify` usage annotated with `//nolint:gosec // G402: user explicitly opted in via PROXMOX_INSECURE=true`.

## Phase 1 MCP Tools (v0.1.0 — shipped)

| Tool | Description | Params |
|---|---|---|
| `list_nodes` | List all nodes in the cluster | — |
| `get_node_status` | Detailed status of a node | `node` |
| `list_cluster_resources` | All resources across cluster | `type` (optional: vm, storage, node, sdn) |
| `get_task_status` | Poll a Proxmox async task by UPID | `node`, `upid` |
| `list_vms` | QEMU VMs on a node | `node` |
| `get_vm_status` | VM status + current config | `node`, `vmid` |
| `start_vm` | Start a VM | `node`, `vmid` |
| `stop_vm` | Hard stop a VM | `node`, `vmid` |
| `shutdown_vm` | Graceful shutdown of a VM | `node`, `vmid` |
| `list_containers` | LXC containers on a node | `node` |
| `get_container_status` | Container status | `node`, `vmid` |
| `start_container` | Start a container | `node`, `vmid` |
| `stop_container` | Stop a container | `node`, `vmid` |

Lifecycle operations return task info immediately (non-blocking). Task status can be checked
separately.

---

## Phase 2 — Full CRUD, Snapshots, Parity, Depth ✅ shipped (v0.2.0)

6 PRs merged (#2–#7). Final tool count: **38 tools**.

| PR | Tools added | Key changes |
|---|---|---|
| #2 — client foundations | — | `postWithBody`, `delete`, `put` in `client.go` |
| #3 — lifecycle parity | `reboot_vm`, `suspend_vm`, `resume_vm`, `shutdown_container`, `reboot_container` | uses `post()` |
| #4 — snapshots | `list/create/rollback/delete` × VM + container (8 tools) | `snapshots.go`, `tools/snapshots.go` |
| #5 — delete operations | `delete_vm`, `delete_container` | `tools/destructive.go`; 3-layer safety gate; `PROXMOX_ALLOW_DESTRUCTIVE` env var |
| #6 — create & clone | `create_vm`, `clone_vm`, `create_container`, `clone_container` | `SensitiveString` type for password redaction |
| #7 — node/cluster depth | `get_cluster_status`, `list_node_storage`, `list_node_tasks`, `get_node_disks`, `get_vm_config`, `get_container_config` | read-only; no new HTTP primitives |

---

## Phase 3 — Config mutation, migration, storage content, backup (target: v0.3.0)

Builds on the `put` and `postWithBody` client methods already in place from Phase 2 PR #2.
Delivered across 4 PRs. Each must pass `make check` before merge.

### PR 7 — Config mutation (4 new tools)

Uses the existing `put()` client method — no new HTTP primitives needed.
Proxmox exposes a sync (`PUT`) and async (`POST`) variant for config; we use `PUT` here so
the result is immediate and no task polling is required.

New request structs `SetVMConfigRequest` and `SetContainerConfigRequest` in `types.go`.
Both expose a focused, safe subset of mutable fields — not a free-form map — to prevent
accidental misconfiguration. Fields use `omitempty` so callers only send what they intend to
change. Client methods take request structs by pointer (gocritic `hugeParam`).

Resize operations use a separate `PUT` endpoint that takes `disk` + `size` params; they
return a task UPID (async). New `ResizeDiskRequest` struct.

| Tool | API endpoint | Params |
|---|---|---|
| `set_vm_config` | `PUT /nodes/{node}/qemu/{vmid}/config` | `node`, `vmid`, `memory`, `cores`, `name`, `onboot`, `description` |
| `set_container_config` | `PUT /nodes/{node}/lxc/{vmid}/config` | `node`, `vmid`, `memory`, `swap`, `hostname`, `onboot`, `description` |
| `resize_vm_disk` | `PUT /nodes/{node}/qemu/{vmid}/resize` | `node`, `vmid`, `disk` (e.g. `scsi0`), `size` (e.g. `+10G` or `50G`) |
| `resize_container_disk` | `PUT /nodes/{node}/lxc/{vmid}/resize` | `node`, `vmid`, `disk` (e.g. `rootfs`), `size` |

`resize_vm_disk` and `resize_container_disk` return task UPIDs — use `get_task_status` to
poll for completion.

Tests: success + apiError for all four client methods (8 new tests). `set_vm_config` and
`set_container_config` additionally test that `omitempty` fields are omitted from the request
body.

### PR 8 — Migration (2 new tools)

Uses existing `postWithBody` — no new HTTP primitives needed.
Both tools return a task UPID immediately (non-blocking).

New `MigrateVMRequest` and `MigrateContainerRequest` structs in `types.go`. Client methods
in `vms.go` and `containers.go` respectively.

| Tool | API endpoint | Params |
|---|---|---|
| `migrate_vm` | `POST /nodes/{node}/qemu/{vmid}/migrate` | `node`, `vmid`, `target` (destination node), `online` (bool, live migrate) |
| `migrate_container` | `POST /nodes/{node}/lxc/{vmid}/migrate` | `node`, `vmid`, `target`, `restart` (bool) |

`online: true` performs a live migration for VMs (no guest downtime when supported).
`restart: true` for containers stops, migrates, and restarts the container on the target node.

Tests: success + apiError for both client methods (4 new tests).

### PR 9 — Storage content (3 new tools)

Uses existing `get()` for listing and the existing `delete()` method for removal.
New file `internal/proxmox/storage.go`. New file `tools/storage.go`.
`RegisterAll` gains `registerStorageTools`.

`delete_storage_content` is gated behind `PROXMOX_ALLOW_DESTRUCTIVE` and follows the same
3-layer safety pattern as `delete_vm`/`delete_container` (operator env var + `confirmed`
field + `DestructiveHint`).

| Tool | API endpoint | Params |
|---|---|---|
| `list_storage_content` | `GET /nodes/{node}/storage/{storage}/content` | `node`, `storage`, `content` (optional filter: `iso`, `vztmpl`, `backup`, `images`) |
| `get_storage_content_info` | `GET /nodes/{node}/storage/{storage}/content/{volume}` | `node`, `storage`, `volume` (volid, e.g. `local:iso/debian.iso`) |
| `delete_storage_content` | `DELETE /nodes/{node}/storage/{storage}/content/{volume}` | `node`, `storage`, `volume`, `confirmed` (must be `true`) |

`list_storage_content` is the primary discovery tool for ISOs and container templates —
needed to populate valid values for `create_vm` and `create_container`.

Tests: success + notFound for `list_storage_content` and `get_storage_content_info`;
success + apiError for `delete_storage_content` (5 new tests).

### PR 10 — Backup (2 new tools)

Uses existing `postWithBody` and `get()` — no new HTTP primitives needed.
New `CreateBackupRequest` struct in `types.go`. Client methods in a new
`internal/proxmox/backup.go`. New `tools/backup.go`. `RegisterAll` gains
`registerBackupTools`.

`create_backup` is async — returns a task UPID to poll with `get_task_status`.
`list_backups` is a convenience wrapper around `list_storage_content` filtered to
`content=backup`; it can be implemented as a dedicated client method or by reusing the
storage content client from PR 9.

| Tool | API endpoint | Params |
|---|---|---|
| `create_backup` | `POST /nodes/{node}/vzdump` | `node`, `vmid`, `storage`, `mode` (`snapshot`/`suspend`/`stop`, default `snapshot`), `compress` (`0`/`gzip`/`lzo`/`zstd`, default `zstd`) |
| `list_backups` | `GET /nodes/{node}/storage/{storage}/content?content=backup` | `node`, `storage` |

`mode: snapshot` is the preferred default — no guest downtime for VMs with QEMU guest agent.
`compress: zstd` gives the best speed/ratio trade-off on modern hardware.

Tests: success + apiError for `create_backup`; success + notFound for `list_backups`
(4 new tests).

## Proxmox API Notes

- Base URL: `https://<host>:8006/api2/json`
- All responses wrapped in `{"data": ...}` — unwrapped transparently by the client
- Many write operations (create, clone, migrate) are async — they return a UPID task ID;
  poll `/nodes/{node}/tasks/{upid}/status` for completion
- API token format: `Authorization: PVEAPIToken=USER@REALM!TOKENID=UUID`
- No CSRF token needed when using API tokens
- `GET /version` — good health-check endpoint

## Idiomatic Go Standards

- Error wrapping: `fmt.Errorf("listing VMs on node %s: %w", node, err)`
- Sentinel errors: `var ErrNotFound = errors.New("resource not found")`
- Pointer receivers on `Client` and mutable types
- Value receivers on small read-only structs
- `context.Context` first param on all I/O functions
- Exported types and functions always have doc comments
- `json` + `jsonschema` struct tags on all API types and MCP input structs
- Table-driven tests with `t.Run` subtests
- No `init()` anywhere

## Implementation Order

1. `go.mod`, `Makefile`, `.golangci.yml`
2. `internal/proxmox/types.go` — structs
3. `internal/proxmox/client.go` — HTTP client
4. `internal/proxmox/client_test.go` — client unit tests
5. `internal/proxmox/{nodes,vms,containers,tasks}.go` — API methods
6. `tools/{register,nodes,vms,containers,cluster}.go` — MCP tools
7. `cmd/proxmox-mcp/main.go` — entrypoint
8. `make check` — all gates pass

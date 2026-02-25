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
│       ├── nodes.go          # node-related API calls
│       ├── vms.go            # QEMU VM API calls
│       ├── containers.go     # LXC container API calls
│       └── tasks.go          # task polling (UPID → status)
├── tools/
│   ├── register.go           # RegisterAll(cfg Config) wires all tools onto the MCP server
│   ├── nodes.go              # node MCP tools
│   ├── vms.go                # VM MCP tools
│   ├── containers.go         # container MCP tools
│   ├── cluster.go            # cluster-wide MCP tools
│   ├── snapshots.go          # snapshot MCP tools (list/create/rollback/delete)
│   └── destructive.go        # delete_vm, delete_container (opt-in via Config.AllowDestructive)
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

## Phase 2 — Full CRUD, Snapshots, Parity, Depth (target: v0.2.0)

Delivered across 6 PRs in strict dependency order. Each PR must pass `make check` and CI
before merge. Final tool count: **38 tools** (13 existing + 25 new).

### PR 1 — Client foundations: `postWithBody`, `delete`, `put`

Extends `internal/proxmox/client.go` with three new private methods. No new tools — pure
client layer that all subsequent PRs depend on.

- `postWithBody(ctx, path string, body, result any) error` — marshals body to JSON, sets
  `Content-Type: application/json`. Required for create/clone/snapshot/migrate.
- `delete(ctx, path string, result any) error` — sends DELETE. Required for
  delete VM/container/snapshot.
- `put(ctx, path string, body, result any) error` — sends PUT with JSON body. Required for
  `set_vm_config` / `set_container_config` (Phase 3+).

All three covered by `httptest`-based unit tests in `client_test.go`.

### PR 2 — Lifecycle parity (5 new tools)

Uses existing `post()` — no client changes needed.

| Tool | API endpoint |
|---|---|
| `reboot_vm` | `POST /nodes/{node}/qemu/{vmid}/status/reboot` |
| `suspend_vm` | `POST /nodes/{node}/qemu/{vmid}/status/suspend` |
| `resume_vm` | `POST /nodes/{node}/qemu/{vmid}/status/resume` |
| `shutdown_container` | `POST /nodes/{node}/lxc/{vmid}/status/shutdown` |
| `reboot_container` | `POST /nodes/{node}/lxc/{vmid}/status/reboot` |

All return a task ID via `taskResult(upid)`.

### PR 3 — Snapshots (8 new tools)

Depends on PR 1 (`postWithBody` for create, `delete` for delete).
New file `internal/proxmox/snapshots.go`. New `Snapshot` struct in `types.go`.
New file `tools/snapshots.go`. `RegisterAll` gains `registerSnapshotTools`.

| Tool | Params |
|---|---|
| `list_vm_snapshots` | `node`, `vmid` |
| `create_vm_snapshot` | `node`, `vmid`, `name`, `description`, `include_ram` |
| `rollback_vm_snapshot` | `node`, `vmid`, `snapname` |
| `delete_vm_snapshot` | `node`, `vmid`, `snapname` |
| `list_container_snapshots` | `node`, `vmid` |
| `create_container_snapshot` | `node`, `vmid`, `name`, `description` |
| `rollback_container_snapshot` | `node`, `vmid`, `snapname` |
| `delete_container_snapshot` | `node`, `vmid`, `snapname` |

### PR 4 — Delete operations (2 new tools)

Depends on PR 1 (`delete()` client method).
`purge` defaults to `false` — disks are kept unless explicitly set to `true`.

**Safety design — 3 layers:**
1. **Operator gate** — `PROXMOX_ALLOW_DESTRUCTIVE=true` required to register the tools at all. `tools.RegisterAll` accepts a `tools.Config{AllowDestructive bool}` struct; tools remain invisible to the MCP client unless opted in.
2. **`DestructiveHint: true` annotation** — signals MCP clients to present confirmation UI before calling.
3. **`confirmed: true` required field** — tool handler returns an error if `confirmed` is not explicitly set to `true`; the model must reason about the action before proceeding.

New file `tools/destructive.go`. `tools.Config` struct added to `tools/register.go`. `RegisterAll` signature updated to accept `Config`.

| Tool | Params |
|---|---|
| `delete_vm` | `node`, `vmid`, `confirmed` (must be `true`), `purge` (bool, default false) |
| `delete_container` | `node`, `vmid`, `confirmed` (must be `true`), `purge` (bool, default false) |

### PR 5 — Create and clone (4 new tools)

Depends on PR 1 (`postWithBody`). New request structs `CreateVMRequest`, `CloneVMRequest`,
`CreateContainerRequest`, and `CloneContainerRequest` in `types.go` expose a focused subset
of Proxmox config options. Client methods take request structs by pointer (gocritic hugeParam).

New `SensitiveString` type in `types.go`: redacts value from `fmt` and `slog` output via
`fmt.Stringer` and `slog.LogValuer`, but marshals the real value to JSON for API calls.
`CreateContainerRequest.Password` uses `SensitiveString` — no `//nolint` directives needed.
`SensitiveString` also implements `json.Unmarshaler` so it works directly as an MCP input field.

| Tool | Params |
|---|---|
| `create_vm` | `node`, `vmid`, `name`, `memory`, `cores`, `iso`, `disk`, `net0`, `start` |
| `clone_vm` | `node`, `vmid`, `newid`, `name`, `target_node` |
| `create_container` | `node`, `vmid`, `ostemplate`, `hostname`, `memory`, `rootfs`, `password`, `net0`, `start` |
| `clone_container` | `node`, `vmid`, `newid`, `hostname`, `target_node` |

### PR 6 — Node and cluster depth (6 new tools) ✅ shipped

Read-only tools using existing `get()`. No new HTTP primitives needed.

| Tool | API endpoint |
|---|---|
| `get_cluster_status` | `GET /cluster/status` |
| `list_node_storage` | `GET /nodes/{node}/storage` |
| `list_node_tasks` | `GET /nodes/{node}/tasks` (params: `node`, `limit`) |
| `get_node_disks` | `GET /nodes/{node}/disks/list` |
| `get_vm_config` | `GET /nodes/{node}/qemu/{vmid}/config` |
| `get_container_config` | `GET /nodes/{node}/lxc/{vmid}/config` |

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

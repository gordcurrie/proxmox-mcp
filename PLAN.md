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
│   ├── register.go           # RegisterAll() wires all tools onto the MCP server
│   ├── nodes.go              # node MCP tools
│   ├── vms.go                # VM MCP tools
│   ├── containers.go         # container MCP tools
│   └── cluster.go            # cluster-wide MCP tools
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

## Initial MCP Tools (v1)

| Tool | Description | Params |
|---|---|---|
| `list_nodes` | List all nodes in the cluster | — |
| `get_node_status` | Detailed status of a node | `node` |
| `list_cluster_resources` | All resources across cluster | `type` (optional: vm, storage, node, sdn) |
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
separately. Architecture is designed so that full CRUD (create/clone/delete VMs, snapshots,
storage, migration, firewall, cluster config, etc.) can be added incrementally without
restructuring.

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

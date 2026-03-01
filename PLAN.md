# Proxmox MCP Server — Project Plan

## Summary

Go MCP server exposing Proxmox VE cluster operations as tools. Uses `modelcontextprotocol/go-sdk`,
a custom `net/http` Proxmox client (no third-party Proxmox library), API token auth, and strict
linting via `golangci-lint`/`gosec`/`govulncheck`.

## Plan Status

| Phase | Status |
|---|---|
| 1–6 (Foundation through CI & Releases) | ✅ Complete — 71 tools shipped (v0.6.0) |
| 7 — Storage Management | ✅ Complete — 76 tools (v0.7.0) |

## Decisions

| Decision | Choice | Rationale |
|---|---|---|
| MCP SDK | `github.com/modelcontextprotocol/go-sdk` | Official Go SDK |
| Proxmox client | Custom `net/http` wrapper | Full control, no external dep |
| Auth | API token only (`PVEAPIToken=`) | Stateless, no CSRF, no ticket renewal |
| Transports | stdio (default) + HTTP (`--transport=http`) | stdio for local clients, HTTP for remote |
| Formatter | `gofumpt` | Stricter than `gofmt` |
| TLS | Verify on by default; `PROXMOX_INSECURE=true` to skip | Proxmox uses self-signed certs by default |
| Destructive tools | Opt-in via `PROXMOX_ALLOW_DESTRUCTIVE=true` | Safe by default |

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `PROXMOX_API_URL` | yes | e.g. `https://pve:8006/api2/json` |
| `PROXMOX_TOKEN_ID` | yes | e.g. `root@pam!mcp` |
| `PROXMOX_TOKEN_SECRET` | yes | Token UUID secret |
| `PROXMOX_INSECURE` | no | `true` to skip TLS verification |
| `PROXMOX_ALLOW_DESTRUCTIVE` | no | `true` to register destructive tools |

## CLI Flags

| Flag | Default | Description |
|---|---|---|
| `--transport` | `stdio` | `stdio` or `http` |
| `--addr` | `localhost:8080` | Listen address when `--transport=http` |

## Makefile Targets

`make fix` → `make fmt` → `make vet` → `make lint` → `make sec` → `make vulncheck` →
`make test` → `make build` — all run by `make check`.

## Linters Enabled

`gosec`, `govet`, `staticcheck`, `errcheck`, `bodyclose`, `noctx`, `gofumpt`, `revive`,
`gocritic`, `unparam`, `unconvert`, `misspell`, `prealloc`, `ineffassign`, `unused`.

`InsecureSkipVerify` annotated with `//nolint:gosec // G402: user explicitly opted in via PROXMOX_INSECURE=true`.

---

## Phase 7 — Storage Management (target: v0.7.0)

**Goal**: Allow agents to register and manage storage targets cluster-wide — required for
the Proxmox Backup Server (PBS) integration with TrueNAS and for NFS-based backup storage.

Proxmox storage definitions are cluster-wide (not per-node). They live under `/storage` and
support many types: `nfs`, `pbs`, `dir`, `cifs`, `zfspool`, etc. `type=pbs` is the key
integration point for pointing Proxmox at a PBS instance running on TrueNAS SCALE.

### End-to-end PBS + TrueNAS workflow (enabled by this phase)

1. **TrueNAS** `create_dataset` — create the PBS datastore path (e.g. `tank/pbs-datastore`)
2. **TrueNAS** `install_custom_app` — deploy PBS as a Docker Compose app, mounting the dataset
3. **Proxmox** `add_storage` *(this phase)* — register `type=pbs` pointing at TrueNAS IP:8007
4. **Proxmox** `create_backup` *(exists)* — run backups to the PBS storage

**NFS fallback**: replace steps 2–3 with TrueNAS `create_nfs_share` (truenas-mcp Phase 10)
and Proxmox `add_storage` with `type=nfs`.

New file `internal/proxmox/storagedef.go` (separate from the existing `storage.go`, which
handles per-node storage *content*). New file `tools/storagedef.go`. `RegisterAll` gains
`registerStorageDefTools`.

### PR #20 — Storage definition read-only (2 new tools) ✅

Both get `ReadOnlyHint: true`. Tests: success + notFound for each (4 new tests).

| Tool | API endpoint | Params |
|---|---|---|
| `list_storages` | `GET /storage` | `type` (optional filter: `nfs`, `pbs`, `dir`, `cifs`, `zfspool`, etc.) |
| `get_storage` | `GET /storage/{storage}` | `storage` (name) |

### PR #21 — Storage definition write (3 new tools) ✅

`add_storage` uses `postWithBody`, `update_storage` uses `put`, `remove_storage` uses
`delete` — no new HTTP primitives needed.

`remove_storage` follows the 3-layer safety pattern (`PROXMOX_ALLOW_DESTRUCTIVE` +
`confirmed: true` + `DestructiveHint`).

New `AddStorageRequest` and `UpdateStorageRequest` structs in `types.go`. Fields use
`omitempty` so only set fields are sent.

| Tool | API endpoint | Params |
|---|---|---|
| `add_storage` | `POST /storage` | `storage` (name), `type` (`nfs`\|`pbs`\|`dir`\|`cifs`\|`zfspool`), `server` (optional), `export` (optional, NFS export path), `path` (optional, dir/local path), `datastore` (optional, PBS datastore name), `username` (optional, PBS/CIFS user), `password` (optional, PBS/CIFS), `content` (optional e.g. `backup,images`), `nodes` (optional, comma-sep to restrict to specific nodes) |
| `update_storage` | `PUT /storage/{storage}` | `storage`, plus any subset of the add params |
| `remove_storage` | `DELETE /storage/{storage}` | `storage`, `confirmed: true` — destructive opt-in |

Tests: success + apiError for add/update; success + notFound for remove (6 new tests).

**Phase 7 target tool count:** 71 + 5 = **76 tools** (70 always-on + 6 destructive opt-in).

Running total after PR #20: **73 tools** (68 always-on + 5 destructive).
Running total after PR #21: **76 tools** (70 always-on + 6 destructive).

---

## Proxmox API Notes

- Base URL: `https://<host>:8006/api2/json`
- All responses wrapped in `{"data": ...}` — unwrapped transparently by the client
- Many write operations are async — they return a UPID task ID;
  poll `/nodes/{node}/tasks/{upid}/status` for completion
- API token format: `Authorization: PVEAPIToken=USER@REALM!TOKENID=UUID`
- No CSRF token needed when using API tokens

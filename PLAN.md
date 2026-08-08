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
| 8 — Network Write Operations | ✅ Complete — 80 tools (PR #25) |
| Token optimisation | ✅ Complete — compact JSON responses, selective `omitempty` (PR #29) |
| Quality Parity with unifi-mcp | ✅ Complete — hardening, CI workflows, interface + tests (PR #32) |
| Access Control & Audit Tools | ✅ Complete — 83 tools: `list_users`, `list_user_tokens`, `get_node_journal` (PR #34) |

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
| Response format | Compact JSON (`json.Marshal`) | ~15% token reduction vs indented; no quality loss |

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

## Quality Parity with unifi-mcp ✅

These items bring proxmox-mcp to the same baseline as the unifi-mcp reference implementation.
Each is a distinct, independently reviewable change.

### Hardening & Robustness

| Item | Detail |
|---|---|
| Response body size cap | Add `maxResponseBytes = 10 << 20` (10 MiB) in `internal/proxmox/client.go`. Wrap `resp.Body` in `io.LimitReader(resp.Body, maxResponseBytes)` before `io.ReadAll`. Prevents memory exhaustion on unexpectedly large responses (e.g. full cluster task logs). |
| HTTP transport hardening | In `cmd/proxmox-mcp/main.go`, add `ReadTimeout: 30s`, `IdleTimeout: 120s`, `MaxHeaderBytes: 1 MiB`, and wrap handler with `http.MaxBytesHandler(..., 4<<20)`. Also change `err != http.ErrServerClosed` to `errors.Is(err, http.ErrServerClosed)`. |
| MCP-spec error responses | Add `errorResult(err error)` to `tools/helpers.go` returning the same 3-tuple as tool handlers: `(*mcp.CallToolResult, any, error)`, with `IsError: true` on the result and `nil, nil` for the other two values. Update all tool error returns from `return nil, nil, fmt.Errorf(...)` to `return errorResult(fmt.Errorf(...))`. Per MCP spec §6, tool execution errors should use `isError: true`, not protocol-level errors. |

### Tooling & CI

| Item | Detail |
|---|---|
| Pin tool versions in Makefile | Replace `@latest` in `install-tools` with explicit pins: `GOFUMPT_VERSION := v0.7.0`, `GOSEC_VERSION := v2.22.8`, `GOVULNCHECK_VERSION := v1.1.4`, `GOLANGCILINT_VERSION := v2.10.1`. Prevents surprise CI breakage when tools release. |
| `test-latest.yml` CI workflow | Daily cron (`22 10 * * *`) with `persist-credentials: false`, latest stable Go, `go get -u -t ./...`, then `go test -race -count=1 ./...`. Catches upstream dependency breaking changes before they reach main. See unifi-mcp `.github/workflows/test-latest.yml` as the template. |
| `govulncheck.yml` CI workflow | Daily cron + `workflow_dispatch` with `persist-credentials: false`, run `govulncheck ./...`. Provides a daily security signal independent of pushes. See unifi-mcp `.github/workflows/govulncheck.yml` as the template. |

### Testing

| Item | Detail |
|---|---|
| Client interface in tools layer | Add `tools/client_iface.go` with a `proxmoxClient` interface listing every method called by the tools layer. Update `RegisterAll` and all `register*Tools` functions to accept `proxmoxClient` instead of `*proxmox.Client`. Mirrors the `unifiClient` interface pattern in unifi-mcp. |
| Tool-layer unit tests | Add mock-based `tools/*_test.go` files once the interface is in place. Currently only `tools/helpers_test.go` exists. Cover happy path, error path, and input validation for each tool group. |

---

## Proxmox API Notes

- Base URL: `https://<host>:8006/api2/json`
- All responses wrapped in `{"data": ...}` — unwrapped transparently by the client
- Many write operations are async — they return a UPID task ID;
  poll `/nodes/{node}/tasks/{upid}/status` for completion
- API token format: `Authorization: PVEAPIToken=USER@REALM!TOKENID=UUID`
- No CSRF token needed when using API tokens

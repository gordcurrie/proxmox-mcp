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

## Proxmox API Notes

- Base URL: `https://<host>:8006/api2/json`
- All responses wrapped in `{"data": ...}` — unwrapped transparently by the client
- Many write operations are async — they return a UPID task ID;
  poll `/nodes/{node}/tasks/{upid}/status` for completion
- API token format: `Authorization: PVEAPIToken=USER@REALM!TOKENID=UUID`
- No CSRF token needed when using API tokens

# Copilot Instructions — proxmox-mcp

This is a Go MCP server that exposes Proxmox VE cluster operations as MCP tools.
See `PLAN.md` for the full project plan, decisions, and implementation order.

## Repository

- **GitHub**: `git@github.com:gordcurrie/proxmox-mcp.git`
- **Owner**: `gordcurrie` | **Repo**: `proxmox-mcp`
- Use owner `gordcurrie` and repo `proxmox-mcp` for all GitHub MCP tool calls (PRs, issues, etc.)

## Stack

- **Language**: Go (latest stable)
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk/mcp`
- **Proxmox client**: custom `net/http` wrapper in `internal/proxmox/` — no third-party Proxmox library
- **Auth**: Proxmox API tokens only (`PVEAPIToken=USER@REALM!TOKENID=UUID`)
- **Transports**: stdio (default) and HTTP (streamable), selected via `--transport` flag

## Code Style — Idiomatic Go

- Wrap errors: `fmt.Errorf("doing X: %w", err)` — never discard errors
- Sentinel errors: `var ErrNotFound = errors.New("resource not found")`
- `context.Context` is always the first parameter on any function that does I/O
- Pointer receivers on `Client` and mutable types; value receivers on small read-only structs
- No `init()` functions anywhere — explicit initialization only
- No global mutable state — inject dependencies
- All exported types and functions must have doc comments
- `json` tags on all API types; `jsonschema` tags on MCP tool input structs
- Table-driven tests using `t.Run` subtests
- Use `gofumpt` formatting (stricter than `gofmt`)

## Documentation — Required for Every PR

Every PR that adds or changes tools **must** update both files before committing:

1. **`README.md`** — add each new tool to the appropriate table (Cluster & Nodes, QEMU VMs,
   LXC Containers, Backups, Storage Content, or Destructive). Include the tool name, a
   one-line description, and its parameters. Destructive tools go in the Destructive table.

2. **`PLAN.md`** — mark the PR section as completed (add ✅ if not already there) and update
   the running tool count at the bottom of the phase section.

Do not skip these updates under any circumstances — documentation is part of the definition
of done for every PR, the same as passing `make check`.

## Git Commits

Always write multi-line commit messages via a temp file to avoid shell quoting issues:

```bash
python3 -c "open('/tmp/msg.txt','w').write('''subject line\n\nbody line 1\nbody line 2\n''')"
git add . && git commit -F /tmp/msg.txt
```

Never pass multi-line messages with `-m` — the shell mangles them.

## Quality Gates — `make check` must pass before every commit

```
make fix        # go fix ./...
make fmt        # gofumpt -w .
make vet        # go vet ./...
make lint       # golangci-lint run ./...
make sec        # gosec ./...
make vulncheck  # govulncheck ./...
make test       # go test -race -count=1 ./...
make build      # go build ./cmd/proxmox-mcp/
```

## Linting

Config is in `.golangci.yml`. Key linters: `gosec`, `govet`, `staticcheck`, `errcheck`,
`bodyclose`, `noctx`, `gofumpt`, `revive`, `gocritic`, `unparam`, `unconvert`.

When `InsecureSkipVerify` is needed (Proxmox self-signed TLS), annotate with:
```go
//nolint:gosec // G402: user explicitly opted in via PROXMOX_INSECURE=true
```

## Package Layout

- `internal/proxmox/` — Proxmox API client and types. No MCP imports here.
- `tools/` — MCP tool registration. Imports both `mcp` and `internal/proxmox`.
- `cmd/proxmox-mcp/` — Entrypoint only. Reads env/flags, wires things together, runs the server.

## Proxmox API Conventions

- All responses are wrapped in `{"data": ...}` — the client unwraps this transparently.
- Many write operations are async and return a UPID task ID.
  Poll `/nodes/{node}/tasks/{upid}/status` for completion via `tasks.go`.
- Lifecycle tools (`start_vm`, `stop_vm`, etc.) return task info immediately (non-blocking).

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `PROXMOX_API_URL` | yes | e.g. `https://pve:8006/api2/json` |
| `PROXMOX_TOKEN_ID` | yes | e.g. `root@pam!mcp` |
| `PROXMOX_TOKEN_SECRET` | yes | Token UUID |
| `PROXMOX_INSECURE` | no | `true` to skip TLS verification |
| `PROXMOX_ALLOW_DESTRUCTIVE` | no | `true` to register `delete_vm` and `delete_container` (default: disabled) |

## Security Rules — Non-Negotiable

### Never Commit Secrets
- No API tokens, passwords, UUIDs, or credentials anywhere in source files
- No hardcoded IPs, hostnames, or environment-specific values
- Secrets come from environment variables only — never from flags, config files, or defaults
- If a test needs credentials, use a clearly fake placeholder (e.g. `test-token-secret`) and
  document that it is a test fixture, not a real value
- `.env` files are for local development only — always in `.gitignore`, never committed
- If you spot a secret in code or history, flag it immediately before doing anything else

### Security-Conscious Development
- Prefer the most restrictive option by default (e.g. TLS verification is ON unless explicitly
  disabled via `PROXMOX_INSECURE=true`)
- Validate and sanitize all inputs from MCP tool arguments before passing to the Proxmox API
- Use `context.Context` with timeouts on all outbound HTTP calls — never hang indefinitely
- Keep dependencies minimal; every new dependency is a potential vulnerability surface
- Run `make vulncheck` after adding or updating any dependency
- Treat `gosec` warnings as bugs, not style suggestions

### Skipping Lint or Security Checks
- **Always ask the user before adding any `//nolint:` directive**
- When a `//nolint:` is genuinely required (e.g. the known `InsecureSkipVerify` opt-in),
  it must include the specific rule ID and a plain-English explanation of why it is safe:
  ```go
  //nolint:gosec // G402: InsecureSkipVerify is only set when PROXMOX_INSECURE=true,
  // which the user must explicitly opt into. Default is secure (verify enabled).
  ```
- Never suppress an entire linter for a file (`//nolint:gosec` at file scope) — always
  target the narrowest possible scope (single line)
- Never disable `errcheck` — if an error truly cannot be handled, document why with a comment

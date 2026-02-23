# proxmox-mcp

An [MCP](https://modelcontextprotocol.io) server that exposes [Proxmox VE](https://www.proxmox.com/en/proxmox-virtual-environment/) cluster operations as tools, built in Go using the official [go-sdk](https://github.com/modelcontextprotocol/go-sdk).

## Tools

| Tool | Description | Parameters |
|---|---|---|
| `list_nodes` | List all nodes in the cluster | — |
| `get_node_status` | Detailed status of a node | `node` |
| `list_cluster_resources` | All resources across the cluster | `type` (optional: `vm`, `storage`, `node`, `sdn`) |
| `list_vms` | QEMU VMs on a node | `node` |
| `get_vm_status` | VM status and current config | `node`, `vmid` |
| `start_vm` | Start a VM (returns task UPID) | `node`, `vmid` |
| `stop_vm` | Hard stop a VM (returns task UPID) | `node`, `vmid` |
| `shutdown_vm` | Graceful ACPI shutdown (returns task UPID) | `node`, `vmid` |
| `list_containers` | LXC containers on a node | `node` |
| `get_container_status` | Container status | `node`, `vmid` |
| `start_container` | Start a container (returns task UPID) | `node`, `vmid` |
| `stop_container` | Stop a container (returns task UPID) | `node`, `vmid` |
| `get_task_status` | Poll the status of an async task | `node`, `upid` |

Lifecycle operations are non-blocking — they return the UPID of the async task immediately. Use `get_task_status` to poll for completion.

## Prerequisites

- Go 1.26+
- A Proxmox VE cluster with API token access
- An API token created in **Datacenter → Permissions → API Tokens**

## Setup

```bash
git clone https://github.com/gordcurrie/proxmox-mcp
cd proxmox-mcp
cp .env.example .env   # then edit .env with your credentials
make build             # binary lands in bin/proxmox-mcp
```

## Configuration

All configuration is via environment variables:

| Variable | Required | Description |
|---|---|---|
| `PROXMOX_API_URL` | yes | e.g. `https://pve:8006/api2/json` |
| `PROXMOX_TOKEN_ID` | yes | e.g. `root@pam!mcp` |
| `PROXMOX_TOKEN_SECRET` | yes | Token UUID secret |
| `PROXMOX_INSECURE` | no | `true` to skip TLS verification (self-signed certs) |

Source your `.env` file before running:

```bash
set -a && source .env && set +a
```

## Running

### stdio (default — for local MCP clients)

```bash
./bin/proxmox-mcp
```

### HTTP (streamable — for remote/shared deployments)

```bash
./bin/proxmox-mcp --transport http --addr localhost:8080
```

## VS Code Copilot configuration

Create `.vscode/mcp.json` in your workspace (already gitignored):

```json
{
  "servers": {
    "proxmox-mcp": {
      "type": "stdio",
      "command": "/path/to/proxmox-mcp/bin/proxmox-mcp",
      "env": {
        "PROXMOX_API_URL": "https://pve:8006/api2/json",
        "PROXMOX_TOKEN_ID": "root@pam!mcp",
        "PROXMOX_TOKEN_SECRET": "your-token-secret"
      }
    }
  }
}
```

Then open the Copilot chat panel, switch to **Agent** mode, and the `proxmox-mcp` server will appear in the available tools.

## Development

```bash
make install-tools   # install golangci-lint, gosec, govulncheck, gofumpt
make check           # full quality gate: fix, fmt, vet, lint, sec, vulncheck, test, build
make test            # tests only
make build           # build only → bin/proxmox-mcp
make clean           # remove bin/
```

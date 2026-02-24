# proxmox-mcp

An [MCP](https://modelcontextprotocol.io) server that exposes [Proxmox VE](https://www.proxmox.com/en/proxmox-virtual-environment/) cluster operations as tools, built in Go using the official [go-sdk](https://github.com/modelcontextprotocol/go-sdk).

## Tools

### Cluster & Nodes

| Tool | Description | Parameters |
|---|---|---|
| `list_nodes` | List all nodes in the cluster | — |
| `get_node_status` | Detailed status of a node | `node` |
| `list_cluster_resources` | All resources across the cluster | `type` (optional: `vm`, `storage`, `node`, `sdn`) |

### QEMU VMs

| Tool | Description | Parameters |
|---|---|---|
| `list_vms` | QEMU VMs on a node | `node` |
| `get_vm_status` | VM status and current config | `node`, `vmid` |
| `start_vm` | Start a VM (returns task UPID) | `node`, `vmid` |
| `stop_vm` | Hard stop a VM (returns task UPID) | `node`, `vmid` |
| `shutdown_vm` | Graceful ACPI shutdown (returns task UPID) | `node`, `vmid` |
| `reboot_vm` | Reboot a VM (returns task UPID) | `node`, `vmid` |
| `suspend_vm` | Suspend a VM (returns task UPID) | `node`, `vmid` |
| `resume_vm` | Resume a suspended VM (returns task UPID) | `node`, `vmid` |
| `list_vm_snapshots` | List all snapshots for a VM | `node`, `vmid` |
| `create_vm_snapshot` | Create a VM snapshot (returns task UPID) | `node`, `vmid`, `snapname`, `description` (optional) |
| `rollback_vm_snapshot` | Roll back a VM to a snapshot (returns task UPID) | `node`, `vmid`, `snapname` |
| `delete_vm_snapshot` | Delete a VM snapshot (returns task UPID) | `node`, `vmid`, `snapname` |
| `create_vm` | Create a new QEMU VM (returns task UPID) | `node`, `vmid`, `name` (optional), `memory` (optional), `cores` (optional), `iso` (optional), `disk` (optional), `net0` (optional), `start` (optional) |
| `clone_vm` | Clone a VM to a new ID (returns task UPID) | `node`, `vmid`, `newid`, `name` (optional), `target_node` (optional) |

### LXC Containers

| Tool | Description | Parameters |
|---|---|---|
| `list_containers` | LXC containers on a node | `node` |
| `get_container_status` | Container status | `node`, `vmid` |
| `start_container` | Start a container (returns task UPID) | `node`, `vmid` |
| `stop_container` | Stop a container (returns task UPID) | `node`, `vmid` |
| `shutdown_container` | Graceful ACPI shutdown (returns task UPID) | `node`, `vmid` |
| `reboot_container` | Reboot a container (returns task UPID) | `node`, `vmid` |
| `list_container_snapshots` | List all snapshots for a container | `node`, `vmid` |
| `create_container_snapshot` | Create a container snapshot (returns task UPID) | `node`, `vmid`, `snapname`, `description` (optional) |
| `rollback_container_snapshot` | Roll back a container to a snapshot (returns task UPID) | `node`, `vmid`, `snapname` |
| `delete_container_snapshot` | Delete a container snapshot (returns task UPID) | `node`, `vmid`, `snapname` |
| `create_container` | Create a new LXC container (returns task UPID) | `node`, `vmid`, `ostemplate`, `hostname` (optional), `memory` (optional), `rootfs` (optional), `password` (optional), `net0` (optional), `start` (optional) |
| `clone_container` | Clone a container to a new ID (returns task UPID) | `node`, `vmid`, `newid`, `hostname` (optional), `target_node` (optional) |

### Tasks

| Tool | Description | Parameters |
|---|---|---|
| `get_task_status` | Poll the status of an async task | `node`, `upid` |

### Destructive (opt-in)

These tools are **not registered by default**. Set `PROXMOX_ALLOW_DESTRUCTIVE=true` to enable them.

| Tool | Description | Parameters |
|---|---|---|
| `delete_vm` | Permanently delete a stopped QEMU VM (returns task UPID) | `node`, `vmid`, `confirmed` (must be `true`), `purge` (optional) |
| `delete_container` | Permanently delete a stopped LXC container (returns task UPID) | `node`, `vmid`, `confirmed` (must be `true`), `purge` (optional) |

Lifecycle and snapshot operations are non-blocking — they return the UPID of the async task immediately. Use `get_task_status` to poll for completion.

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
| `PROXMOX_ALLOW_DESTRUCTIVE` | no | `true` to register `delete_vm` and `delete_container` tools (default: disabled) |

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

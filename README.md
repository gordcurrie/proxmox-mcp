# proxmox-mcp

An [MCP](https://modelcontextprotocol.io) server that exposes [Proxmox VE](https://www.proxmox.com/en/proxmox-virtual-environment/) cluster operations as tools, built in Go using the official [go-sdk](https://github.com/modelcontextprotocol/go-sdk).

## Tools

### Cluster & Nodes

| Tool | Description | Parameters |
|---|---|---|
| `list_nodes` | List all nodes in the cluster | — |
| `get_node_status` | Detailed status of a node | `node` |
| `list_cluster_resources` | All resources across the cluster | `type` (optional: `vm`, `storage`, `node`, `sdn`) |
| `get_cluster_status` | Cluster status and quorum information | — |
| `list_node_storage` | Storage pools available on a node | `node` |
| `list_node_tasks` | Recent tasks on a node | `node`, `limit` (optional) |
| `get_node_disks` | Physical disks detected on a node | `node` |

### QEMU VMs

| Tool | Description | Parameters |
|---|---|---|
| `list_vms` | QEMU VMs on a node | `node` |
| `get_vm_status` | VM status and current config | `node`, `vmid` |
| `get_vm_config` | Full VM configuration | `node`, `vmid` |
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
| `set_vm_config` | Update VM config (sync, no task) | `node`, `vmid`, `name` (optional), `memory` (optional), `cores` (optional), `onboot` (optional), `description` (optional) |
| `resize_vm_disk` | Resize a VM disk (returns task UPID) | `node`, `vmid`, `disk` (e.g. `scsi0`), `size` (e.g. `+10G` or `50G`) |
| `migrate_vm` | Migrate a VM to another node (returns task UPID) | `node`, `vmid`, `target`, `online` (optional, live migrate) |
| `restore_vm` | Restore a VM from a vzdump backup archive (returns task UPID) | `node`, `vmid`, `archive` (volid), `storage` (optional), `start` (optional) |
| `move_vm_disk` | Move a VM disk to a different storage pool (returns task UPID) | `node`, `vmid`, `disk` (e.g. `scsi0`), `storage` (target pool), `delete_source` (optional) |

### LXC Containers

| Tool | Description | Parameters |
|---|---|---|
| `list_containers` | LXC containers on a node | `node` |
| `get_container_status` | Container status | `node`, `vmid` |
| `get_container_config` | Full container configuration | `node`, `vmid` |
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
| `set_container_config` | Update container config (sync, no task) | `node`, `vmid`, `hostname` (optional), `memory` (optional), `swap` (optional), `onboot` (optional), `description` (optional) |
| `resize_container_disk` | Resize a container disk (returns task UPID) | `node`, `vmid`, `disk` (e.g. `rootfs`), `size` (e.g. `+5G` or `10G`) |
| `migrate_container` | Migrate a container to another node (returns task UPID) | `node`, `vmid`, `target`, `restart` (optional, stop+migrate+start) |
| `restore_container` | Restore a container from a vzdump backup archive (returns task UPID) | `node`, `vmid`, `archive` (volid), `storage` (optional), `hostname` (optional), `start` (optional) |

### Backups

| Tool | Description | Parameters |
|---|---|---|
| `create_backup` | Create a vzdump backup of a VM or container (returns task UPID) | `node`, `vmid`, `storage` (optional), `mode` (optional: `snapshot`\|`suspend`\|`stop`, default `snapshot`), `compress` (optional: `zstd`\|`gzip`\|`lzo`\|`0`, default `zstd`) |
| `list_backups` | List all backup volumes in a storage pool | `node`, `storage` |

### Tasks

| Tool | Description | Parameters |
|---|---|---|
| `get_task_status` | Poll the status of an async task | `node`, `upid` |

### Network

| Tool | Description | Parameters |
|---|---|---|
| `list_node_network` | List network interfaces on a node | `node`, `type` (optional: `bridge`, `bond`, `eth`, `alias`, `vlan`, `OVSBridge`, `OVSBond`, `OVSPort`, `OVSIntPort`, `any_bridge`) |
| `get_node_network_interface` | Get configuration of a specific network interface | `node`, `iface` (e.g. `vmbr0`) |
| `create_node_network_interface` | Create a new network interface on a node (staged until `apply_node_network_changes`) | `node`, `iface`, `type` (`bridge`\|`bond`\|`eth`\|`alias`\|`vlan`), `address` (optional), `netmask` (optional), `gateway` (optional), `address6` (optional), `gateway6` (optional), `mtu` (optional), `autostart` (optional), `bridge_ports` (optional), `bridge_stp` (optional), `bridge_fd` (optional), `bond_mode` (optional), `slaves` (optional), `comments` (optional) |
| `update_node_network_interface` | Update an existing network interface on a node (staged until `apply_node_network_changes`) | `node`, `iface`, `type` (required), plus any optional fields as in `create_node_network_interface` |
| `apply_node_network_changes` | Apply all staged network configuration changes on a node, reloading the network stack | `node` |

### Firewall

| Tool | Description | Parameters |
|---|---|---|
| `list_cluster_firewall_rules` | List all firewall rules at the datacenter level | — |
| `get_cluster_firewall_options` | Get datacenter firewall policy options (default in/out policies, logging) | — |
| `list_vm_firewall_rules` | List all firewall rules for a QEMU VM | `node`, `vmid` |
| `get_vm_firewall_options` | Get firewall policy options for a QEMU VM | `node`, `vmid` |
| `list_container_firewall_rules` | List all firewall rules for an LXC container | `node`, `vmid` |
| `get_container_firewall_options` | Get firewall policy options for an LXC container | `node`, `vmid` |
| `add_vm_firewall_rule` | Add a firewall rule to a QEMU VM | `node`, `vmid`, `type` (`in`\|`out`), `action` (`ACCEPT`\|`DROP`\|`REJECT`), `proto` (optional), `dport` (optional), `sport` (optional), `source` (optional), `dest` (optional), `iface` (optional), `comment` (optional), `enable` (optional) |
| `delete_vm_firewall_rule` | Delete a firewall rule from a QEMU VM by position | `node`, `vmid`, `pos` (zero-based) |
| `add_container_firewall_rule` | Add a firewall rule to an LXC container | `node`, `vmid`, `type`, `action`, `proto` (optional), `dport` (optional), `sport` (optional), `source` (optional), `dest` (optional), `iface` (optional), `comment` (optional), `enable` (optional) |
| `delete_container_firewall_rule` | Delete a firewall rule from an LXC container by position | `node`, `vmid`, `pos` (zero-based) |

### Pool Management

| Tool | Description | Parameters |
|---|---|---|
| `list_pools` | List all resource pools in the cluster | — |
| `get_pool` | Get full details of a pool including its member VMs, containers, and storage | `poolid` |
| `create_pool` | Create a new resource pool | `poolid`, `comment` (optional) |
| `update_pool` | Update a pool: change comment or add/remove member VMs and storage | `poolid`, `comment` (optional), `vms` (optional, comma-separated VM/CT IDs), `storage` (optional, comma-separated storage names), `delete` (optional, set `true` to remove listed members instead of adding) |

### Storage Content

| Tool | Description | Parameters |
|---|---|---|
| `list_storage_content` | List volumes in a storage pool | `node`, `storage`, `content` (optional: `iso`, `vztmpl`, `backup`, `images`) |
| `get_storage_content_info` | Detailed info about a specific volume | `node`, `storage`, `volume` (full volid, e.g. `local:iso/debian.iso`) |

### Storage Definitions

| Tool | Description | Parameters |
|---|---|---|
| `list_storages` | List all cluster-wide storage definitions | `type` (optional filter: `nfs`, `pbs`, `dir`, `cifs`, `zfspool`, etc.) |
| `get_storage` | Get full configuration of a storage definition | `storage` (name) |
| `add_storage` | Add a new storage target to the cluster | `storage` (name), `type` (required: `nfs`\|`pbs`\|`dir`\|`cifs`\|`zfspool`\|...), `server` (optional), `export` (optional, NFS path), `path` (optional, dir path), `datastore` (optional, PBS datastore), `username` (optional), `password` (optional), `fingerprint` (optional, PBS TLS fingerprint), `content` (optional, e.g. `backup,images`), `nodes` (optional, comma-sep), `shared` (optional bool) |
| `update_storage` | Update an existing storage definition | `storage` (name), plus any of: `server`, `export`, `path`, `datastore`, `username`, `password`, `fingerprint`, `content`, `nodes`, `shared` |

### Destructive (opt-in)

These tools are **not registered by default**. Set `PROXMOX_ALLOW_DESTRUCTIVE=true` to enable them.

| Tool | Description | Parameters |
|---|---|---|
| `delete_vm` | Permanently delete a stopped QEMU VM (returns task UPID) | `node`, `vmid`, `confirmed` (must be `true`), `purge` (optional) |
| `delete_container` | Permanently delete a stopped LXC container (returns task UPID) | `node`, `vmid`, `confirmed` (must be `true`), `purge` (optional) |
| `delete_storage_content` | Permanently delete a volume from a storage pool (returns task UPID) | `node`, `storage`, `volume` (full volid), `confirmed` (must be `true`) |
| `reboot_node` | Reboot an entire Proxmox node | `node`, `confirmed` (must be `true`) |
| `shutdown_node` | Shut down an entire Proxmox node | `node`, `confirmed` (must be `true`) |
| `delete_pool` | Permanently delete an empty resource pool | `poolid`, `confirmed` (must be `true`) |
| `remove_storage` | Remove a storage definition from the cluster (does not affect underlying data) | `storage` (name), `confirmed` (must be `true`) |
| `delete_node_network_interface` | Remove a network interface from a node (staged until `apply_node_network_changes`) | `node`, `iface`, `confirmed` (must be `true`) |

Lifecycle and snapshot operations are non-blocking — they return the UPID of the async task immediately. Use `get_task_status` to poll for completion.

## Installation

### Download a pre-built binary

Download the latest release for your platform from the [Releases](https://github.com/gordcurrie/proxmox-mcp/releases) page.

| Platform | Binary |
|---|---|
| Linux (amd64) | `proxmox-mcp_linux_amd64` |
| Linux (arm64) | `proxmox-mcp_linux_arm64` |
| macOS (amd64) | `proxmox-mcp_darwin_amd64` |
| macOS (arm64) | `proxmox-mcp_darwin_arm64` |
| Windows (amd64) | `proxmox-mcp_windows_amd64.exe` |

Make it executable and place it on your `PATH` (substitute the filename for your platform):

```bash
chmod +x <binary-name>
mv <binary-name> /usr/local/bin/proxmox-mcp
```

> Windows users: rename the `.exe` and add it to a directory on your `%PATH%`.

### Build from source

Requires Go 1.26+. You will also need a Proxmox VE API token — create one in **Datacenter → Permissions → API Tokens**.

```bash
git clone https://github.com/gordcurrie/proxmox-mcp
cd proxmox-mcp
cp .env.example .env   # copy the example env file
$EDITOR .env           # set PROXMOX_* values (see table below)
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
| `PROXMOX_ALLOW_DESTRUCTIVE` | no | `true` to register `delete_vm`, `delete_container`, `delete_storage_content`, `reboot_node`, `shutdown_node`, and `delete_pool` tools (default: disabled) |

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
        "PROXMOX_API_URL": "https://your-proxmox-host:8006/api2/json",
        "PROXMOX_TOKEN_ID": "user@realm!tokenid",
        "PROXMOX_TOKEN_SECRET": "your-token-secret"
      }
    }
  }
}
```

Then open the Copilot chat panel, switch to **Agent** mode, and the `proxmox-mcp` server will appear in the available tools.

## Claude Desktop configuration

Add the server to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "proxmox-mcp": {
      "command": "/path/to/proxmox-mcp/bin/proxmox-mcp",
      "env": {
        "PROXMOX_API_URL": "https://your-proxmox-host:8006/api2/json",
        "PROXMOX_TOKEN_ID": "user@realm!tokenid",
        "PROXMOX_TOKEN_SECRET": "your-token-secret"
      }
    }
  }
}
```

Restart Claude Desktop after saving the config — the Proxmox tools will appear in the tool selector.

## OpenCode configuration

Add the server to `opencode.json` in your project root (or `~/.config/opencode/opencode.json` for global config):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "proxmox-mcp": {
      "type": "local",
      "command": ["/path/to/proxmox-mcp/bin/proxmox-mcp"],
      "enabled": true,
      "environment": {
        "PROXMOX_API_URL": "https://your-proxmox-host:8006/api2/json",
        "PROXMOX_TOKEN_ID": "user@realm!tokenid",
        "PROXMOX_TOKEN_SECRET": "your-token-secret"
      }
    }
  }
}
```

## Development

```bash
make install-tools   # install golangci-lint, gosec, govulncheck, gofumpt
make check           # full quality gate: fix, fmt, vet, lint, sec, vulncheck, test, build
make test            # tests only (with race detector)
make build           # build only → bin/proxmox-mcp
make clean           # remove bin/proxmox-mcp
```

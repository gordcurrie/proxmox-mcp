---
name: proxmox
description: >
  Manage and inspect a Proxmox VE cluster using the proxmox-mcp tools. Use this
  skill when asked about Proxmox nodes, cluster health or quorum, QEMU VMs, LXC
  containers, snapshots, backups and restores, storage, resource pools, HA,
  node networking, or guest firewalls — for example "is my cluster healthy",
  "restart the grafana container", "snapshot VM 108 before I upgrade it",
  "when did container 100 last back up", or "move this VM to pve2".
license: MIT
---

# Skill: Proxmox VE (via proxmox-mcp)

This skill tells you how to use the tools exposed by
[gordcurrie/proxmox-mcp](https://github.com/gordcurrie/proxmox-mcp) safely and
efficiently. For a full health and security review, use the `audit-proxmox`
skill instead.

---

## Before you start

The tools are only available if the proxmox-mcp server is configured in your
MCP client. If no `list_nodes` / `list_cluster_resources` tools are available,
stop and point the user at the setup section of the
[README](https://github.com/gordcurrie/proxmox-mcp#installation). Do not try to
reach the Proxmox API directly with curl unless the user asks you to.

If a call fails with a TLS or certificate error, the host is most likely using
its default self-signed certificate: the user needs `PROXMOX_INSECURE=true`. If
it fails with `401`, the API token ID or secret is wrong or was revoked. A
`403` means the token lacks the permission for that call (check the token's
role and whether it uses privilege separation).

---

## Ground rules

1. **Read before you write.** Look up the current state with a `list_*` or
   `get_*` tool before changing anything, and show the user what will change.
2. **Find the node and VMID first.** Almost every guest tool takes both `node`
   and `vmid`. Resolve a guest name to its `vmid`, `node` and `type` (`qemu` or
   `lxc`) with `list_cluster_resources` (`type: "vm"` returns both VMs and
   containers). Never guess a VMID or assume which node a guest is on — guests
   move. Use the `qemu` tools (`*_vm`) for `type: "qemu"` and the container
   tools (`*_container`) for `type: "lxc"`.
3. **Every change needs the user's approval.** Get explicit approval for each
   mutation (start, stop, reboot, config change, resize, snapshot, rollback,
   migrate, restore, firewall rule, network change, storage or pool change)
   before calling the tool. A previous approval does not carry over to a
   different change. Only the destructive tools listed below declare a
   `confirmed` parameter: set it to `true` only after that approval. Do not
   send `confirmed` to any other tool.
4. **Protect before risky changes.** Before `set_vm_config` /
   `set_container_config`, `resize_vm_disk` / `resize_container_disk`,
   `rollback_*_snapshot`, `migrate_vm` / `migrate_container`, or
   `move_vm_disk` with `delete_source: true`, offer to take a snapshot (or a
   backup with `create_backup` when the storage cannot snapshot) and wait for
   it to finish before going ahead.
5. **Prefer graceful.** Use `shutdown_vm` / `shutdown_container` rather than
   `stop_vm` / `stop_container`. `stop_*` is an immediate power-off and can
   corrupt guest data; use it only when the user asks or a shutdown has
   already failed or hung.
6. **Poll async tasks.** Lifecycle, snapshot, backup, restore, clone, resize,
   migrate and disk-move tools return `{"task_id": "UPID:...", "status":
   "dispatched"}`. That only means the task started. Call `get_task_status`
   with the `node` named in the UPID (`UPID:<node>:...`) until `status` is
   `stopped`, then check `exitstatus`: `OK` means success, anything else is the
   error message. Do not report success before that.
7. **No pagination, but mind the task limit.** proxmox-mcp list tools return
   the full result in one call, except `list_node_tasks`: with `limit` `0` or
   omitted, Proxmox applies its own default (50 tasks). Pass an explicit
   `limit` (for example `500`) when looking further back.

---

## Common tasks

### Is my cluster healthy?

1. `get_cluster_status` — the `type: "cluster"` entry has `quorate` (`1` =
   quorate) and `nodes`; each `type: "node"` entry has `online` and `ip`. A
   standalone host (not in a cluster) returns no `type: "cluster"` entry: that
   is normal, not a quorum failure. Skip quorum and HA checks there.
2. `list_nodes` — per-node `status`, `cpu` (fraction 0–1), `mem` / `maxmem`,
   `disk` / `maxdisk` (bytes) and `uptime` (seconds).
3. `get_node_status` on a node for `loadavg`, `memory`, `swap`, `rootfs`,
   `pveversion` and `kversion`.
4. Flag: cluster not quorate, a node not `online`, memory or root disk above
   ~90 %, very short uptime (unexpected reboot), or nodes running different
   `pveversion`s.
5. For storage, ZFS, disk and HA checks, see the `audit-proxmox` skill.

### What is running where?

`list_cluster_resources` with `type: "vm"`. Each entry has `vmid`, `name`,
`type` (`qemu` / `lxc`), `node`, `status`, `cpu`, `mem` / `maxmem` and
`uptime`. Present a table sorted by VMID or node.

### Start, stop, reboot

- VMs: `start_vm`, `shutdown_vm`, `reboot_vm`, `stop_vm`, `suspend_vm`,
  `resume_vm`. Containers: `start_container`, `shutdown_container`,
  `reboot_container`, `stop_container`.
- Check `list_ha_resources` first. For an HA-managed guest (`sid` such as
  `vm:108` or `ct:100`), the HA manager owns its state: a stop or shutdown
  becomes an HA request, and the guest may be recovered on another node. Tell
  the user this before acting.
- Name the guest and say what depends on it before asking for approval.

### Change guest config

1. `get_vm_config` / `get_container_config` to show the current values.
2. Offer a snapshot (rule 4).
3. `set_vm_config` (`name`, `memory`, `cores`, `onboot`, `description`) or
   `set_container_config` (`hostname`, `memory`, `swap`, `onboot`,
   `description`). Only supplied fields change. Some changes (for example
   cores or memory on a running VM without hotplug) apply only after a reboot;
   say so.

### Resize a disk

`resize_vm_disk` / `resize_container_disk` take `disk` (for example `scsi0` or
`rootfs`) and `size` — absolute (`50G`) or an increment (`+10G`). Disks can
only grow; Proxmox rejects a shrink. Prefer the `+` form so the intent is
explicit. The guest filesystem usually has to be grown inside the guest
afterwards.

### Snapshots

- `list_vm_snapshots` / `list_container_snapshots` before creating, to avoid
  duplicate names. The list includes a `current` entry, which is the live
  state, not a snapshot.
- `create_vm_snapshot` (optionally `include_ram: true` for a running VM) /
  `create_container_snapshot`. Use a descriptive `snapname` such as
  `pre-upgrade-2026-10-20` (letters, digits, `-` and `_`; must start with a
  letter).
- Snapshots need snapshot-capable storage (ZFS, LVM-thin, Ceph, qcow2). If the
  task fails because the storage cannot snapshot, offer `create_backup`
  instead.
- `rollback_*_snapshot` discards **all** changes made since that snapshot.
  `delete_*_snapshot` cannot be undone. Both are available without the
  destructive flag, so confirm carefully: name the guest, the snapshot and its
  date.

### Backups and restores

- `list_storages` shows which storages hold `backup` content (a `pbs` type is
  Proxmox Backup Server). A storage with `shared: 1` shows the same backups
  from every node, so list it from one node; a non-shared storage (such as
  `local`) must be listed on each node.
- `list_backups` returns `volid`, `vmid`, `ctime` (Unix time), `size` and
  `format`. To answer "when did X last back up", take the highest `ctime` for
  that `vmid`.
- `create_backup` — `mode` defaults to `snapshot` (no downtime); `suspend` and
  `stop` pause or stop the guest. Poll the task.
- `restore_vm` / `restore_container` take the `archive` volid from
  `list_backups`. They will not overwrite an existing guest: the `vmid` must be
  free. To restore into an existing VMID the user must delete the guest first
  (destructive) — explain this and do not attempt it on your own.

### Clone, create, migrate

- `clone_vm` / `clone_container` need a free `newid`; check
  `list_cluster_resources` first.
- `create_vm` / `create_container` — find ISOs and templates with
  `list_storage_content` (`content: "iso"` or `"vztmpl"`).
- `migrate_vm` (`online: true` for live migration) / `migrate_container`
  (`restart: true` stops, moves and restarts a running container). Check the
  target node with `list_nodes` and that the guest's storage exists on it.
  Migration can take minutes; poll the task.
- `move_vm_disk` moves a disk between storages; with `delete_source: true` the
  original volume is removed after the move.

### Storage and pools

- `list_storages` / `get_storage` for cluster-wide definitions;
  `list_node_storage` for usage on one node; `list_storage_content` /
  `get_storage_content_info` for volumes.
- `add_storage` / `update_storage` change cluster-wide storage definitions.
  Never echo passwords back to the user.
- `list_pools` / `get_pool`; `create_pool`; `update_pool` adds members, or
  removes them with `delete: true`.

### Firewall

- Firewall rules only take effect when the datacenter firewall is enabled
  (`get_cluster_firewall_options` → `enable: 1`) **and** the guest firewall is
  enabled (`get_vm_firewall_options` / `get_container_firewall_options` →
  `enable: 1`) **and** the guest NIC has `firewall=1`. Missing keys in the
  options response mean the Proxmox default (disabled; `policy_in` `DROP`,
  `policy_out` `ACCEPT`).
- `add_vm_firewall_rule` / `add_container_firewall_rule` — always pass
  `enable: true`. If `enable` is omitted, Proxmox saves the rule
  **disabled** (`enable: 0`), despite what the tool description says. After
  adding, re-list the rules and confirm the new rule shows `enable: 1`. Do not report a rule as active until you have checked. `type`
  is `in` or `out`; `action` is `ACCEPT`, `DROP` or `REJECT`. The rule is
  saved immediately (no task), but it is only enforced when the firewall is
  enabled at every level (see above).
- `delete_vm_firewall_rule` / `delete_container_firewall_rule` take a
  zero-based `pos`. Positions shift after every delete: re-list the rules
  before each delete and confirm the rule's content, not just its number.

### Node networking

- `list_node_network` / `get_node_network_interface` to read.
- `create_node_network_interface` / `update_node_network_interface` only stage
  changes. Nothing changes until `apply_node_network_changes`, which reloads
  the node's network stack.
- A wrong bridge, address or gateway change can cut the node off — including
  the connection this MCP server uses. Show the full staged change and get a
  separate approval before calling `apply_node_network_changes`.

### Tasks and logs

- `list_node_tasks` (with a `limit`) shows recent tasks; a finished task's
  `status` is `OK` or the error text.
- `get_node_journal` returns systemd journal lines; use `last_entries` or
  `since` / `until` (Unix timestamps) to keep the output small.

---

## Destructive tools

These tools are only registered when the server runs with
`PROXMOX_ALLOW_DESTRUCTIVE=true`, and each requires `confirmed: true`:
`delete_vm`, `delete_container`, `delete_storage_content`, `remove_storage`,
`delete_pool`, `delete_node_network_interface`, `reboot_node` and
`shutdown_node`. If the user asks for one of these and the tool is missing,
explain that it is disabled by design and how to enable it. Do not look for a
workaround (for example, do not use `stop_*` plus other tools to approximate a
delete, or reboot every guest to approximate a node reboot).

`reboot_node` and `shutdown_node` take down every guest on that node that is
not migrated away first, and may drop the connection this server uses.

`delete_node_network_interface` only stages the removal, like the other
network writes. Nothing changes until `apply_node_network_changes`, which
applies **every** staged change on that node at once. Show the full staged set
and get a separate approval before applying (see Node networking).

---

## Limits

proxmox-mcp does not expose: the guest console or exec inside a guest, backup
job schedules (`/cluster/backup`), replication, Ceph, SDN configuration
writes, permissions/ACL writes, or user and token management. If asked for
these, say so and suggest the Proxmox web UI.

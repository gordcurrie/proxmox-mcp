---
name: audit-proxmox
description: >
  Health and security audit of a Proxmox VE cluster using the proxmox-mcp
  tools. Use this skill when asked to audit, review, check, or assess a Proxmox
  cluster — including cluster quorum, node resources and versions, disk SMART
  health, ZFS pool health, storage capacity, HA state, backup coverage,
  snapshot sprawl, firewall configuration, users and API tokens, and SSH
  authentication activity. Produces a prioritised findings report.
license: MIT
---

# Skill: Audit Proxmox (via proxmox-mcp)

This skill performs a structured health and security audit of a Proxmox VE
cluster using the tools exposed by
[gordcurrie/proxmox-mcp](https://github.com/gordcurrie/proxmox-mcp). Work
through each section below in order, collect findings, then produce a
**Prioritised Findings Report** at the end. For everyday operations, use the
`proxmox` skill instead.

---

## Before you start

Confirm the proxmox-mcp server is configured and these tools are available:
`get_cluster_status`, `list_nodes`, `get_node_status`, `list_cluster_resources`,
`get_node_disks`, `get_disk_smart`, `list_zfs_pools`, `get_zfs_pool`,
`list_storages`, `list_node_storage`, `get_ha_status`, `list_ha_resources`,
`list_backups`, `list_vm_snapshots`, `list_container_snapshots`,
`get_cluster_firewall_options`, `list_cluster_firewall_rules`,
`get_vm_firewall_options`, `get_container_firewall_options`, `list_users`,
`list_user_tokens`, `get_node_journal`, `list_node_tasks`.

This audit is **read-only**. Do not call any tool that changes state during
the audit, even to fix a finding. List remediations in the report and let the
user decide what to act on afterwards (the `proxmox` skill covers how).

proxmox-mcp list tools return the full result in one call; there is no
pagination. Timestamps (`ctime`, `snaptime`, `starttime`, `expire`) are Unix
seconds; convert them to dates in the report.

---

## Section 1 — Cluster and quorum

1. `get_cluster_status`. The `type: "cluster"` entry has `name`, `nodes` and
   `quorate`; each `type: "node"` entry has `name`, `ip` and `online`.
2. `list_nodes` for `status`, `cpu`, `mem` / `maxmem`, `disk` / `maxdisk` and
   `uptime` per node.

**Findings to raise:**
- `[CRITICAL]` Cluster not quorate (`quorate: 0`)
- `[HIGH]` Any node with `online: 0` or `status` other than `online`
- `[MEDIUM]` A cluster with exactly 2 nodes and no QDevice — losing either node
  loses quorum (ask the user whether a QDevice exists; it is not visible here)
- `[INFO]` Node count and names

---

## Section 2 — Node resources and versions

Call `get_node_status` for **every** node.

**Check for:**
- `memory.used / memory.total` above 90 % — `[MEDIUM]`
- `swap.used / swap.total` above 50 % — `[LOW]`; sustained memory pressure
- `rootfs.used / rootfs.total` above 85 % — `[MEDIUM]`; above 95 % — `[HIGH]`
  (a full root disk breaks the cluster filesystem and logging)
- `loadavg[0]` well above `cpuinfo.cpus` — `[LOW]`
- `pveversion` or `kversion` differing between nodes — `[MEDIUM]`; mixed
  versions are only expected mid-upgrade
- `uptime` under one day with no user-planned reboot — `[LOW]`; possible
  unexpected reboot
- `uptime` above ~90 days — `[LOW]`; likely missing kernel security updates
- `boot-info.secureboot: 0` — `[INFO]`

---

## Section 3 — Physical disks (SMART)

1. `get_node_disks` on every node. Note `devpath`, `model`, `type`, `used`,
   `health` and `wearout`.
2. Call `get_disk_smart` for **every** disk (`disk` = `devpath`). Do not skip
   disks that look healthy in step 1.

**Check for:**
- `health` other than `PASSED` / `OK` — `[CRITICAL]`
- `wearout` (percent of life **remaining**; `100` = new, `N/A` = unknown)
  below 10 — `[HIGH]`; below 30 — `[MEDIUM]`
- Non-zero raw values for `Reallocated_Sector_Ct`, `Reported_Uncorrect`,
  `Current_Pending_Sector`, `Offline_Uncorrectable`, `Grown_Bad_Blocks`,
  `Program_Fail_Count` or `Erase_Fail_Count` — `[HIGH]`
- Non-zero `UDMA_CRC_Error_Count` — `[LOW]`; usually a cable or backplane issue
- NVMe disks report a `text` block instead of `attributes`: look for
  `Critical Warning` not `0x00`, `Percentage Used` above 80 %, or non-zero
  `Media and Data Integrity Errors` — `[HIGH]`

---

## Section 4 — ZFS pools

1. `list_zfs_pools` on every node. Note `name`, `health`, `size`, `alloc`,
   `free` and `frag`.
2. `get_zfs_pool` for every pool. It returns the pool `state`, `errors`,
   `scan` (last scrub or resilver, as text), optional `status` / `action`
   advice, and a nested `children` tree of vdevs and disks, each with `name`,
   `state`, `read`, `write` and `cksum`. Walk the whole tree.

**Check for:**
- Pool `health` / `state` `DEGRADED`, `FAULTED` or `UNAVAIL` — `[CRITICAL]`
- Any vdev or disk not `ONLINE`, or non-zero `read` / `write` / `cksum` —
  `[HIGH]`
- `errors` other than `No known data errors` — `[HIGH]`
- `alloc / size` above 80 % — `[MEDIUM]`; above 90 % — `[HIGH]` (ZFS
  performance drops sharply when nearly full)
- `frag` above 50 % — `[LOW]`
- A pool whose top-level vdev is a single disk (no `mirror-*` or `raidz*`
  child) — `[MEDIUM]`; one disk failure loses the pool
- `scan` showing no scrub, or a scrub older than ~35 days — `[LOW]`
- A `status` saying features are not enabled — `[INFO]`; do not recommend
  `zpool upgrade` without warning that it can break older software

---

## Section 5 — Storage capacity

1. `list_storages` for definitions (`storage`, `type`, `content`, `shared`,
   `nodes`).
2. `list_node_storage` on every node for `enabled`, `active`, `used`,
   `total` and `used_fraction` (0–1).

**Check for:**
- A storage with `enabled: 1` but `active: 0` on a node — `[HIGH]`
- `used_fraction` above 0.85 — `[MEDIUM]`; above 0.95 — `[HIGH]`
- No storage with `backup` in `content` — `[HIGH]`
- The only backup storage lives on the same node or disks as the guests
  (for example `local`) — `[MEDIUM]`; one hardware failure loses guests and
  backups together
- Never repeat storage passwords or fingerprints in the report.

---

## Section 6 — High availability

1. `get_ha_status`. Note the `quorum` entry (`quorate`), the `master` entry,
   the `fencing` entry (`armed-state`), every `lrm:<node>` entry and every
   `service:*` entry (`sid`, `node`, `state`, `crm_state`, `request_state`).
2. `list_ha_resources` and `list_ha_groups`.

**Check for:**
- Any service in `error`, `fence` or `recovery` state — `[HIGH]`
- A service whose `state` differs from its `request_state` for more than a few
  minutes — `[MEDIUM]`
- HA resources defined but no `master` entry, or `fencing` not armed —
  `[HIGH]`
- HA-managed guests that use only node-local storage (they cannot be
  recovered on another node) — `[MEDIUM]`; check with `get_vm_config` /
  `get_container_config`
- `[INFO]` Which guests are HA-managed and which are not

---

## Section 7 — Backup coverage

1. `list_cluster_resources` with `type: "vm"` to get every guest (`vmid`,
   `name`, `type`, `node`, `status`). This list does not mark templates:
   before flagging a stopped guest, check `get_vm_config` /
   `get_container_config` for `template: 1` and skip templates.
2. For every storage with `backup` in `content`, call `list_backups`. A
   `shared: 1` storage needs only one node; a non-shared storage must be
   listed on **each** node it is enabled on.
3. For each guest, find its newest backup (highest `ctime` for its `vmid`)
   across all backup storages.

**Check for:**
- A running guest with **no backup at all** — `[HIGH]`
- A guest whose newest backup is older than 7 days — `[MEDIUM]`; older than
  30 days — `[HIGH]`
- A guest with only one backup (no retention history) — `[LOW]`
- Recent `vzdump` tasks that failed: `list_node_tasks` on each node (for
  example `limit: 200`), entries with `type: "vzdump"` and `status` other than
  `OK` — `[HIGH]`

Backup job schedules are not exposed by proxmox-mcp; judge coverage from the
backups that actually exist.

---

## Section 8 — Snapshot hygiene

For every guest, call `list_vm_snapshots` or `list_container_snapshots`.
Ignore the `current` entry — it is the live state, not a snapshot.

**Check for:**
- Snapshots older than 30 days (`snaptime`) — `[LOW]`; they hold old blocks,
  use space and slow down some operations. Report the count and oldest date
  per guest, not every snapshot.
- More than ~20 snapshots on one guest — `[LOW]`
- `[INFO]` Snapshots are not backups: a guest with snapshots but no backup is
  still covered by the Section 7 finding.

---

## Section 9 — Firewall

1. `get_cluster_firewall_options` and `list_cluster_firewall_rules`.
2. For every guest: `get_vm_firewall_options` / `get_container_firewall_options`,
   then `list_vm_firewall_rules` / `list_container_firewall_rules` where the
   guest firewall is enabled.

Keys missing from an options response mean the Proxmox default: `enable` is
`0`, `policy_in` is `DROP`, `policy_out` is `ACCEPT`. Guest rules only apply
when the datacenter firewall is enabled **and** the guest firewall is enabled
**and** the guest NIC has `firewall=1` (check `net*` in the guest config).

**Check for:**
- Datacenter firewall disabled (`enable` missing or `0`) — `[MEDIUM]`; no
  Proxmox firewall rule at any level is enforced. Lower to `[INFO]` if the
  user confirms filtering is done upstream.
- Guest firewall enabled with rules, but the datacenter firewall disabled —
  `[MEDIUM]`; the rules look protective but do nothing
- Guest `policy_in: ACCEPT` with the firewall enabled — `[MEDIUM]`
- An `in` rule with `action: ACCEPT`, no `source` and a management port
  (`22`, `8006`, `3128`, `5900:5999`) — `[HIGH]`
- An `in` rule with `sport` set to the service port (for example `sport: 22`
  next to `dport: 22`) — `[MEDIUM]`; client source ports are ephemeral, so
  the rule almost never matches and the traffic falls through to the policy
- Disabled rules (`enable: 0`) — `[LOW]`; confirm they are intentional

---

## Section 10 — Users and API tokens

1. `list_users`. Note `userid`, `realm-type`, `enable`, `expire` and `keys`.
2. `list_user_tokens` for **every** user. Note `tokenid`, `comment`,
   `expire` and `privsep`.

**Check for:**
- Users the owner does not recognise — `[HIGH]`
- Enabled users with an `expire` date in the past, or disabled users that
  still have tokens — `[LOW]`
- `root@pam` or other admin users without two-factor authentication — `[MEDIUM]`.
  A non-empty `keys` field (often redacted as `x`) suggests TFA is configured;
  confirm in the web UI before reporting.
- Tokens with `privsep: 0` — the token has the user's **full** privileges.
  `[HIGH]` for tokens on `root@pam`; `[MEDIUM]` for other users. Recommend
  privilege-separated tokens with a minimal role.
- Tokens with `expire: 0` (never expire) — `[LOW]`
- Tokens with no `comment` — `[INFO]`; hard to tell what they are for
- Never claim to see token secrets — they are never returned.

---

## Section 11 — Authentication activity

Call `get_node_journal` on every node with `since` set to 7 days ago (Unix
time). If the output is very large, use `last_entries` (for example `2000`)
instead. The result is an array of syslog-style lines; the first and last
entries are journal cursors (`s=...`), not log lines.

**Check for:**
- `Failed password`, `Invalid user` or `authentication failure` from `sshd`,
  especially many attempts from one source IP — `[HIGH]` if the source is
  outside the local network, `[MEDIUM]` otherwise
- `Accepted password` for `root` over SSH — `[MEDIUM]`; prefer key-only SSH
- `pvedaemon` / `pveproxy` `authentication failure` lines for the web UI or
  API — `[MEDIUM]`
- Quote the shortest decisive log line, with the timestamp and node.

---

## Section 12 — Limits of proxmox-mcp

Some audit-relevant data is **not available** via these tools and needs the
Proxmox web UI or a shell:

| Data | Where to find it |
|---|---|
| Backup job schedules and retention | Datacenter → Backup |
| Permissions / ACLs and roles | Datacenter → Permissions |
| TFA configuration per user | Datacenter → Permissions → Two Factor |
| Pending package and security updates | Node → Updates |
| Subscription and repository configuration | Node → Repositories |
| Replication jobs | Datacenter → Replication |
| Ceph health | Node → Ceph |
| PBS datastore verify and prune jobs | Proxmox Backup Server UI |
| QDevice presence | `pvecm status` on a node |

---

## Output — Prioritised Findings Report

After completing all sections, produce a report in this format:

```
# Proxmox Audit — <cluster name> — <date>

## Summary
- Nodes: N (M online), quorate: yes/no
- Guests: N VMs, M containers (K HA-managed)
- Total findings: X critical, Y high, Z medium, W low, V info

## Findings

### [CRITICAL] <title>
**Section:** <N>
**Detail:** <what was found: node, VMID and name, disk, pool, user or token>
**Recommendation:** <what to do>

### [HIGH] <title>
...

### [MEDIUM] <title>
...

### [LOW] <title>
...

### [INFO] <title>
...

## Items requiring manual review
- [ ] Backup job schedules and retention
- [ ] Permissions and roles
- [ ] Pending updates on each node
- [ ] PBS verify and prune jobs
```

Severity definitions:
- **CRITICAL** — data loss or outage is happening or imminent; act immediately
- **HIGH** — significant exposure or risk of data loss; fix before the next audit
- **MEDIUM** — meaningful risk; fix within a reasonable window
- **LOW** — hygiene issue; fix when convenient
- **INFO** — noted for the record; no action required

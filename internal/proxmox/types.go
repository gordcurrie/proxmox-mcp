// Package proxmox provides a client for the Proxmox VE REST API.
package proxmox

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
)

// SensitiveString is a string type that redacts its value from fmt and slog
// output to prevent accidental logging of secrets. It marshals its real value
// to JSON so it can be used directly in API request bodies.
type SensitiveString string

// String implements fmt.Stringer — returns "[REDACTED]" so the value is never
// emitted by fmt.Print, fmt.Sprintf, log, or any logger that calls String().
func (s SensitiveString) String() string { return "[REDACTED]" }

// GoString implements fmt.GoStringer — returns "[REDACTED]" so the value is
// never leaked via %#v formatting.
func (s SensitiveString) GoString() string { return "[REDACTED]" }

// MarshalJSON implements json.Marshaler — emits the real underlying value so
// that API request bodies are serialised correctly.
func (s SensitiveString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// LogValue implements slog.LogValuer — prevents the value from appearing in
// structured log output produced by log/slog.
func (s SensitiveString) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

// UnmarshalJSON implements json.Unmarshaler — decodes a plain JSON string into
// a SensitiveString so it can be used as an MCP tool input field.
func (s *SensitiveString) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("unmarshalling SensitiveString: %w", err)
	}
	*s = SensitiveString(v)
	return nil
}

// ErrNotFound is returned when a requested resource does not exist (HTTP 404).
var ErrNotFound = errors.New("resource not found")

// apiResponse is the standard envelope wrapping every Proxmox API response.
type apiResponse struct {
	Data json.RawMessage `json:"data"`
}

// Node represents a single entry from GET /nodes.
type Node struct {
	Node    string  `json:"node"`
	Status  string  `json:"status"`
	Type    string  `json:"type"`
	CPU     float64 `json:"cpu"`
	MaxCPU  int     `json:"maxcpu"`
	Mem     int64   `json:"mem"`
	MaxMem  int64   `json:"maxmem"`
	Disk    int64   `json:"disk"`
	MaxDisk int64   `json:"maxdisk"`
	Uptime  int64   `json:"uptime"`
}

// VM represents a single entry from GET /nodes/{node}/qemu.
type VM struct {
	VMID      int     `json:"vmid"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	CPU       float64 `json:"cpu"`
	Mem       int64   `json:"mem"`
	MaxMem    int64   `json:"maxmem"`
	Disk      int64   `json:"disk"`
	MaxDisk   int64   `json:"maxdisk"`
	Uptime    int64   `json:"uptime"`
	NetOut    int64   `json:"netout"`
	NetIn     int64   `json:"netin"`
	DiskRead  int64   `json:"diskread"`
	DiskWrite int64   `json:"diskwrite"`
}

// Container represents a single entry from GET /nodes/{node}/lxc.
type Container struct {
	VMID    int     `json:"vmid"`
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	CPU     float64 `json:"cpu"`
	Mem     int64   `json:"mem"`
	MaxMem  int64   `json:"maxmem"`
	Disk    int64   `json:"disk"`
	MaxDisk int64   `json:"maxdisk"`
	Uptime  int64   `json:"uptime"`
	NetOut  int64   `json:"netout"`
	NetIn   int64   `json:"netin"`
}

// ClusterResource represents a single entry from GET /cluster/resources.
type ClusterResource struct {
	ID      string  `json:"id"`
	Type    string  `json:"type"`
	Node    string  `json:"node,omitempty"`
	Status  string  `json:"status,omitempty"`
	Name    string  `json:"name,omitempty"`
	VMID    int     `json:"vmid,omitempty"`
	CPU     float64 `json:"cpu,omitempty"`
	MaxCPU  int     `json:"maxcpu,omitempty"`
	Mem     int64   `json:"mem,omitempty"`
	MaxMem  int64   `json:"maxmem,omitempty"`
	Disk    int64   `json:"disk,omitempty"`
	MaxDisk int64   `json:"maxdisk,omitempty"`
	Uptime  int64   `json:"uptime,omitempty"`
}

// TaskStatus represents the result from GET /nodes/{node}/tasks/{upid}/status.
type TaskStatus struct {
	UPID       string `json:"upid"`
	Node       string `json:"node"`
	PID        int    `json:"pid"`
	Status     string `json:"status"`     // "running" | "stopped"
	ExitStatus string `json:"exitstatus"` // "OK" or error string
	Type       string `json:"type"`
	ID         string `json:"id"`
	User       string `json:"user"`
	StartTime  int64  `json:"starttime"`
}

// Snapshot represents a single entry from GET /nodes/{node}/qemu/{vmid}/snapshot
// or GET /nodes/{node}/lxc/{vmid}/snapshot.
type Snapshot struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parent      string `json:"parent,omitempty"`
	SnapTime    int64  `json:"snaptime,omitempty"`
	VMState     int    `json:"vmstate,omitempty"` // 1 if RAM was included (VM only)
}

// CreateVMSnapshotRequest is the request body for POST /nodes/{node}/qemu/{vmid}/snapshot.
type CreateVMSnapshotRequest struct {
	Snapname    string `json:"snapname"`
	Description string `json:"description,omitempty"`
	VMState     int    `json:"vmstate,omitempty"` // 1 to include RAM state
}

// CreateContainerSnapshotRequest is the request body for POST /nodes/{node}/lxc/{vmid}/snapshot.
type CreateContainerSnapshotRequest struct {
	Snapname    string `json:"snapname"`
	Description string `json:"description,omitempty"`
}

// CreateVMRequest is the request body for POST /nodes/{node}/qemu.
// Required fields (vmid) are always included; optional fields use omitempty
// and are omitted when zero-valued.
type CreateVMRequest struct {
	VMID   int    `json:"vmid"`
	Name   string `json:"name,omitempty"`
	Memory int    `json:"memory,omitempty"` // MB
	Cores  int    `json:"cores,omitempty"`
	IDE2   string `json:"ide2,omitempty"`   // ISO drive, e.g. "local:iso/file.iso,media=cdrom"
	SCSI0  string `json:"scsi0,omitempty"`  // Primary disk, e.g. "local-lvm:32"
	SCSIHW string `json:"scsihw,omitempty"` // Controller, e.g. "virtio-scsi-pci"
	Net0   string `json:"net0,omitempty"`   // Network, e.g. "virtio,bridge=vmbr0"
	OSType string `json:"ostype,omitempty"` // e.g. "l26" for Linux 2.6+
	Start  int    `json:"start,omitempty"`  // 1 to start after creation
}

// CloneVMRequest is the request body for POST /nodes/{node}/qemu/{vmid}/clone.
type CloneVMRequest struct {
	NewID  int    `json:"newid"`
	Name   string `json:"name,omitempty"`
	Target string `json:"target,omitempty"` // target node; defaults to source node
}

// CreateContainerRequest is the request body for POST /nodes/{node}/lxc.
// Required fields (vmid, ostemplate) are always included; optional fields use
// omitempty and are omitted when zero-valued.
type CreateContainerRequest struct {
	VMID       int             `json:"vmid"`
	OSTemplate string          `json:"ostemplate"` // e.g. "local:vztmpl/debian-12-standard_12.7-1_amd64.tar.zst"
	Hostname   string          `json:"hostname,omitempty"`
	Memory     int             `json:"memory,omitempty"` // MB
	RootFS     string          `json:"rootfs,omitempty"` // e.g. "local-lvm:8"
	Password   SensitiveString `json:"password,omitempty"`
	Net0       string          `json:"net0,omitempty"`  // e.g. "name=eth0,bridge=vmbr0,dhcp=1"
	Start      int             `json:"start,omitempty"` // 1 to start after creation
}

// CloneContainerRequest is the request body for POST /nodes/{node}/lxc/{vmid}/clone.
type CloneContainerRequest struct {
	NewID    int    `json:"newid"`
	Hostname string `json:"hostname,omitempty"`
	Target   string `json:"target,omitempty"` // target node; defaults to source node
}

// SetVMConfigRequest is the request body for PUT /nodes/{node}/qemu/{vmid}/config.
// All fields are optional — only non-zero/non-nil fields are sent, so callers only
// need to populate the fields they want to change.
type SetVMConfigRequest struct {
	Name        string `json:"name,omitempty"`
	Memory      int    `json:"memory,omitempty"` // MB
	Cores       int    `json:"cores,omitempty"`
	OnBoot      *int   `json:"onboot,omitempty"` // nil = omit; 0 = disabled; 1 = start at boot
	Description string `json:"description,omitempty"`
}

// SetContainerConfigRequest is the request body for PUT /nodes/{node}/lxc/{vmid}/config.
// All fields are optional — only non-zero/non-nil fields are sent.
type SetContainerConfigRequest struct {
	Hostname    string `json:"hostname,omitempty"`
	Memory      int    `json:"memory,omitempty"` // MB
	Swap        *int   `json:"swap,omitempty"`   // MB; nil = omit; 0 = disable swap
	OnBoot      *int   `json:"onboot,omitempty"` // nil = omit; 0 = disabled; 1 = start at boot
	Description string `json:"description,omitempty"`
}

// ResizeDiskRequest is the request body for PUT /nodes/{node}/qemu/{vmid}/resize
// and PUT /nodes/{node}/lxc/{vmid}/resize.
type ResizeDiskRequest struct {
	Disk string `json:"disk"` // e.g. "scsi0" for VMs, "rootfs" for containers
	Size string `json:"size"` // absolute (e.g. "50G") or relative increment (e.g. "+10G")
}

// MigrateVMRequest is the request body for POST /nodes/{node}/qemu/{vmid}/migrate.
type MigrateVMRequest struct {
	Target string `json:"target"`           // destination node name
	Online *int   `json:"online,omitempty"` // nil = omit; 1 = live migration (no guest downtime when supported)
}

// MigrateContainerRequest is the request body for POST /nodes/{node}/lxc/{vmid}/migrate.
type MigrateContainerRequest struct {
	Target  string `json:"target"`            // destination node name
	Restart *int   `json:"restart,omitempty"` // nil = omit; 1 = stop, migrate, restart on target
}

// CreateBackupRequest is the request body for POST /nodes/{node}/vzdump.
// All fields except VMID are optional — zero values use Proxmox defaults.
type CreateBackupRequest struct {
	VMID     int    `json:"vmid"`               // VM or container ID to back up
	Storage  string `json:"storage,omitempty"`  // target storage pool
	Mode     string `json:"mode,omitempty"`     // snapshot | suspend | stop (default: snapshot)
	Compress string `json:"compress,omitempty"` // 0 | gzip | lzo | zstd (default: zstd)
}

// RestoreVMRequest is the request body for POST /nodes/{node}/qemu when
// restoring a QEMU VM from a vzdump backup archive. The Archive field
// maps to Proxmox's "archive" parameter.
type RestoreVMRequest struct {
	VMID    int    `json:"vmid"`
	Archive string `json:"archive"`           // backup volid, e.g. "local:backup/vzdump-qemu-100-....vma.zst"
	Storage string `json:"storage,omitempty"` // target storage pool for restored disks
	Start   int    `json:"start,omitempty"`   // 1 to start VM immediately after restore
}

// RestoreContainerRequest is the request body for POST /nodes/{node}/lxc when
// restoring an LXC container from a vzdump backup archive. Proxmox reuses the
// "ostemplate" field for the backup volid when restore=1.
type RestoreContainerRequest struct {
	VMID     int    `json:"vmid"`
	Archive  string `json:"ostemplate"`         // backup volid — Proxmox uses "ostemplate" for this parameter
	Restore  int    `json:"restore"`            // must be 1 to signal restore (not fresh creation)
	Storage  string `json:"storage,omitempty"`  // target storage pool for rootfs
	Hostname string `json:"hostname,omitempty"` // override hostname after restore
	Start    int    `json:"start,omitempty"`    // 1 to start container immediately after restore
}

// StorageContent represents a single volume entry from
// GET /nodes/{node}/storage/{storage}/content.
type StorageContent struct {
	VolID     string `json:"volid"`
	Content   string `json:"content"`
	Format    string `json:"format,omitempty"`
	Size      int64  `json:"size,omitempty"`
	CTime     int64  `json:"ctime,omitempty"`
	VMID      int    `json:"vmid,omitempty"`
	Notes     string `json:"notes,omitempty"`
	Protected int    `json:"protected,omitempty"` // 1 if protected from deletion
}

// NodeCommandRequest is the request body for POST /nodes/{node}/status.
// Command must be one of "reboot" or "shutdown".
type NodeCommandRequest struct {
	Command string `json:"command"` // "reboot" or "shutdown"
}

// MoveVMDiskRequest is the request body for
// POST /nodes/{node}/qemu/{vmid}/move_disk.
// The disk is moved to a different storage pool. If DeleteSource is true the
// source volume is removed after a successful move.
type MoveVMDiskRequest struct {
	Disk         string `json:"disk"`             // e.g. "scsi0"
	Storage      string `json:"storage"`          // destination storage pool
	DeleteSource *int   `json:"delete,omitempty"` // nil = omit; 1 = delete source after move
}

// APIError represents an HTTP-level error returned by the Proxmox API.
type APIError struct {
	StatusCode int
	Body       string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return "proxmox API error " + itoa(e.StatusCode) + ": " + e.Body
}

// itoa is a minimal int-to-string helper to avoid importing strconv in types.go.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

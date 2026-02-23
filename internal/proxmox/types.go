// Package proxmox provides a client for the Proxmox VE REST API.
package proxmox

import (
	"encoding/json"
	"errors"
)

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

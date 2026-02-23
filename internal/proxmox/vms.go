package proxmox

import (
	"context"
	"fmt"
	"strconv"
)

// ListVMs returns all QEMU virtual machines on the specified node.
func (c *Client) ListVMs(ctx context.Context, node string) ([]VM, error) {
	var vms []VM
	if err := c.get(ctx, "/nodes/"+node+"/qemu", &vms); err != nil {
		return nil, fmt.Errorf("listing VMs on node %s: %w", node, err)
	}
	return vms, nil
}

// GetVMStatus returns detailed status and configuration for a specific VM.
// The returned map contains the full Proxmox API response for
// /nodes/{node}/qemu/{vmid}/status/current.
func (c *Client) GetVMStatus(ctx context.Context, node string, vmid int) (map[string]any, error) {
	var status map[string]any
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/status/current"
	if err := c.get(ctx, path, &status); err != nil {
		return nil, fmt.Errorf("getting status for VM %d on node %s: %w", vmid, node, err)
	}
	return status, nil
}

// StartVM starts a QEMU VM. It returns the UPID of the asynchronous task.
// The task completes asynchronously; use GetTaskStatus to poll for completion.
func (c *Client) StartVM(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/status/start"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("starting VM %d on node %s: %w", vmid, node, err)
	}
	return upid, nil
}

// StopVM performs a hard stop of a QEMU VM. It returns the UPID of the
// asynchronous task.
func (c *Client) StopVM(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/status/stop"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("stopping VM %d on node %s: %w", vmid, node, err)
	}
	return upid, nil
}

// ShutdownVM sends a graceful ACPI shutdown signal to a QEMU VM.
// It returns the UPID of the asynchronous task.
func (c *Client) ShutdownVM(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/status/shutdown"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("shutting down VM %d on node %s: %w", vmid, node, err)
	}
	return upid, nil
}

// RebootVM reboots a QEMU VM. It returns the UPID of the asynchronous task.
func (c *Client) RebootVM(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/status/reboot"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("rebooting VM %d on node %s: %w", vmid, node, err)
	}
	return upid, nil
}

// SuspendVM suspends a QEMU VM. It returns the UPID of the asynchronous task.
func (c *Client) SuspendVM(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/status/suspend"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("suspending VM %d on node %s: %w", vmid, node, err)
	}
	return upid, nil
}

// ResumeVM resumes a suspended QEMU VM. It returns the UPID of the asynchronous task.
func (c *Client) ResumeVM(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/status/resume"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("resuming VM %d on node %s: %w", vmid, node, err)
	}
	return upid, nil
}
